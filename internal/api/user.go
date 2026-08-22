package api

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// UserResolver resolves the User GraphQL type.
type UserResolver struct {
	db   *persistence.DB
	user *persistence.User
}

// Me resolves Query.me.
func (r *RootResolver) Me(ctx context.Context) (*UserResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &UserResolver{db: r.db, user: user}, nil
}

// Users resolves Query.users — all users known to this instance.
func (r *RootResolver) Users(ctx context.Context) ([]*UserResolver, error) {
	users, err := r.db.ListAllUsers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*UserResolver, len(users))
	for i, u := range users {
		out[i] = &UserResolver{db: r.db, user: u}
	}
	return out, nil
}

func (r *UserResolver) ID() graphql.ID      { return graphql.ID(r.user.ID) }
func (r *UserResolver) DisplayName() string { return r.user.DisplayName }

// Persons resolves User.persons — all persons this user has access to.
func (r *UserResolver) Persons(ctx context.Context) ([]*PersonResolver, error) {
	persons, err := r.db.ListPersonsForUser(ctx, r.user.ID)
	if err != nil {
		return nil, err
	}
	out := make([]*PersonResolver, len(persons))
	for i, p := range persons {
		out[i] = &PersonResolver{db: r.db, person: p}
	}
	return out, nil
}

// ActivePerson resolves User.activePerson from the session-bound context value.
func (r *UserResolver) ActivePerson(ctx context.Context) (*PersonResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, nil
	}
	return &PersonResolver{db: r.db, person: person}, nil
}
