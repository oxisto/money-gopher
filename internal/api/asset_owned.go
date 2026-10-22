package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// A security is one kind of asset. A personal asset — the office room
// rented out, the car, a loan given — is a securities row with an owner,
// visible to that person alone; rows without one are the shared, listed
// instruments. The resolvers here are the owned side: listing them, booking
// one, and saying who owns it.

// assetsPortfolioPurpose marks the one portfolio per person that personal
// assets are booked into.
const assetsPortfolioPurpose = "ASSETS"

// ownedBy is the owner column value for a person.
func ownedBy(person *persistence.Person) sql.NullString {
	return sql.NullString{String: person.ID, Valid: true}
}

// Assets resolves Query.assets: the active person's own.
func (r *RootResolver) Assets(ctx context.Context) ([]*AssetResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	assets, err := r.db.ListAssetsForPerson(ctx, ownedBy(person))
	if err != nil {
		return nil, err
	}

	resolvers := make([]*AssetResolver, 0, len(assets))
	for _, a := range assets {
		resolvers = append(resolvers, &AssetResolver{db: r.db, security: a})
	}
	return resolvers, nil
}

// Owner resolves Security.owner: null for a shared security.
func (r *AssetResolver) Owner(ctx context.Context) (*PersonResolver, error) {
	if !r.security.PersonID.Valid {
		return nil, nil
	}

	person, err := r.db.GetPerson(ctx, r.security.PersonID.String)
	if err != nil {
		return nil, err
	}
	return &PersonResolver{db: r.db, person: person}, nil
}

// BookAssetInput is the input for Mutation.bookAsset.
type BookAssetInput struct {
	DisplayName   string
	Kind          string
	Currency      string
	AcquiredAt    graphql.Time
	Cost          Decimal
	ValueToday    *Decimal
	CashAccountID *graphql.ID
}

// BookAsset resolves Mutation.bookAsset: everything that owning an asset
// without a broker document amounts to, in one transaction — the asset
// itself, its manual price series, the transfer-in at cost into the
// person's "Other assets" portfolio (created on first use), and today's
// valuation when given.
func (r *RootResolver) BookAsset(ctx context.Context, args struct{ Input BookAssetInput }) (*AssetResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	in := args.Input

	name := strings.TrimSpace(in.DisplayName)
	if name == "" {
		return nil, errors.New("the asset needs a name")
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if len(currency) != 3 {
		return nil, errors.New("the currency is the three-letter code the cost and valuations are in")
	}
	cost := persistence.Decimal(in.Cost)
	if !cost.Valid || cost.Sign() <= 0 {
		return nil, errors.New("the cost must be positive")
	}
	var valueToday persistence.Decimal
	if in.ValueToday != nil {
		valueToday = persistence.Decimal(*in.ValueToday)
		if !valueToday.Valid || valueToday.Sign() <= 0 {
			return nil, errors.New("the value must be positive")
		}
	}
	if in.AcquiredAt.Time.IsZero() {
		return nil, errors.New("say when the asset became yours")
	}

	var security *persistence.Security
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		portfolio, err := q.GetAssetsPortfolio(ctx, person.ID)
		if errors.Is(err, sql.ErrNoRows) {
			if in.CashAccountID == nil {
				return errors.New("pick the cash account the asset's income settles against; the Other assets portfolio is created with it")
			}
			account, err := q.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: string(*in.CashAccountID), PersonID: person.ID})
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("cash account %q not found", *in.CashAccountID)
			} else if err != nil {
				return err
			}
			portfolio, err = q.CreatePortfolio(ctx, persistence.CreatePortfolioParams{
				ID:            uuid.NewString(),
				PersonID:      person.ID,
				DisplayName:   "Other assets",
				CashAccountID: account.ID,
				Purpose:       sql.NullString{String: assetsPortfolioPurpose, Valid: true},
			})
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		security, err = q.CreateSecurity(ctx, persistence.CreateSecurityParams{
			ID:          uuid.NewString(),
			DisplayName: name,
			AssetKind:   sql.NullString{String: in.Kind, Valid: true},
			PersonID:    ownedBy(person),
		})
		if err != nil {
			return err
		}

		// The price series valuations go on: no provider, nobody quotes it.
		listing, err := q.CreateListing(ctx, persistence.CreateListingParams{
			ID:         uuid.NewString(),
			SecurityID: security.ID,
			Ticker:     tickerFor(name),
			Currency:   currency,
		})
		if err != nil {
			return err
		}

		// Ownership and cost basis: one unit, transferred in at cost. No
		// cash moves; the money was spent before the app knew about it.
		if _, err := q.CreateTransaction(ctx, persistence.CreateTransactionParams{
			ID:          uuid.NewString(),
			Type:        "DELIVERY_INBOUND",
			Time:        in.AcquiredAt.Time.UTC(),
			PortfolioID: sql.NullString{String: portfolio.ID, Valid: true},
			SecurityID:  sql.NullString{String: security.ID, Valid: true},
			Units:       1,
			Price:       cost,
			Currency:    currency,
			Source:      "MANUAL",
		}); err != nil {
			return err
		}

		if valueToday.Valid {
			if err := q.CreateQuote(ctx, persistence.CreateQuoteParams{
				ListingID: listing.ID,
				Time:      time.Now().UTC(),
				Price:     valueToday,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AssetResolver{db: r.db, security: security}, nil
}

var tickerChars = regexp.MustCompile(`[^A-Z0-9]+`)

// tickerFor names a manual price series after its asset: "Office room" →
// "OFFICE-ROOM". The same rule as the UI's tickerFor, so a series created
// either way looks alike.
func tickerFor(name string) string {
	t := strings.Trim(tickerChars.ReplaceAllString(strings.ToUpper(name), "-"), "-")
	if len(t) > 24 {
		t = strings.TrimRight(t[:24], "-")
	}
	if t == "" {
		return "ASSET"
	}
	return t
}
