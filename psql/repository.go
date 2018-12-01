package psql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/tanordheim/babyname-tinder"
)

// Repository encapsulates all PostgreSQL storage mechanics.
type Repository struct {
	db *sqlx.DB
}

var _ babynames.Repository = &Repository{}

func (r *Repository) withTX(ctx context.Context, f func(*sqlx.Tx) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "Unable to start PostgreSQL transaction")
	}
	defer tx.Rollback()

	if err := f(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "Unable to commit PostgreSQL transaction")
	}
	return nil
}

// ImportNames imports a set of names to the database if they don't exist already.
func (r *Repository) ImportNames(ctx context.Context, names []string) error {
	return r.withTX(ctx, func(tx *sqlx.Tx) error {

		stmt, err := tx.PrepareContext(ctx, `
			INSERT INTO names (
				id,
				name
			) VALUES (
				$1,
				$2
			) ON CONFLICT (id) DO NOTHING
		`)
		if err != nil {
			return errors.Wrap(err, "Unable to prepare insert statement")
		}

		for i := 0; i < len(names); i++ {
			_, err := stmt.ExecContext(ctx, getIDForName(names[i]), names[i])
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("Unable to insert name %s", names[i]))
			}
		}

		return nil
	})
}

func (r *Repository) removeLikeFor(ctx context.Context, tx *sqlx.Tx, role babynames.Role, name string) error {
	_, err := tx.ExecContext(
		ctx,
		`
			DELETE FROM
				likes
			WHERE
				name_id = $1 AND
				role_id = $2
		`,
		getIDForName(name),
		role,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to remove any existing likes on name %s for role %v", name, role))
	}
	return nil
}

func (r *Repository) removeDislikeFor(ctx context.Context, tx *sqlx.Tx, role babynames.Role, name string) error {
	_, err := tx.ExecContext(
		ctx,
		`
			DELETE FROM
				dislikes
			WHERE
				name_id = $1 AND
				role_id = $2
		`,
		getIDForName(name),
		role,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to remove any existing dislikes on name %s for role %v", name, role))
	}
	return nil
}

// Like flags a name as liked for the specified role.
func (r *Repository) Like(ctx context.Context, role babynames.Role, name string) error {
	return r.withTX(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`
				INSERT INTO likes (
					role_id,
					name_id,
					liked_at
				) VALUES (
					$1,
					$2,
					CURRENT_TIMESTAMP
				) ON CONFLICT (role_id, name_id) DO NOTHING
			`,
			role,
			getIDForName(name),
		)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to like name '%s' as role '%v'", name, role))
		}
		if err := r.removeDislikeFor(ctx, tx, role, name); err != nil {
			return err
		}
		return nil
	})
}

// UndoLike removes a like for a name.
func (r *Repository) UndoLike(ctx context.Context, role babynames.Role, name string) error {
	return r.withTX(ctx, func(tx *sqlx.Tx) error {
		return r.removeLikeFor(ctx, tx, role, name)
	})
}

// Superlike flags a name as super-liked for the specified role.
func (r *Repository) Superlike(ctx context.Context, role babynames.Role, name string) error {
	return r.withTX(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`
				INSERT INTO likes (
					role_id,
					name_id,
					liked_at,
					superlike
				) VALUES (
					$1,
					$2,
					CURRENT_TIMESTAMP,
					't'
				) ON CONFLICT (role_id, name_id) DO NOTHING
			`,
			role,
			getIDForName(name),
		)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to superlike name '%s' as role '%v'", name, role))
		}
		if err := r.removeDislikeFor(ctx, tx, role, name); err != nil {
			return err
		}

		// Delete any potential dislikes on this name from the other role
		if err := r.removeDislikeFor(ctx, tx, babynames.InverseRole(role), name); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to delete any dislikes due to superlike of name '%s' by role '%v'", name, role))
		}

		return nil
	})
}

// Dislike flags a name as disliked for the specified role, returning the number of times the role has disliked the name.
func (r *Repository) Dislike(ctx context.Context, role babynames.Role, name string) (int, error) {
	var dislikeCount int

	err := r.withTX(ctx, func(tx *sqlx.Tx) error {
		row := tx.QueryRowxContext(
			ctx,
			`
				INSERT INTO dislikes (
					role_id,
					name_id,
					disliked_first_at,
					disliked_last_at,
					disliked_times
				) VALUES (
					$1,
					$2,
					CURRENT_TIMESTAMP,
					CURRENT_TIMESTAMP,
					1
				) ON CONFLICT (role_id, name_id) DO UPDATE SET
					disliked_last_at = CURRENT_TIMESTAMP,
					disliked_times = dislikes.disliked_times + 1
				RETURNING disliked_times
			`,
			role,
			getIDForName(name),
		)
		if err := row.Scan(&dislikeCount); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to dislike name '%s' as role '%v'", name, role))
		}

		if err := r.removeLikeFor(ctx, tx, role, name); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return dislikeCount, nil
}

// UndoDislike removes a dislike for a name.
func (r *Repository) UndoDislike(ctx context.Context, role babynames.Role, name string) error {
	return r.withTX(ctx, func(tx *sqlx.Tx) error {
		return r.removeDislikeFor(ctx, tx, role, name)
	})
}

// GetPendingSuperlike gets any pending superlikes that requires the role's attention.
func (r *Repository) GetPendingSuperlike(ctx context.Context, role babynames.Role) (string, error) {
	otherRole := babynames.InverseRole(role)

	var name string
	row := r.db.QueryRowxContext(
		ctx,
		`
			SELECT
				name
			FROM
				names
			INNER JOIN likes ON likes.role_id = $1 AND likes.superlike = 't' AND likes.name_id = names.id
			LEFT JOIN dislikes ON dislikes.role_id = $2 AND dislikes.name_id = names.id
			LEFT JOIN likes AS other_likes ON other_likes.role_id = $2 AND other_likes.name_id = names.id
			WHERE
				dislikes.name_id IS NULL AND
				other_likes.name_id IS NULL
			LIMIT 1
		`,
		otherRole,
		role,
	)

	if err := row.Scan(&name); err != nil && err != sql.ErrNoRows {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to retrieve pending superlike for role '%v'", role))
	}
	return name, nil
}

func (r *Repository) acknowledgeMatch(ctx context.Context, role babynames.Role, name string) error {
	return r.withTX(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(
			ctx,
			`
				INSERT INTO acknowledged_matches (
					role_id,
					name_id,
					acknowledged_at
				) VALUES (
					$1,
					$2,
					CURRENT_TIMESTAMP
				)
			`,
			role,
			getIDForName(name),
		)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Unable to acknowledge match on name '%s' for role '%v'", name, role))
		}
		return nil
	})
}

// GetAndAcknowledgeUnseenMatch returns a matched name between both roles that the specified role has not yet seen. This function will flag the name as seen in the process.
func (r *Repository) GetAndAcknowledgeUnseenMatch(ctx context.Context, role babynames.Role) (string, error) {
	var name string
	row := r.db.QueryRowxContext(
		ctx,
		`
			SELECT
				name
			FROM
				names
			INNER JOIN likes AS role_likes ON role_likes.role_id = $1 AND role_likes.name_id = names.id
			INNER JOIN likes AS inverse_likes ON inverse_likes.role_id = $2 AND inverse_likes.name_id = names.id
			LEFT JOIN acknowledged_matches ON acknowledged_matches.role_id = $1 AND acknowledged_matches.name_id = names.id
			WHERE acknowledged_matches.name_id IS NULL
			ORDER BY name
			LIMIT 1
		`,
		role,
		babynames.InverseRole(role),
	)
	if err := row.Scan(&name); err != nil && err != sql.ErrNoRows {
		return "", errors.Wrap(err, fmt.Sprintf("Unable to retrieve unseen match for role '%v'", role))
	}

	if name != "" {
		if err := r.acknowledgeMatch(ctx, role, name); err != nil {
			return "", err
		}
	}
	return name, nil
}

// GetNextName gets the next name in the queue for the role.
func (r *Repository) GetNextName(ctx context.Context, role babynames.Role) (string, int, error) {
	var name string
	var dislikes int
	row := r.db.QueryRowxContext(
		ctx,
		`
			SELECT
				names.name,
				COALESCE(dislikes.disliked_times, 0) as disliked_times
			FROM
				names
			LEFT JOIN likes ON likes.role_id = $1 AND likes.name_id = names.id
			LEFT JOIN dislikes ON dislikes.role_id = $1 AND dislikes.name_id = names.id
			WHERE
				likes.name_id IS NULL AND
				(dislikes.name_id IS NULL OR dislikes.disliked_times < $2)
			ORDER BY random()
			LIMIT 1
		`,
		role,
		babynames.DislikesBeforeRemoved,
	)
	if err := row.Scan(&name, &dislikes); err != nil && err != sql.ErrNoRows {
		return "", 0, errors.Wrap(err, fmt.Sprintf("Unable to retrieve next name for role '%v'", role))
	}

	return name, dislikes, nil
}

// GetLikedNames gets a list of all liked names by the role.
func (r *Repository) GetLikedNames(ctx context.Context, role babynames.Role) ([]babynames.LikedName, error) {
	rows, err := r.db.QueryxContext(
		ctx,
		`
			SELECT
				names.name,
				likes.superlike,
				likes.liked_at
			FROM
				names
			INNER JOIN likes ON likes.role_id = $1 AND likes.name_id = names.id
			ORDER BY names.name
		`,
		role,
	)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to retrieve liked names for role '%v'", role))
	}
	defer rows.Close()

	res := []babynames.LikedName{}
	for rows.Next() {
		var (
			name      string
			superlike bool
			likedAt   time.Time
		)
		if err := rows.Scan(&name, &superlike, &likedAt); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to read liked name for role '%v'", role))
		}

		res = append(res, babynames.LikedName{
			Name:       name,
			Superliked: superlike,
			LikedAt:    likedAt,
		})
	}

	return res, nil
}

// GetDislikedNames gets a list of all disliked names by the role.
func (r *Repository) GetDislikedNames(ctx context.Context, role babynames.Role) ([]babynames.DislikedName, error) {
	rows, err := r.db.QueryxContext(
		ctx,
		`
			SELECT
				names.name,
				dislikes.disliked_times,
				dislikes.disliked_first_at,
				dislikes.disliked_last_at
			FROM
				names
			INNER JOIN dislikes ON dislikes.role_id = $1 AND dislikes.name_id = names.id
			ORDER BY names.name
		`,
		role,
	)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to retrieve disliked names for role '%v'", role))
	}
	defer rows.Close()

	res := []babynames.DislikedName{}
	for rows.Next() {
		var (
			name    string
			count   int
			firstAt time.Time
			lastAt  time.Time
		)
		if err := rows.Scan(&name, &count, &firstAt, &lastAt); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to read disliked name for role '%v'", role))
		}

		res = append(res, babynames.DislikedName{
			Name:         name,
			Count:        count,
			FirstDislike: firstAt,
			LastDislike:  lastAt,
		})
	}

	return res, nil
}

// GetMatches gets a list of all names that are matched with the other role.
func (r *Repository) GetMatches(ctx context.Context, role babynames.Role) ([]babynames.Match, error) {
	rows, err := r.db.QueryxContext(
		ctx,
		`
			SELECT
				names.name,
				role_likes.liked_at AS role_liked_at,
				role_likes.superlike AS role_superlike,
				other_role_likes.liked_at AS other_role_liked_at,
				other_role_likes.superlike AS other_role_superlike
			FROM
				names
			INNER JOIN likes AS role_likes ON role_likes.role_id = $1 AND role_likes.name_id = names.id
			INNER JOIN likes AS other_role_likes ON other_role_likes.role_id = $2 AND other_role_likes.name_id = names.id
			ORDER BY names.name
		`,
		role,
		babynames.InverseRole(role),
	)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Unable to retrieve matched names for role '%v'", role))
	}
	defer rows.Close()

	res := []babynames.Match{}
	for rows.Next() {
		var (
			name            string
			roleLikedAt     time.Time
			roleSuperliked  bool
			otherLikedAt    time.Time
			otherSuperliked bool
		)
		if err := rows.Scan(&name, &roleLikedAt, &roleSuperliked, &otherLikedAt, &otherSuperliked); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unable to read matched name for role '%v'", role))
		}

		res = append(res, babynames.Match{
			Name: name,
			Roles: map[babynames.Role]babynames.MatchRole{
				role: babynames.MatchRole{
					LikedAt:    roleLikedAt,
					Superliked: roleSuperliked,
				},
				babynames.InverseRole(role): babynames.MatchRole{
					LikedAt:    otherLikedAt,
					Superliked: otherSuperliked,
				},
			},
		})
	}

	return res, nil
}

// GetStats retrieves the progression stats of a role.
func (r *Repository) GetStats(ctx context.Context, role babynames.Role) (babynames.Stats, error) {
	// Get total number of names
	var total int
	err := r.db.QueryRowxContext(ctx, "SELECT COUNT(1) FROM names").Scan(&total)
	if err != nil {
		return babynames.Stats{}, errors.Wrap(err, "Unable to count all names")
	}

	// Get number of liked names
	var liked int
	err = r.db.QueryRowxContext(ctx, "SELECT COUNT(1) FROM likes WHERE role_id = $1", role).Scan(&liked)
	if err != nil {
		return babynames.Stats{}, errors.Wrap(err, fmt.Sprintf("Unable to count liked names for role '%v'", role))
	}

	// Get number of disliked names
	var disliked int
	err = r.db.QueryRowxContext(ctx, "SELECT COUNT(1) FROM dislikes WHERE role_id = $1", role).Scan(&disliked)
	if err != nil {
		return babynames.Stats{}, errors.Wrap(err, fmt.Sprintf("Unable to count disliked names for role '%v'", role))
	}

	// Get number of queued names (names that has not been liked, and disliked less than the required number of times for exclusion)
	var queued int
	err = r.db.QueryRowxContext(
		ctx,
		`
			SELECT
				COUNT(1)
			FROM
				names
			LEFT JOIN likes ON likes.role_id = $1 AND likes.name_id = names.id
			LEFT JOIN dislikes ON dislikes.role_id = $1 AND dislikes.name_id = names.id AND dislikes.disliked_times >= $2 
			WHERE
				likes.name_id IS NULL AND
				dislikes.name_id IS NULL
		`,
		role,
		babynames.DislikesBeforeRemoved,
	).Scan(&queued)
	if err != nil {
		return babynames.Stats{}, errors.Wrap(err, fmt.Sprintf("Unable to count queued names for role '%v'", role))
	}

	// Get number of matched names
	var matched int
	err = r.db.QueryRowxContext(
		ctx,
		`
			SELECT
				COUNT(1)
			FROM
				names
			INNER JOIN likes ON likes.role_id = $1 AND likes.name_id = names.id
			INNER JOIN likes AS other_likes ON other_likes.role_id = $2 AND other_likes.name_id = names.id
		`,
		role,
		babynames.InverseRole(role),
	).Scan(&matched)
	if err != nil {
		return babynames.Stats{}, errors.Wrap(err, fmt.Sprintf("Unable to count matched names for role '%v'", role))
	}

	return babynames.Stats{
		Total:    total,
		Liked:    liked,
		Disliked: disliked,
		Queued:   queued,
		Matched:  matched,
	}, nil
}
