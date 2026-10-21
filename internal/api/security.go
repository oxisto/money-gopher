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

// ListingInput is one exchange listing in security mutations. See its
// schema doc comment for how ID drives create/update/delete reconciliation
// in updateSecurity.
type ListingInput struct {
	ID            *graphql.ID
	Exchange      *string
	Ticker        string
	Currency      string
	QuoteProvider *string
}

// CreateSecurityInput is the input for Mutation.createSecurity.
type CreateSecurityInput struct {
	DisplayName          string
	Kind                 *string
	Identifiers          *[]SecurityIdentifierInput
	Listings             *[]ListingInput
	UnderlyingSecurityID *graphql.ID
}

// CreateSecurity resolves Mutation.createSecurity.
func (r *RootResolver) CreateSecurity(ctx context.Context, args struct{ Input CreateSecurityInput }) (*SecurityResolver, error) {
	var security *persistence.Security

	err := r.db.Tx(ctx, func(q *persistence.Queries) (err error) {
		security, err = q.CreateSecurity(ctx, persistence.CreateSecurityParams{
			ID:                   uuid.NewString(),
			DisplayName:          args.Input.DisplayName,
			Kind:                 nullString(args.Input.Kind),
			UnderlyingSecurityID: idToNullString(args.Input.UnderlyingSecurityID),
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
	DisplayName          *string
	Kind                 *string
	Identifiers          *[]SecurityIdentifierInput
	Listings             *[]ListingInput
	UnderlyingSecurityID *graphql.ID
}

// UpdateSecurity resolves Mutation.updateSecurity.
func (r *RootResolver) UpdateSecurity(ctx context.Context, args struct {
	ID    graphql.ID
	Input UpdateSecurityInput
}) (*SecurityResolver, error) {
	var security *persistence.Security

	// Fields not in the input are written back as read, so the read and the
	// write must see the same snapshot — otherwise a concurrent update to them
	// in between is silently undone.
	err := r.db.TxWith(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead}, func(q *persistence.Queries) (err error) {
		security, err = q.GetSecurity(ctx, string(args.ID))
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("security %q not found", args.ID)
		} else if err != nil {
			return err
		}

		if args.Input.DisplayName != nil || args.Input.Kind != nil || args.Input.UnderlyingSecurityID != nil {
			displayName := security.DisplayName
			if args.Input.DisplayName != nil {
				displayName = *args.Input.DisplayName
			}
			kind := security.Kind
			if args.Input.Kind != nil {
				kind = nullString(args.Input.Kind)
			}
			underlying := security.UnderlyingSecurityID
			if args.Input.UnderlyingSecurityID != nil {
				underlying = idToNullString(args.Input.UnderlyingSecurityID)
			}

			security, err = q.UpdateSecurity(ctx, persistence.UpdateSecurityParams{
				DisplayName:          displayName,
				UnderlyingSecurityID: underlying,
				Kind:                 kind,
				ID:                   string(args.ID),
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

		if args.Input.Listings != nil {
			if err = reconcileListings(ctx, q, security.ID, *args.Input.Listings); err != nil {
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

// reconcileListings applies the desired listing set to a security that may
// already have listings, matching by ListingInput.ID rather than blindly
// replacing — a matched listing is updated in place (keeping its ID, and so
// its quote history, intact), an unmatched one is created, and an existing
// listing missing from the input set is removed.
func reconcileListings(ctx context.Context, q *persistence.Queries, securityID string, listings []ListingInput) error {
	existing, err := q.ListListings(ctx, securityID)
	if err != nil {
		return err
	}
	stillPresent := make(map[string]bool, len(existing))
	for _, e := range existing {
		stillPresent[e.ID] = false
	}

	for _, l := range listings {
		if l.ID != nil {
			if _, ok := stillPresent[string(*l.ID)]; ok {
				stillPresent[string(*l.ID)] = true
				if _, err := q.UpdateListing(ctx, persistence.UpdateListingParams{
					ID:            string(*l.ID),
					SecurityID:    securityID,
					Exchange:      nullString(l.Exchange),
					Ticker:        l.Ticker,
					Currency:      l.Currency,
					QuoteProvider: nullString(l.QuoteProvider),
				}); err != nil {
					return err
				}
				continue
			}
		}

		if _, err := q.CreateListing(ctx, persistence.CreateListingParams{
			ID:            uuid.NewString(),
			SecurityID:    securityID,
			Exchange:      nullString(l.Exchange),
			Ticker:        l.Ticker,
			Currency:      l.Currency,
			QuoteProvider: nullString(l.QuoteProvider),
		}); err != nil {
			return err
		}
	}

	for id, kept := range stillPresent {
		if kept {
			continue
		}
		if _, err := q.DeleteListing(ctx, persistence.DeleteListingParams{ID: id, SecurityID: securityID}); err != nil {
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

func (r *SecurityResolver) Kind() *string {
	if !r.security.Kind.Valid {
		return nil
	}
	return &r.security.Kind.String
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

// UnderlyingSecurity resolves Security.underlyingSecurity — the security
// this one derives from (a right, option or warrant's underlying stock),
// null when this security isn't a derivative of anything else tracked here.
func (r *SecurityResolver) UnderlyingSecurity(ctx context.Context) (*SecurityResolver, error) {
	if !r.security.UnderlyingSecurityID.Valid {
		return nil, nil
	}

	underlying, err := r.db.GetSecurity(ctx, r.security.UnderlyingSecurityID.String)
	if err != nil {
		return nil, err
	}

	return &SecurityResolver{db: r.db, security: underlying}, nil
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

func (r *QuoteResolver) Price() Price {
	return price(r.quote.Price, r.currency)
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}

	return sql.NullString{String: *s, Valid: true}
}

func idToNullString(id *graphql.ID) sql.NullString {
	if id == nil {
		return sql.NullString{}
	}

	return sql.NullString{String: string(*id), Valid: true}
}

func stringPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}

	return &s.String
}
