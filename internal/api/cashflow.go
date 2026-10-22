package api

import (
	"context"
	"database/sql"
	"sort"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// The cash-flow view is the household's month as the schedules expect it:
// what comes into the cash accounts, what leaves them, what is moved into
// savings or investments, and what is left. It looks at the cash accounts
// only — the Giro, estimated or not — so a loan's interest charge and its
// payment from the Giro count once, as the payment.

// flowKind sorts a movement on a cash account into the three buckets.
type flowKind int

const (
	flowNone flowKind = iota
	flowIn
	flowOut
	flowSaved
)

// expectedKind classifies one occurrence of a schedule by its type and
// the kinds of the accounts it touches.
func expectedKind(s *persistence.Schedule, kinds map[string]string) flowKind {
	target := kinds[s.CashAccountID]
	switch s.Type {
	case "DEPOSIT_CASH", "DIVIDEND":
		if target == "CASH" {
			return flowIn
		}
	case "INTEREST":
		if target == "CASH" {
			return flowIn
		}
	case "WITHDRAW_CASH", "ACCOUNT_FEES":
		if target == "CASH" {
			return flowOut
		}
	case "TRANSFER":
		if !s.CounterpartAccountID.Valid || kinds[s.CounterpartAccountID.String] != "CASH" {
			return flowNone
		}
		switch target {
		case "SAVINGS":
			return flowSaved
		case "LOAN":
			return flowOut
		}
	}
	return flowNone
}

// actualKind classifies a real transaction on a cash account. A transfer
// leg is read by the account on the other side; a trade settled on the
// account is money invested, i.e. saved.
func actualKind(tx *persistence.Transaction, counterpartKind string) (flowKind, int64) {
	switch tx.Type {
	case "DEPOSIT_CASH":
		if tx.TransferID.Valid {
			switch counterpartKind {
			case "SAVINGS":
				return flowSaved, -tx.CashDelta // taken back out of savings
			case "CASH":
				return flowNone, 0
			}
		}
		return flowIn, tx.CashDelta
	case "DIVIDEND", "INTEREST":
		return flowIn, tx.CashDelta
	case "WITHDRAW_CASH":
		if tx.TransferID.Valid {
			switch counterpartKind {
			case "SAVINGS":
				return flowSaved, -tx.CashDelta
			case "LOAN":
				return flowOut, -tx.CashDelta
			case "CASH":
				return flowNone, 0
			}
		}
		return flowOut, -tx.CashDelta
	case "ACCOUNT_FEES":
		return flowOut, -tx.CashDelta
	case "BUY":
		return flowSaved, -tx.CashDelta
	case "SELL":
		return flowSaved, -tx.CashDelta
	}
	return flowNone, 0
}

// CashFlowMonthResolver resolves CashFlowMonth.
type CashFlowMonthResolver struct {
	month                            time.Time
	currency                         string
	in, out, saved                   int64
	flows                            []*ExpectedFlowResolver
	actual                           bool
	actualIn, actualOut, actualSaved int64
}

func (r *CashFlowMonthResolver) Month() graphql.Time  { return graphql.Time{Time: r.month} }
func (r *CashFlowMonthResolver) ExpectedIn() Money    { return money(r.in, r.currency) }
func (r *CashFlowMonthResolver) ExpectedOut() Money   { return money(r.out, r.currency) }
func (r *CashFlowMonthResolver) ExpectedSaved() Money { return money(r.saved, r.currency) }
func (r *CashFlowMonthResolver) ExpectedFree() Money  { return money(r.in-r.out-r.saved, r.currency) }
func (r *CashFlowMonthResolver) Flows() []*ExpectedFlowResolver {
	return r.flows
}
func (r *CashFlowMonthResolver) ActualIn() *Money    { return r.actualMoney(r.actualIn) }
func (r *CashFlowMonthResolver) ActualOut() *Money   { return r.actualMoney(r.actualOut) }
func (r *CashFlowMonthResolver) ActualSaved() *Money { return r.actualMoney(r.actualSaved) }
func (r *CashFlowMonthResolver) ActualFree() *Money {
	return r.actualMoney(r.actualIn - r.actualOut - r.actualSaved)
}
func (r *CashFlowMonthResolver) actualMoney(v int64) *Money {
	if !r.actual {
		return nil
	}
	m := money(v, r.currency)
	return &m
}

// CashFlowResolver resolves CashFlow.
type CashFlowResolver struct {
	currency  string
	from      time.Time
	months    []*CashFlowMonthResolver
	schedules []*ScheduleResolver
}

func (r *CashFlowResolver) Currency() string                 { return r.currency }
func (r *CashFlowResolver) From() graphql.Time               { return graphql.Time{Time: r.from} }
func (r *CashFlowResolver) Months() []*CashFlowMonthResolver { return r.months }
func (r *CashFlowResolver) Schedules() []*ScheduleResolver   { return r.schedules }

// CashFlow resolves Query.cashFlow.
func (r *RootResolver) CashFlow(ctx context.Context, args struct {
	From   *graphql.Time
	Months *int32
}) (*CashFlowResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	from := time.Date(now.Year(), now.Month()-5, 1, 0, 0, 0, 0, time.UTC)
	if args.From != nil {
		t := args.From.Time.UTC()
		from = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	months := 12
	if args.Months != nil && *args.Months > 0 && *args.Months <= 60 {
		months = int(*args.Months)
	}
	to := from.AddDate(0, months, 0)

	accounts, err := r.db.ListCashAccounts(ctx, person.ID)
	if err != nil {
		return nil, err
	}
	kinds := map[string]string{}
	currency := defaultCurrency
	var tracked []*persistence.CashAccount
	for _, a := range accounts {
		kinds[a.ID] = kindOf(a)
		if kindOf(a) == "CASH" {
			currency = a.Currency
			if !a.EstimatedAt.Valid {
				tracked = append(tracked, a)
			}
		}
	}

	schedules, err := r.db.ListSchedules(ctx, person.ID)
	if err != nil {
		return nil, err
	}
	buckets := make([]*CashFlowMonthResolver, months)
	for i := range buckets {
		buckets[i] = &CashFlowMonthResolver{month: from.AddDate(0, i, 0), currency: currency, flows: []*ExpectedFlowResolver{}}
	}
	index := func(t time.Time) int {
		return (t.Year()-from.Year())*12 + int(t.Month()-from.Month())
	}
	var resolvers []*ScheduleResolver
	for _, s := range schedules {
		kind := expectedKind(s, kinds)
		sr := &ScheduleResolver{db: r.db, s: s}
		resolvers = append(resolvers, sr)
		if kind == flowNone {
			continue
		}
		flows, err := sr.Flows(ctx, struct{ From, To graphql.Time }{graphql.Time{Time: from}, graphql.Time{Time: to}})
		if err != nil {
			return nil, err
		}
		for _, f := range flows {
			i := index(f.date)
			if i < 0 || i >= months {
				continue
			}
			amount := int64(f.amount.Amount)
			if amount < 0 {
				amount = -amount
			}
			b := buckets[i]
			switch kind {
			case flowIn:
				b.in += amount
			case flowOut:
				b.out += amount
			case flowSaved:
				b.saved += amount
			}
			b.flows = append(b.flows, f)
		}
	}

	// Actuals: months over, on the tracked cash accounts.
	if len(tracked) > 0 {
		for _, a := range tracked {
			txs, err := r.db.ListTransactionsByCashAccount(ctx, sql.NullString{String: a.ID, Valid: true})
			if err != nil {
				return nil, err
			}
			for _, tx := range txs {
				i := index(tx.Time.UTC())
				if i < 0 || i >= months {
					continue
				}
				counterpart := ""
				if tx.TransferID.Valid {
					if other, err := r.db.GetTransferCounterpart(ctx, persistence.GetTransferCounterpartParams{TransferID: tx.TransferID, ID: tx.ID}); err == nil && other.CashAccountID.Valid {
						counterpart = kinds[other.CashAccountID.String]
					}
				}
				kind, amount := actualKind(tx, counterpart)
				switch kind {
				case flowIn:
					buckets[i].actualIn += amount
				case flowOut:
					buckets[i].actualOut += amount
				case flowSaved:
					buckets[i].actualSaved += amount
				}
			}
		}
		thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		for _, b := range buckets {
			b.actual = b.month.Before(thisMonth)
		}
	}
	for _, b := range buckets {
		sort.SliceStable(b.flows, func(i, j int) bool { return b.flows[i].date.Before(b.flows[j].date) })
	}
	sort.SliceStable(resolvers, func(i, j int) bool { return resolvers[i].s.DisplayName < resolvers[j].s.DisplayName })
	return &CashFlowResolver{currency: currency, from: from, months: buckets, schedules: resolvers}, nil
}
