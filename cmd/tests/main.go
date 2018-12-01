package main

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/tanordheim/babyname-tinder"
	"github.com/tanordheim/babyname-tinder/psql"
)

type liked struct {
	name      string
	superlike bool
}
type disliked struct {
	name  string
	count int
}

func main() {
	ctx := context.Background()
	repo := psql.NewRepository(os.Getenv("DATABASE_URL"))

	// Create some test names
	names := make([]string, 10)
	for i := 0; i < len(names); i++ {
		names[i] = fmt.Sprintf("Test Name %d", i)
	}
	if err := repo.ImportNames(ctx, names); err != nil {
		panic(errors.Wrap(err, "Unable to import fake names"))
	}

	assertLike := func(role babynames.Role, name string) {
		if err := repo.Like(ctx, role, name); err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to like name '%s' as role '%v'", name, role)))
		}
	}
	assertDislike := func(role babynames.Role, name string, count int) {
		actual, err := repo.Dislike(ctx, role, name)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to dislike name '%s' as role '%v'", name, role)))
		}
		if actual != count {
			panic(fmt.Errorf("Expected disliking name '%s' as role '%v' to have count of %d, got %d", name, role, count, actual))
		}
	}
	assertSuperlike := func(role babynames.Role, name string) {
		if err := repo.Superlike(ctx, role, name); err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to superlike name '%s' as role '%v'", name, role)))
		}
	}
	assertPendingSuperlike := func(role babynames.Role, name string) {
		actual, err := repo.GetPendingSuperlike(ctx, role)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to get pending superlike for role '%v'", role)))
		}
		if actual != name {
			panic(fmt.Errorf("Expected pending superlike for role '%v' to be '%s', got '%s'", role, name, actual))
		}
	}
	assertUnseenMatch := func(role babynames.Role, name string) {
		actual, err := repo.GetAndAcknowledgeUnseenMatch(ctx, role)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to get unseen matches for role '%v'", role)))
		}
		if actual != name {
			panic(fmt.Errorf("Expected unseen match '%s' for role '%v', got '%s'", name, role, actual))
		}
	}
	assertNextNames := func(role babynames.Role, names ...string) {
		seenNames := map[string]bool{}

		for i := 0; i < 100; i++ {
			name, _, err := repo.GetNextName(ctx, role)
			if err != nil {
				panic(errors.Wrap(err, fmt.Sprintf("Unable to get next name for role '%v'", role)))
			}
			seenNames[name] = true
		}

		allNames := []string{}
		for k := range seenNames {
			allNames = append(allNames, k)
		}

		// Make sure we got all the names we wanted
		for _, name := range names {
			if _, ok := seenNames[name]; !ok {
				panic(fmt.Errorf("Expected name '%s' to be in queue for role '%v', but name was not found: got %+v", name, role, allNames))
			}
		}

		// Make sure we didn't get any names we didn't want
		for _, gotName := range allNames {
			found := false
			for _, wantName := range names {
				if wantName == gotName {
					found = true
					break
				}
			}

			if !found {
				panic(fmt.Errorf("Found unexpected name '%s' in queue for role '%v': only expected %+v", gotName, role, names))
			}
		}
	}
	assertNoNextName := func(role babynames.Role) {
		name, _, err := repo.GetNextName(ctx, role)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to get next name for role '%v'", role)))
		}
		if name != "" {
			panic(fmt.Errorf("Expected no next name for role '%v', got '%s'", role, name))
		}
	}
	assertLiked := func(role babynames.Role, names ...liked) {
		liked, err := repo.GetLikedNames(ctx, role)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to get liked names for role '%v'", role)))
		}
		if len(liked) != len(names) {
			panic(fmt.Errorf("Expected %d liked names for role '%v' (%+v), got %d names (%+v)", len(names), role, names, len(liked), liked))
		}

		// Assert all the names we wanted are there
		for _, name := range names {
			found := false
			for _, like := range liked {
				if like.Name == name.name {
					if like.Superliked != name.superlike {
						panic(fmt.Errorf("Expected liked name '%s' to have superlike=%v for role '%v', got superlike=%v", name.name, name.superlike, role, like.Superliked))
					}
					found = true
					break
				}
			}

			if !found {
				panic(fmt.Errorf("Expected liked names for role '%v' to contain '%s': got %+v", role, name.name, liked))
			}
		}
	}
	assertDisliked := func(role babynames.Role, names ...disliked) {
		disliked, err := repo.GetDislikedNames(ctx, role)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to get disliked names for role '%v'", role)))
		}
		if len(disliked) != len(names) {
			panic(fmt.Errorf("Expected %d disliked names for role '%v' (%+v), got %d names (%+v)", len(names), role, names, len(disliked), disliked))
		}

		// Assert all the names we wanted are there
		for _, name := range names {
			found := false
			for _, dislike := range disliked {
				if dislike.Name == name.name {
					if dislike.Count != name.count {
						panic(fmt.Errorf("Expected disliked name '%s' to have count=%d for role '%v', got count=%d", name.name, name.count, role, dislike.Count))
					}
					found = true
					break
				}
			}

			if !found {
				panic(fmt.Errorf("Expected disliked names for role '%v' to contain '%s': got %+v", role, name.name, disliked))
			}
		}
	}
	assertMatches := func(role babynames.Role, names ...string) {
		matches, err := repo.GetMatches(ctx, role)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to get matched names for role '%v'", role)))
		}
		if len(matches) != len(names) {
			panic(fmt.Errorf("Expected %d matched names for role '%v' (%+v), got %d names (%+v)", len(names), role, names, len(matches), matches))
		}

		// Assert all the names we wanted are there
		for _, name := range names {
			found := false
			for _, match := range matches {
				if match.Name == name {
					found = true
				}
			}

			if !found {
				panic(fmt.Errorf("Expected matched name for role '%v' to contain '%s': got %+v", role, name, matches))
			}
		}
	}
	assertStats := func(role babynames.Role, liked, disliked, queued, matched int) {
		res, err := repo.GetStats(ctx, role)
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("Unable to get stats for role '%v'", role)))
		}
		if res.Total != 10 {
			panic(fmt.Errorf("Expected 10 names total, got %d", res.Total))
		}
		if res.Liked != liked {
			panic(fmt.Errorf("Expected %d liked names, got %d", liked, res.Liked))
		}
		if res.Disliked != disliked {
			panic(fmt.Errorf("Expected %d disliked names, got %d", disliked, res.Disliked))
		}
		if res.Queued != queued {
			panic(fmt.Errorf("Expected %d queued names, got %d", queued, res.Queued))
		}
		if res.Matched != matched {
			panic(fmt.Errorf("Expected %d matched names, got %d", matched, res.Matched))
		}
	}

	// Check stats with no data
	assertStats(babynames.DadRole, 0, 0, 10, 0)
	assertStats(babynames.MomRole, 0, 0, 10, 0)

	// Like some of the names as mom and dad
	assertLike(babynames.DadRole, "Test Name 1")
	assertLike(babynames.DadRole, "Test Name 2")
	assertLike(babynames.DadRole, "Test Name 2") // duplicate like
	assertLike(babynames.MomRole, "Test Name 3")
	assertLike(babynames.MomRole, "Test Name 4")

	// Check stats with some likes
	assertStats(babynames.DadRole, 2, 0, 8, 0)
	assertStats(babynames.MomRole, 2, 0, 8, 0)

	// Dislike some of the names as mom and dad
	assertDislike(babynames.DadRole, "Test Name 3", 1)
	assertDislike(babynames.DadRole, "Test Name 4", 1)
	assertDislike(babynames.DadRole, "Test Name 4", 2) // duplicate dislike
	assertDislike(babynames.MomRole, "Test Name 5", 1)
	assertDislike(babynames.MomRole, "Test Name 2", 1)
	assertDislike(babynames.MomRole, "Test Name 2", 2) // duplicate dislike

	// Check stats with some dislikes
	assertStats(babynames.DadRole, 2, 2, 7, 0)
	assertStats(babynames.MomRole, 2, 2, 7, 0)

	// Superlike some of the names as mom and dad
	assertSuperlike(babynames.DadRole, "Test Name 5")
	assertSuperlike(babynames.MomRole, "Test Name 6")

	// Try to get a pending superlike
	assertPendingSuperlike(babynames.DadRole, "Test Name 6")
	assertPendingSuperlike(babynames.MomRole, "Test Name 5") // even though it's disliked

	// Dislike a superliked name for the mom and assert it no longer shows up
	assertDislike(babynames.MomRole, "Test Name 5", 1)
	assertPendingSuperlike(babynames.MomRole, "")

	// Like a superliked name for the dad and assert it no longer shows up
	assertLike(babynames.DadRole, "Test Name 6")
	assertPendingSuperlike(babynames.DadRole, "")

	// Create a few matches and assert we can get them out
	assertLike(babynames.DadRole, "Test Name 7")
	assertLike(babynames.DadRole, "Test Name 8")
	assertLike(babynames.MomRole, "Test Name 7")
	assertLike(babynames.MomRole, "Test Name 8")
	assertUnseenMatch(babynames.DadRole, "Test Name 6")
	assertUnseenMatch(babynames.DadRole, "Test Name 7")
	assertUnseenMatch(babynames.DadRole, "Test Name 8")
	assertUnseenMatch(babynames.DadRole, "")
	assertUnseenMatch(babynames.MomRole, "Test Name 6")
	assertUnseenMatch(babynames.MomRole, "Test Name 7")
	assertUnseenMatch(babynames.MomRole, "Test Name 8")
	assertUnseenMatch(babynames.MomRole, "")

	// Get some pending names and assert we get everything for both roles
	assertNextNames(babynames.DadRole, "Test Name 0", "Test Name 3", "Test Name 9")
	assertNextNames(babynames.MomRole, "Test Name 0", "Test Name 1", "Test Name 5", "Test Name 9")

	// Like/dislike the remaining names
	assertLike(babynames.DadRole, "Test Name 0")
	assertLike(babynames.DadRole, "Test Name 3")
	assertDislike(babynames.DadRole, "Test Name 9", 1)
	assertDislike(babynames.DadRole, "Test Name 9", 2)
	assertNoNextName(babynames.DadRole)
	assertLike(babynames.MomRole, "Test Name 0")
	assertLike(babynames.MomRole, "Test Name 1")
	assertDislike(babynames.MomRole, "Test Name 5", 2)
	assertDislike(babynames.MomRole, "Test Name 9", 1)
	assertLike(babynames.MomRole, "Test Name 9") // should remove the previous dislike
	assertNoNextName(babynames.MomRole)

	// List out liked names
	assertLiked(
		babynames.DadRole,
		liked{"Test Name 0", false},
		liked{"Test Name 1", false},
		liked{"Test Name 2", false},
		liked{"Test Name 3", false},
		liked{"Test Name 5", true},
		liked{"Test Name 6", false},
		liked{"Test Name 7", false},
		liked{"Test Name 8", false},
	)
	assertLiked(
		babynames.MomRole,
		liked{"Test Name 0", false},
		liked{"Test Name 1", false},
		liked{"Test Name 3", false},
		liked{"Test Name 4", false},
		liked{"Test Name 6", true},
		liked{"Test Name 7", false},
		liked{"Test Name 8", false},
		liked{"Test Name 9", false},
	)

	// List out disliked names
	assertDisliked(
		babynames.DadRole,
		disliked{"Test Name 4", 2},
		disliked{"Test Name 9", 2},
	)
	assertDisliked(
		babynames.MomRole,
		disliked{"Test Name 2", 2},
		disliked{"Test Name 5", 2},
	)

	// List out matches
	for _, role := range []babynames.Role{babynames.DadRole, babynames.MomRole} {
		assertMatches(
			role,
			"Test Name 0",
			"Test Name 1",
			"Test Name 3",
			"Test Name 6",
			"Test Name 7",
			"Test Name 8",
		)
	}

	// Check stats after everything is processed
	assertStats(babynames.DadRole, 8, 2, 0, 6)
	assertStats(babynames.MomRole, 8, 2, 0, 6)
}
