package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// SecurityResolver resolves the Security GraphQL type. Securities are global,
// so unlike portfolios and cash accounts they are not scoped to a user.
type SecurityResolver struct {
	db       *persistence.DB
	security *persistence.Security
}

// SecurityIdentifierResolver resolves the SecurityIdentifier GraphQL type.
type SecurityIdentifierResolver struct {
	identifier *persistence.SecurityIdentifier
}

// ListingResolver resolves the Listing GraphQL type.
type ListingResolver struct {
	db      *persistence.DB
	listing *persistence.Listing
}

// QuoteResolver resolves the Quote GraphQL type.
type QuoteResolver struct {
	quote    *persistence.Quote
	currency string
}

// Securities resolves Query.securities.
func (r *RootResolver) Securities(ctx context.Context) ([]*SecurityResolver, error) {
	securities, err := r.db.ListSecurities(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*SecurityResolver, 0, len(securities))
	for _, s := range securities {
		resolvers = append(resolvers, &SecurityResolver{db: r.db, security: s})
	}

	return resolvers, nil
}

// Security resolves Query.security.
func (r *RootResolver) Security(ctx context.Context, args struct{ ID graphql.ID }) (*SecurityResolver, error) {
	security, err := r.db.GetSecurity(ctx, string(args.ID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &SecurityResolver{db: r.db, security: security}, nil
}

// SecurityIdentifierInput is one external identifier in security mutations.
type SecurityIdentifierInput struct {
	Kind  string
	Value string
}

// ListingInput is one exchange listing in security mutations.
type ListingInput struct {
	Exchange      *string
	Ticker        string
	Currency      string
	QuoteProvider *string
}

// CreateSecurityInput is the input for Mutation.createSecurity.
type CreateSecurityInput struct {
	DisplayName string
	Identifiers *[]SecurityIdentifierInput
	Listings    *[]ListingInput
}

// CreateSecurity resolves Mutation.createSecurity.
func (r *RootResolver) CreateSecurity(ctx context.Context, args struct{ Input CreateSecurityInput }) (*SecurityResolver, error) {
	var security *persistence.Security

	err := r.db.Tx(ctx, func(q *persistence.Queries) (err error) {
		security, err = q.CreateSecurity(ctx, persistence.CreateSecurityParams{
			ID:          uuid.NewString(),
			DisplayName: args.Input.DisplayName,
		})
		if err != nil {
			return err
		}

		if args.Input.Identifiers != nil {
			if err = createIdentifiers(ctx, q, security.ID, *args.Input.Identifiers); err != nil {
				return err
			}
		}
		if args.Input.Listings != nil {
			if err = createListings(ctx, q, security.ID, *args.Input.Listings); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &SecurityResolver{db: r.db, security: security}, nil
}

// UpdateSecurityInput is the input for Mutation.updateSecurity. Provided
// identifiers or listings replace the existing set entirely.
type UpdateSecurityInput struct {
	DisplayName *string
	Identifiers *[]SecurityIdentifierInput
	Listings    *[]ListingInput
}

// UpdateSecurity resolves Mutation.updateSecurity.
func (r *RootResolver) UpdateSecurity(ctx context.Context, args struct {
	ID    graphql.ID
	Input UpdateSecurityInput
}) (*SecurityResolver, error) {
	var security *persistence.Security

	err := r.db.Tx(ctx, func(q *persistence.Queries) (err error) {
		security, err = q.GetSecurity(ctx, string(args.ID))
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("security %q not found", args.ID)
		} else if err != nil {
			return err
		}

		if args.Input.DisplayName != nil {
			security, err = q.UpdateSecurity(ctx, persistence.UpdateSecurityParams{
				DisplayName: *args.Input.DisplayName,
				ID:          string(args.ID),
			})
			if err != nil {
				return err
			}
		}

		if args.Input.Identifiers != nil {
			if err = q.DeleteSecurityIdentifiers(ctx, security.ID); err != nil {
				return err
			}
			if err = createIdentifiers(ctx, q, security.ID, *args.Input.Identifiers); err != nil {
				return err
			}
		}

		// Note: replacing listings drops their quote history via ON DELETE
		// CASCADE. Fine for now, revisit once quotes arrive in M4.
		if args.Input.Listings != nil {
			if err = q.DeleteListings(ctx, security.ID); err != nil {
				return err
			}
			if err = createListings(ctx, q, security.ID, *args.Input.Listings); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &SecurityResolver{db: r.db, security: security}, nil
}

// DeleteSecurity resolves Mutation.deleteSecurity. Deleting fails while
// transactions still reference the security.
func (r *RootResolver) DeleteSecurity(ctx context.Context, args struct{ ID graphql.ID }) (graphql.ID, error) {
	rows, err := r.db.DeleteSecurity(ctx, string(args.ID))
	if err != nil {
		return "", err
	}
	if rows == 0 {
		return "", fmt.Errorf("security %q not found", args.ID)
	}

	return args.ID, nil
}

func createIdentifiers(ctx context.Context, q *persistence.Queries, securityID string, identifiers []SecurityIdentifierInput) error {
	for _, i := range identifiers {
		err := q.CreateSecurityIdentifier(ctx, persistence.CreateSecurityIdentifierParams{
			SecurityID: securityID,
			Kind:       i.Kind,
			Value:      i.Value,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func createListings(ctx context.Context, q *persistence.Queries, securityID string, listings []ListingInput) error {
	for _, l := range listings {
		_, err := q.CreateListing(ctx, persistence.CreateListingParams{
			ID:            uuid.NewString(),
			SecurityID:    securityID,
			Exchange:      nullString(l.Exchange),
			Ticker:        l.Ticker,
			Currency:      l.Currency,
			QuoteProvider: nullString(l.QuoteProvider),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SecurityResolver) ID() graphql.ID {
	return graphql.ID(r.security.ID)
}

func (r *SecurityResolver) DisplayName() string {
	return r.security.DisplayName
}

// Identifiers resolves Security.identifiers.
func (r *SecurityResolver) Identifiers(ctx context.Context) ([]*SecurityIdentifierResolver, error) {
	identifiers, err := r.db.ListSecurityIdentifiers(ctx, r.security.ID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*SecurityIdentifierResolver, 0, len(identifiers))
	for _, i := range identifiers {
		resolvers = append(resolvers, &SecurityIdentifierResolver{identifier: i})
	}

	return resolvers, nil
}

// Listings resolves Security.listings.
func (r *SecurityResolver) Listings(ctx context.Context) ([]*ListingResolver, error) {
	listings, err := r.db.ListListings(ctx, r.security.ID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*ListingResolver, 0, len(listings))
	for _, l := range listings {
		resolvers = append(resolvers, &ListingResolver{db: r.db, listing: l})
	}

	return resolvers, nil
}

func (r *SecurityIdentifierResolver) Kind() string {
	return r.identifier.Kind
}

func (r *SecurityIdentifierResolver) Value() string {
	return r.identifier.Value
}

func (r *ListingResolver) ID() graphql.ID {
	return graphql.ID(r.listing.ID)
}

func (r *ListingResolver) Exchange() *string {
	return stringPtr(r.listing.Exchange)
}

func (r *ListingResolver) Ticker() string {
	return r.listing.Ticker
}

func (r *ListingResolver) Currency() string {
	return r.listing.Currency
}

func (r *ListingResolver) QuoteProvider() *string {
	return stringPtr(r.listing.QuoteProvider)
}

// LatestQuote resolves Listing.latestQuote as the newest stored quote.
func (r *ListingResolver) LatestQuote(ctx context.Context) (*QuoteResolver, error) {
	quote, err := r.db.GetLatestQuote(ctx, r.listing.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &QuoteResolver{quote: quote, currency: r.listing.Currency}, nil
}

func (r *QuoteResolver) Time() graphql.Time {
	return graphql.Time{Time: r.quote.Time}
}

func (r *QuoteResolver) Price() Money {
	return money(r.quote.Price, r.currency)
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}

	return sql.NullString{String: *s, Valid: true}
}

func stringPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}

	return &s.String
}
