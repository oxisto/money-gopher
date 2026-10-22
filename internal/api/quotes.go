package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// QuoteUpdateResultResolver resolves the QuoteUpdateResult GraphQL type.
type QuoteUpdateResultResolver struct {
	updated int32
	errs    []string
}

// TriggerQuoteUpdate resolves Mutation.triggerQuoteUpdate. Failed listings
// are reported in the result instead of failing the whole mutation, so one
// broken provider does not hide the successful updates.
func (r *RootResolver) TriggerQuoteUpdate(ctx context.Context, args struct{ AssetIDs *[]graphql.ID }) (*QuoteUpdateResultResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var updated int
	if args.AssetIDs == nil {
		updated, err = r.quotes.UpdateAll(ctx)
	} else {
		ids := make([]string, 0, len(*args.AssetIDs))
		for _, id := range *args.AssetIDs {
			// Only what the person may see: a shared security or their own asset.
			if _, err := r.db.GetSecurityVisibleTo(ctx, persistence.GetSecurityVisibleToParams{ID: string(id), PersonID: ownedBy(person)}); errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("security %q not found", id)
			} else if err != nil {
				return nil, err
			}
			ids = append(ids, string(id))
		}
		updated, err = r.quotes.UpdateSecurities(ctx, ids)
	}

	result := &QuoteUpdateResultResolver{updated: int32(updated), errs: []string{}}
	if err != nil {
		// errors.Join separates the per-listing errors with newlines.
		result.errs = strings.Split(err.Error(), "\n")
	}

	return result, nil
}

func (r *QuoteUpdateResultResolver) UpdatedListings() int32 {
	return r.updated
}

func (r *QuoteUpdateResultResolver) Errors() []string {
	return r.errs
}

// RecordQuote resolves Mutation.recordQuote: a price entered by hand. For a
// listing without a provider this is how an asset nobody quotes — the
// office room, the car — gets valued at all; for one with a provider it
// corrects the feed. The quote at that time is replaced.
func (r *RootResolver) RecordQuote(ctx context.Context, args struct {
	ListingID graphql.ID
	Time      graphql.Time
	Price     Decimal
}) (*QuoteResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	listing, err := r.db.GetListingVisibleTo(ctx, persistence.GetListingVisibleToParams{ID: string(args.ListingID), PersonID: ownedBy(person)})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("listing %q not found", args.ListingID)
	} else if err != nil {
		return nil, err
	}

	price := persistence.Decimal(args.Price)
	if !price.Valid || price.Sign() <= 0 {
		return nil, errors.New("the price must be positive")
	}

	quote := &persistence.Quote{ListingID: listing.ID, Time: args.Time.Time.UTC(), Price: price}
	if err := r.db.CreateQuote(ctx, persistence.CreateQuoteParams{
		ListingID: quote.ListingID, Time: quote.Time, Price: quote.Price,
	}); err != nil {
		return nil, err
	}

	return &QuoteResolver{quote: quote, currency: listing.Currency}, nil
}

// DeleteQuote resolves Mutation.deleteQuote.
func (r *RootResolver) DeleteQuote(ctx context.Context, args struct {
	ListingID graphql.ID
	Time      graphql.Time
}) (bool, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return false, err
	}
	if _, err := r.db.GetListingVisibleTo(ctx, persistence.GetListingVisibleToParams{ID: string(args.ListingID), PersonID: ownedBy(person)}); errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	n, err := r.db.DeleteQuote(ctx, persistence.DeleteQuoteParams{
		ListingID: string(args.ListingID), Time: args.Time.Time.UTC(),
	})
	return n > 0, err
}

// Quotes resolves Listing.quotes, oldest first.
func (r *ListingResolver) Quotes(ctx context.Context, args struct{ From, To *graphql.Time }) ([]*QuoteResolver, error) {
	var (
		from time.Time
		to   = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)
	)
	if args.From != nil {
		from = args.From.Time
	}
	if args.To != nil {
		to = args.To.Time
	}

	quotes, err := r.db.ListQuotes(ctx, persistence.ListQuotesParams{
		ListingID: r.listing.ID,
		Time:      from,
		Time_2:    to,
	})
	if err != nil {
		return nil, err
	}

	resolvers := make([]*QuoteResolver, 0, len(quotes))
	for _, quote := range quotes {
		resolvers = append(resolvers, &QuoteResolver{quote: quote, currency: r.listing.Currency})
	}

	return resolvers, nil
}
