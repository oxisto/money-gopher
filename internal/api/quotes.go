package api

import (
	"context"
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
func (r *RootResolver) TriggerQuoteUpdate(ctx context.Context, args struct{ SecurityIDs *[]graphql.ID }) (*QuoteUpdateResultResolver, error) {
	if _, err := auth.UserFromContext(ctx); err != nil {
		return nil, err
	}

	var (
		updated int
		err     error
	)
	if args.SecurityIDs == nil {
		updated, err = r.quotes.UpdateAll(ctx)
	} else {
		ids := make([]string, 0, len(*args.SecurityIDs))
		for _, id := range *args.SecurityIDs {
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
