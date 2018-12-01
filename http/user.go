package http

import (
	"context"

	"github.com/tanordheim/babyname-tinder"
)

type contextType int

const (
	currentUserContextKey contextType = 0
)

type user struct {
	EmailAddress string
	Role         babynames.Role
}

func newUser(email string, role babynames.Role) *user {
	return &user{
		EmailAddress: email,
		Role:         role,
	}
}

func setCurrentUser(ctx context.Context, user *user) context.Context {
	return context.WithValue(ctx, currentUserContextKey, user)
}

func getCurrentUser(ctx context.Context) *user {
	return ctx.Value(currentUserContextKey).(*user)
}
