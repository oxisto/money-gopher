package api

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// PersonResolver resolves the Person GraphQL type.
type PersonResolver struct {
	db     *persistence.DB
	person *persistence.Person
}

func (r *PersonResolver) ID() graphql.ID      { return graphql.ID(r.person.ID) }
func (r *PersonResolver) DisplayName() string { return r.person.DisplayName }

// CreatePerson resolves Mutation.createPerson.
func (r *RootResolver) CreatePerson(ctx context.Context, args struct{ DisplayName string }) (*PersonResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	person, err := r.db.CreatePerson(ctx, persistence.CreatePersonParams{
		ID:          uuid.NewString(),
		DisplayName: args.DisplayName,
	})
	if err != nil {
		return nil, err
	}

	if err := r.db.GrantPersonAccess(ctx, persistence.GrantPersonAccessParams{
		UserID:   user.ID,
		PersonID: person.ID,
	}); err != nil {
		return nil, err
	}

	return &PersonResolver{db: r.db, person: person}, nil
}

// SwitchPerson resolves Mutation.switchPerson. It updates the active person
// stored on the current session, returning the new active person.
func (r *RootResolver) SwitchPerson(ctx context.Context, args struct{ PersonID graphql.ID }) (*PersonResolver, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	sessionID, ok := auth.SessionIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no session in context")
	}

	// Verify the user has access to the requested person.
	count, err := r.db.CheckPersonAccess(ctx, persistence.CheckPersonAccessParams{
		UserID:   user.ID,
		PersonID: string(args.PersonID),
	})
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, fmt.Errorf("person %q not found or not accessible", args.PersonID)
	}

	person, err := r.db.GetPerson(ctx, string(args.PersonID))
	if err != nil {
		return nil, err
	}

	if err := r.db.UpdateSessionPerson(ctx, persistence.UpdateSessionPersonParams{
		PersonID: nullString(&person.ID),
		ID:       sessionID,
	}); err != nil {
		return nil, err
	}

	return &PersonResolver{db: r.db, person: person}, nil
}
