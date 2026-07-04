package api

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// UserResolver resolves the User GraphQL type.
type UserResolver struct {
	user *persistence.User
}

// Me resolves Query.me.
func (r *RootResolver) Me(ctx context.Context) (*UserResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &UserResolver{user: user}, nil
}

func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(r.user.ID)
}

func (r *UserResolver) DisplayName() string {
	return r.user.DisplayName
}
