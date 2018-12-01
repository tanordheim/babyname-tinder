package babynames

import (
	"context"
	"time"
)

// DislikesBeforeRemoved sets number of times a name should have to be disliked before no longer showing up
const DislikesBeforeRemoved = 2

// LikedName describes a name that has been liked.
type LikedName struct {
	Name       string
	Superliked bool
	LikedAt    time.Time
}

// DislikedName describes a name that has been disliked.
type DislikedName struct {
	Name         string
	Count        int
	FirstDislike time.Time
	LastDislike  time.Time
}

// Match describes a name that has been liked by both roles.
type Match struct {
	Name  string
	Roles map[Role]MatchRole
}

// MatchRole describes when and how a role participated in a match.
type MatchRole struct {
	LikedAt    time.Time
	Superliked bool
}

// Stats represents the progress of a role.
type Stats struct {
	Total    int
	Liked    int
	Disliked int
	Queued   int
	Matched  int
}

// Repository defines the data access layer behavior.
type Repository interface {
	ImportNames(context.Context, []string) error
	Like(context.Context, Role, string) error
	Superlike(context.Context, Role, string) error
	UndoLike(context.Context, Role, string) error
	Dislike(context.Context, Role, string) (int, error)
	UndoDislike(context.Context, Role, string) error
	GetPendingSuperlike(context.Context, Role) (string, error)
	GetAndAcknowledgeUnseenMatch(context.Context, Role) (string, error)
	GetNextName(context.Context, Role) (string, int, error)
	GetLikedNames(context.Context, Role) ([]LikedName, error)
	GetDislikedNames(context.Context, Role) ([]DislikedName, error)
	GetMatches(context.Context, Role) ([]Match, error)
	GetStats(context.Context, Role) (Stats, error)
}
