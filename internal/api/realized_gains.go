package api

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/finance"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// RealizedGains resolves Query.realizedGains: dividends, interest, and the
// profit/loss of completed sales across every portfolio and cash account of
// the active person. Unlike PortfolioSnapshot's unrealized gains, these are
// cash events that already happened, not a mark-to-market valuation.
// Defaults to the current calendar year to date.
func (r *RootResolver) RealizedGains(ctx context.Context, args struct{ From, To *graphql.Time }) (*RealizedGainsResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	from := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
	to := now
	if args.From != nil {
		from = args.From.Time
	}
	if args.To != nil {
		to = args.To.Time
	}

	txs, err := allPersonTransactions(ctx, r.db, person.ID)
	if err != nil {
		return nil, err
	}
	txs, err = withoutLoanInterest(ctx, r.db, person.ID, txs)
	if err != nil {
		return nil, err
	}

	return &RealizedGainsResolver{
		db:    r.db,
		from:  from,
		to:    to,
		gains: finance.RealizedGains(txs, from, to),
	}, nil
}

// allPersonTransactions gathers every transaction belonging to a person: all
// portfolio transactions (trades, dividends) plus all cash account
// transactions (interest, fees, transfers), deduplicated by ID — a trade
// also settles against its portfolio's cash account, so it would otherwise
// be listed twice.
func allPersonTransactions(ctx context.Context, db *persistence.DB, personID string) ([]*persistence.Transaction, error) {
	portfolios, err := db.ListPortfolios(ctx, personID)
	if err != nil {
		return nil, err
	}
	accounts, err := db.ListCashAccounts(ctx, personID)
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var out []*persistence.Transaction
	add := func(txs []*persistence.Transaction) {
		for _, tx := range txs {
			if seen[tx.ID] {
				continue
			}
			seen[tx.ID] = true
			out = append(out, tx)
		}
	}

	for _, p := range portfolios {
		txs, err := db.ListTransactionsByPortfolio(ctx, sql.NullString{String: p.ID, Valid: true})
		if err != nil {
			return nil, err
		}
		add(txs)
	}
	for _, a := range accounts {
		txs, err := db.ListTransactionsByCashAccount(ctx, sql.NullString{String: a.ID, Valid: true})
		if err != nil {
			return nil, err
		}
		add(txs)
	}

	return out, nil
}

// RealizedGainsResolver resolves the RealizedGains GraphQL type.
type RealizedGainsResolver struct {
	db    *persistence.DB
	from  time.Time
	to    time.Time
	gains []finance.RealizedGain
}

func (r *RealizedGainsResolver) From() graphql.Time { return graphql.Time{Time: r.from} }
func (r *RealizedGainsResolver) To() graphql.Time   { return graphql.Time{Time: r.to} }

// currency takes the currency from the first gain found; empty (no realized
// gains in the period) falls back to defaultCurrency, as snapshot totals do.
func (r *RealizedGainsResolver) currency() string {
	for _, g := range r.gains {
		if g.Currency != "" {
			return g.Currency
		}
	}
	return defaultCurrency
}

func (r *RealizedGainsResolver) Currency() string { return r.currency() }

func (r *RealizedGainsResolver) sum(typ string) (v int64) {
	for _, g := range r.gains {
		if g.Type == typ {
			v += g.Amount
		}
	}
	return
}

func (r *RealizedGainsResolver) Dividends() Money {
	return money(r.sum("DIVIDEND"), r.currency())
}

func (r *RealizedGainsResolver) Interest() Money {
	return money(r.sum("INTEREST"), r.currency())
}

func (r *RealizedGainsResolver) TradingGains() Money {
	return money(r.sum("SELL"), r.currency())
}

func (r *RealizedGainsResolver) Total() Money {
	var total int64
	for _, g := range r.gains {
		total += g.Amount
	}
	return money(total, r.currency())
}

// Entries resolves RealizedGains.entries, most recent first.
func (r *RealizedGainsResolver) Entries(ctx context.Context) ([]*RealizedGainResolver, error) {
	resolvers := make([]*RealizedGainResolver, 0, len(r.gains))
	for _, g := range r.gains {
		var sec *persistence.Security
		if g.SecurityID != "" {
			s, err := r.db.GetSecurity(ctx, g.SecurityID)
			if err != nil {
				return nil, err
			}
			sec = s
		}
		resolvers = append(resolvers, &RealizedGainResolver{db: r.db, gain: g, security: sec})
	}

	sort.Slice(resolvers, func(i, j int) bool { return resolvers[i].gain.Time.After(resolvers[j].gain.Time) })

	return resolvers, nil
}

// RealizedGainResolver resolves the RealizedGain GraphQL type.
type RealizedGainResolver struct {
	db       *persistence.DB
	gain     finance.RealizedGain
	security *persistence.Security
}

func (r *RealizedGainResolver) ID() graphql.ID     { return graphql.ID(r.gain.TransactionID) }
func (r *RealizedGainResolver) Time() graphql.Time { return graphql.Time{Time: r.gain.Time} }
func (r *RealizedGainResolver) Type() string       { return r.gain.Type }

func (r *RealizedGainResolver) Asset() *AssetResolver {
	if r.security == nil {
		return nil
	}
	return &AssetResolver{db: r.db, security: r.security}
}

// Portfolio resolves RealizedGain.portfolio. Re-scoped to the active person
// here (rather than trusted from the already-person-scoped transaction
// lookup) to match every other resolver's belt-and-suspenders pattern.
func (r *RealizedGainResolver) Portfolio(ctx context.Context) (*PortfolioResolver, error) {
	if r.gain.PortfolioID == "" {
		return nil, nil
	}
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	portfolio, err := r.db.GetPortfolio(ctx, persistence.GetPortfolioParams{
		ID: r.gain.PortfolioID, PersonID: person.ID,
	})
	if err != nil {
		return nil, err
	}
	return &PortfolioResolver{db: r.db, portfolio: portfolio}, nil
}

// CashAccount resolves RealizedGain.cashAccount.
func (r *RealizedGainResolver) CashAccount(ctx context.Context) (*CashAccountResolver, error) {
	if r.gain.CashAccountID == "" {
		return nil, nil
	}
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	account, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{
		ID: r.gain.CashAccountID, PersonID: person.ID,
	})
	if err != nil {
		return nil, err
	}
	return &CashAccountResolver{db: r.db, account: account}, nil
}

func (r *RealizedGainResolver) Amount() Money {
	return money(r.gain.Amount, r.gain.Currency)
}

// withoutLoanInterest drops the INTEREST transactions of loan accounts:
// interest charged is a financing cost, not income, and would otherwise
// net against the interest earned on the Income page.
func withoutLoanInterest(ctx context.Context, db *persistence.DB, personID string, txs []*persistence.Transaction) ([]*persistence.Transaction, error) {
	accounts, err := db.ListCashAccounts(ctx, personID)
	if err != nil {
		return nil, err
	}
	loans := map[string]bool{}
	for _, a := range accounts {
		if kindOf(a) == "LOAN" {
			loans[a.ID] = true
		}
	}
	if len(loans) == 0 {
		return txs, nil
	}
	kept := txs[:0:0]
	for _, tx := range txs {
		if tx.Type == "INTEREST" && tx.CashAccountID.Valid && loans[tx.CashAccountID.String] {
			continue
		}
		kept = append(kept, tx)
	}
	return kept, nil
}
