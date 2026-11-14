package api

import (
	"context"
	"database/sql"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/finance"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/plan"
)

// A plan projects the whole net worth, bucketed by kind of wealth. Every
// bucket carries a compound annual return and a volatility; the defaults
// are rough long-run figures the owner is expected to adjust.

// planBuckets lists the buckets in the order the assumptions are shown.
var planBuckets = []string{"STOCK", "ETF", "FUND", "BOND", "PROPERTY", "VEHICLE", "LENT", "OTHER", "CASH", "SAVINGS", "DEBT"}

// planDefaults holds each bucket's default return and volatility in
// basis points. A debt has no return: its interest is a cash-flow
// expense, its repayment a transfer.
var planDefaults = map[string][2]int64{
	"STOCK": {700, 1800}, "ETF": {600, 1500}, "FUND": {500, 1200}, "BOND": {250, 500},
	"PROPERTY": {200, 400}, "VEHICLE": {-1000, 0}, "LENT": {300, 0}, "OTHER": {0, 0},
	"CASH": {150, 0}, "SAVINGS": {200, 0}, "DEBT": {0, 0},
}

const (
	planPaths      = 1000
	defaultHorizon = 240
	maxHorizon     = 600
	maxPlanCost    = 0.1
	bpPerUnit      = 10000
)

// bucketOfAsset maps an asset kind onto a plan bucket.
func bucketOfAsset(kind sql.NullString) string {
	if !kind.Valid {
		return "OTHER"
	}
	switch kind.String {
	case "LOAN":
		return "LENT"
	case "STOCK", "ETF", "FUND", "BOND", "PROPERTY", "VEHICLE":
		return kind.String
	}
	return "OTHER"
}

// bucketOfAccount maps a cash account's kind onto a plan bucket.
func bucketOfAccount(a *persistence.CashAccount) string {
	switch kindOf(a) {
	case "SAVINGS":
		return "SAVINGS"
	case "LOAN":
		return "DEBT"
	}
	return "CASH"
}

func isPlanBucket(b string) bool {
	_, ok := planDefaults[b]
	return ok
}

// monthStart is the first of the month of t, midnight UTC.
func monthStart(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func toBp(f float64) int64    { return int64(math.Round(f * bpPerUnit)) }
func fromBp(bp int64) float64 { return float64(bp) / bpPerUnit }

// PlanAssumptionInput mirrors the GraphQL input of the same name.
type PlanAssumptionInput struct {
	Bucket       string
	Weight       *float64
	AnnualReturn *float64
	Volatility   *float64
}

// CreatePlanInput mirrors the GraphQL input of the same name.
type CreatePlanInput struct {
	DisplayName    string
	Start          *graphql.Time
	HorizonMonths  *int32
	StartingAmount *int32
	Contribution   *int32
	AnnualCost     *float64
	Assumptions    *[]PlanAssumptionInput
}

// UpdatePlanInput mirrors the GraphQL input of the same name.
type UpdatePlanInput struct {
	DisplayName         *string
	Start               *graphql.Time
	HorizonMonths       *int32
	StartingAmount      *int32
	ClearStartingAmount *bool
	Contribution        *int32
	ClearContribution   *bool
	AnnualCost          *float64
	Assumptions         *[]PlanAssumptionInput
}

// assumptionRows turns the input into one row per bucket: the given
// values where declared, the bucket's defaults otherwise.
func assumptionRows(planID string, in []PlanAssumptionInput) ([]*persistence.PlanAssumption, error) {
	given := map[string]PlanAssumptionInput{}
	for _, a := range in {
		if !isPlanBucket(a.Bucket) {
			return nil, fmt.Errorf("unknown bucket %q", a.Bucket)
		}
		if _, dup := given[a.Bucket]; dup {
			return nil, fmt.Errorf("bucket %s given twice", a.Bucket)
		}
		if a.Weight != nil && (*a.Weight < 0 || *a.Weight > 1) {
			return nil, fmt.Errorf("weight of %s must be between 0 and 1", a.Bucket)
		}
		if a.AnnualReturn != nil && (*a.AnnualReturn < -1 || *a.AnnualReturn > 1) {
			return nil, fmt.Errorf("annual return of %s must be between -1 and 1", a.Bucket)
		}
		if a.Volatility != nil && *a.Volatility < 0 {
			return nil, fmt.Errorf("volatility of %s must not be negative", a.Bucket)
		}
		given[a.Bucket] = a
	}
	rows := make([]*persistence.PlanAssumption, 0, len(planBuckets))
	for _, b := range planBuckets {
		row := &persistence.PlanAssumption{PlanID: planID, Bucket: b, ReturnBp: planDefaults[b][0], VolatilityBp: planDefaults[b][1]}
		if a, ok := given[b]; ok {
			if a.Weight != nil {
				row.WeightBp = sql.NullInt64{Int64: toBp(*a.Weight), Valid: true}
			}
			if a.AnnualReturn != nil {
				row.ReturnBp = toBp(*a.AnnualReturn)
			}
			if a.Volatility != nil {
				row.VolatilityBp = toBp(*a.Volatility)
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func validatePlan(p *persistence.Plan) error {
	if p.DisplayName == "" {
		return fmt.Errorf("a plan needs a name")
	}
	if p.HorizonMonths < 1 || p.HorizonMonths > maxHorizon {
		return fmt.Errorf("horizon must be between 1 and %d months, got %d", maxHorizon, p.HorizonMonths)
	}
	if p.CostBp < 0 || fromBp(p.CostBp) > maxPlanCost {
		return fmt.Errorf("annual cost must be between 0 and %v", maxPlanCost)
	}
	return nil
}

// replaceAssumptions writes the rows as the plan's only assumptions.
func replaceAssumptions(ctx context.Context, q *persistence.Queries, planID string, rows []*persistence.PlanAssumption) error {
	if err := q.DeletePlanAssumptions(ctx, planID); err != nil {
		return err
	}
	for _, a := range rows {
		if err := q.CreatePlanAssumption(ctx, persistence.CreatePlanAssumptionParams{
			PlanID: planID, Bucket: a.Bucket, WeightBp: a.WeightBp, ReturnBp: a.ReturnBp, VolatilityBp: a.VolatilityBp,
		}); err != nil {
			return err
		}
	}
	return nil
}

// CreatePlan resolves Mutation.createPlan.
func (r *RootResolver) CreatePlan(ctx context.Context, args struct{ Input CreatePlanInput }) (*PlanResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	in := args.Input
	p := &persistence.Plan{
		ID:            uuid.NewString(),
		PersonID:      person.ID,
		DisplayName:   in.DisplayName,
		StartDate:     monthStart(time.Now()),
		HorizonMonths: defaultHorizon,
		CreatedAt:     time.Now().UTC(),
	}
	if in.Start != nil {
		p.StartDate = monthStart(in.Start.Time)
	}
	if in.HorizonMonths != nil {
		p.HorizonMonths = int64(*in.HorizonMonths)
	}
	if in.StartingAmount != nil {
		p.StartingAmount = sql.NullInt64{Int64: int64(*in.StartingAmount), Valid: true}
	}
	if in.Contribution != nil {
		p.Contribution = sql.NullInt64{Int64: int64(*in.Contribution), Valid: true}
	}
	if in.AnnualCost != nil {
		p.CostBp = toBp(*in.AnnualCost)
	}
	if err := validatePlan(p); err != nil {
		return nil, err
	}
	var assumptions []PlanAssumptionInput
	if in.Assumptions != nil {
		assumptions = *in.Assumptions
	}
	rows, err := assumptionRows(p.ID, assumptions)
	if err != nil {
		return nil, err
	}

	var created *persistence.Plan
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		var err error
		created, err = q.CreatePlan(ctx, persistence.CreatePlanParams{
			ID: p.ID, PersonID: p.PersonID, DisplayName: p.DisplayName, StartDate: p.StartDate, HorizonMonths: p.HorizonMonths,
			StartingAmount: p.StartingAmount, Contribution: p.Contribution, CostBp: p.CostBp, CreatedAt: p.CreatedAt,
		})
		if err != nil {
			return err
		}
		return replaceAssumptions(ctx, q, p.ID, rows)
	})
	if err != nil {
		return nil, err
	}
	return newPlanResolver(r.db, created), nil
}

// UpdatePlan resolves Mutation.updatePlan.
func (r *RootResolver) UpdatePlan(ctx context.Context, args struct {
	ID    graphql.ID
	Input UpdatePlanInput
}) (*PlanResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	p, err := r.db.GetPlan(ctx, persistence.GetPlanParams{ID: string(args.ID), PersonID: person.ID})
	if err != nil {
		return nil, fmt.Errorf("plan %q not found", args.ID)
	}
	in := args.Input
	if in.DisplayName != nil {
		p.DisplayName = *in.DisplayName
	}
	if in.Start != nil {
		p.StartDate = monthStart(in.Start.Time)
	}
	if in.HorizonMonths != nil {
		p.HorizonMonths = int64(*in.HorizonMonths)
	}
	if in.StartingAmount != nil {
		p.StartingAmount = sql.NullInt64{Int64: int64(*in.StartingAmount), Valid: true}
	}
	if in.ClearStartingAmount != nil && *in.ClearStartingAmount {
		p.StartingAmount = sql.NullInt64{}
	}
	if in.Contribution != nil {
		p.Contribution = sql.NullInt64{Int64: int64(*in.Contribution), Valid: true}
	}
	if in.ClearContribution != nil && *in.ClearContribution {
		p.Contribution = sql.NullInt64{}
	}
	if in.AnnualCost != nil {
		p.CostBp = toBp(*in.AnnualCost)
	}
	if err := validatePlan(p); err != nil {
		return nil, err
	}
	var rows []*persistence.PlanAssumption
	if in.Assumptions != nil {
		if rows, err = assumptionRows(p.ID, *in.Assumptions); err != nil {
			return nil, err
		}
	}

	var updated *persistence.Plan
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		var err error
		updated, err = q.UpdatePlan(ctx, persistence.UpdatePlanParams{
			DisplayName: p.DisplayName, StartDate: p.StartDate, HorizonMonths: p.HorizonMonths,
			StartingAmount: p.StartingAmount, Contribution: p.Contribution, CostBp: p.CostBp,
			ID: p.ID, PersonID: person.ID,
		})
		if err != nil {
			return err
		}
		if rows != nil {
			return replaceAssumptions(ctx, q, p.ID, rows)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return newPlanResolver(r.db, updated), nil
}

// DeletePlan resolves Mutation.deletePlan.
func (r *RootResolver) DeletePlan(ctx context.Context, args struct{ ID graphql.ID }) (graphql.ID, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return "", err
	}
	rows, err := r.db.DeletePlan(ctx, persistence.DeletePlanParams{ID: string(args.ID), PersonID: person.ID})
	if err != nil {
		return "", err
	}
	if rows == 0 {
		return "", fmt.Errorf("plan %q not found", args.ID)
	}
	return args.ID, nil
}

// Plans resolves Query.plans.
func (r *RootResolver) Plans(ctx context.Context) ([]*PlanResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	plans, err := r.db.ListPlans(ctx, person.ID)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*PlanResolver, 0, len(plans))
	for _, p := range plans {
		resolvers = append(resolvers, newPlanResolver(r.db, p))
	}
	return resolvers, nil
}

// Plan resolves Query.plan.
func (r *RootResolver) Plan(ctx context.Context, args struct{ ID graphql.ID }) (*PlanResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	p, err := r.db.GetPlan(ctx, persistence.GetPlanParams{ID: string(args.ID), PersonID: person.ID})
	if err != nil {
		return nil, nil
	}
	return newPlanResolver(r.db, p), nil
}

// PlanResolver resolves Plan. The allocation, the net worth inputs and
// the projection are computed once per resolver, since several fields
// share them.
type PlanResolver struct {
	db *persistence.DB
	p  *persistence.Plan

	once struct{ assumptions, allocation, netWorth, contribution, projection sync.Once }
	err  struct{ assumptions, allocation, netWorth, contribution error }

	assumptions []*persistence.PlanAssumption
	allocation  map[string]int64
	currency    string
	txs         []*persistence.Transaction
	estimated   map[string]int64
	quote       finance.QuoteFunc
	startNW     int64 // the net worth at start, or now when start is ahead
	nowNW       int64
	flows       int64 // the contribution from the schedules
	projection  []plan.Point
}

func newPlanResolver(db *persistence.DB, p *persistence.Plan) *PlanResolver {
	return &PlanResolver{db: db, p: p}
}

func (r *PlanResolver) ID() graphql.ID               { return graphql.ID(r.p.ID) }
func (r *PlanResolver) DisplayName() string          { return r.p.DisplayName }
func (r *PlanResolver) Start() graphql.Time          { return graphql.Time{Time: r.p.StartDate} }
func (r *PlanResolver) HorizonMonths() int32         { return int32(r.p.HorizonMonths) }
func (r *PlanResolver) AnnualCost() float64          { return fromBp(r.p.CostBp) }
func (r *PlanResolver) StartingAmountDeclared() bool { return r.p.StartingAmount.Valid }
func (r *PlanResolver) ContributionDeclared() bool   { return r.p.Contribution.Valid }

func (r *PlanResolver) loadAssumptions(ctx context.Context) ([]*persistence.PlanAssumption, error) {
	r.once.assumptions.Do(func() {
		rows, err := r.db.ListPlanAssumptions(ctx, r.p.ID)
		if err != nil {
			r.err.assumptions = err
			return
		}
		byBucket := map[string]*persistence.PlanAssumption{}
		for _, a := range rows {
			byBucket[a.Bucket] = a
		}
		for _, b := range planBuckets {
			if _, ok := byBucket[b]; !ok {
				byBucket[b] = &persistence.PlanAssumption{PlanID: r.p.ID, Bucket: b, ReturnBp: planDefaults[b][0], VolatilityBp: planDefaults[b][1]}
			}
			r.assumptions = append(r.assumptions, byBucket[b])
		}
	})
	return r.assumptions, r.err.assumptions
}

// loadAllocation values every bucket as of now: each portfolio's
// positions by the security's asset kind, each cash account's balance by
// its kind (the declared one for an estimated account). The sum is the
// net worth now as netWorthHistory computes it.
func (r *PlanResolver) loadAllocation(ctx context.Context) (map[string]int64, string, error) {
	r.once.allocation.Do(func() {
		if _, _, _, err := r.loadNetWorth(ctx); err != nil {
			r.err.allocation = err
			return
		}
		alloc := map[string]int64{}
		for _, b := range planBuckets {
			alloc[b] = 0
		}
		now := time.Now()
		portfolios, err := r.db.ListPortfolios(ctx, r.p.PersonID)
		if err != nil {
			r.err.allocation = err
			return
		}
		for _, pf := range portfolios {
			txs, err := r.db.ListTransactionsByPortfolio(ctx, sql.NullString{String: pf.ID, Valid: true})
			if err != nil {
				r.err.allocation = err
				return
			}
			for _, pos := range finance.SnapshotAt(txs, now, r.quote).Positions {
				sec, err := securityByID(ctx, r.db, pos.SecurityID)
				if err != nil {
					r.err.allocation = err
					return
				}
				alloc[bucketOfAsset(sec.AssetKind)] += pos.MarketValue
			}
		}
		accounts, err := r.db.ListCashAccounts(ctx, r.p.PersonID)
		if err != nil {
			r.err.allocation = err
			return
		}
		for _, a := range accounts {
			balance := a.EstimatedBalance.Int64
			if !a.EstimatedBalance.Valid {
				if balance, err = r.db.GetCashAccountBalance(ctx, sql.NullString{String: a.ID, Valid: true}); err != nil {
					r.err.allocation = err
					return
				}
			}
			alloc[bucketOfAccount(a)] += balance
		}
		r.allocation = alloc
	})
	return r.allocation, r.currency, r.err.allocation
}

// loadNetWorth evaluates the net worth at the plan's start and now.
func (r *PlanResolver) loadNetWorth(ctx context.Context) (startNW, nowNW int64, currency string, err error) {
	r.once.netWorth.Do(func() {
		txs, estimated, currency, err := netWorthInputs(ctx, r.db, r.p.PersonID)
		if err != nil {
			r.err.netWorth = err
			return
		}
		r.txs, r.estimated, r.currency = txs, estimated, currency
		r.quote = latestQuote(ctx, r.db)
		now := time.Now()
		start := r.p.StartDate
		if start.After(now) {
			start = now
		}
		points := finance.NetWorthAt(txs, estimated, []time.Time{start, now}, r.quote)
		r.startNW, r.nowNW = points[0].NetWorth, points[1].NetWorth
	})
	return r.startNW, r.nowNW, r.currency, r.err.netWorth
}

// loadContribution averages the next twelve months' expected in − out
// from the cash-flow view: what stays on the cash accounts plus what
// moves to savings.
func (r *PlanResolver) loadContribution(ctx context.Context) (int64, error) {
	r.once.contribution.Do(func() {
		person, err := r.db.GetPerson(ctx, r.p.PersonID)
		if err != nil {
			r.err.contribution = err
			return
		}
		cf, err := cashFlowFor(ctx, r.db, person, monthStart(time.Now()), 12)
		if err != nil {
			r.err.contribution = err
			return
		}
		var sum int64
		for _, m := range cf.months {
			sum += m.in - m.out
		}
		r.flows = int64(math.Round(float64(sum) / float64(len(cf.months))))
	})
	return r.flows, r.err.contribution
}

// StartingAmount resolves Plan.startingAmount.
func (r *PlanResolver) StartingAmount(ctx context.Context) (Money, error) {
	startNW, _, currency, err := r.loadNetWorth(ctx)
	if err != nil {
		return Money{}, err
	}
	if r.p.StartingAmount.Valid {
		return money(r.p.StartingAmount.Int64, currency), nil
	}
	return money(startNW, currency), nil
}

// Contribution resolves Plan.contribution.
func (r *PlanResolver) Contribution(ctx context.Context) (Money, error) {
	_, _, currency, err := r.loadNetWorth(ctx)
	if err != nil {
		return Money{}, err
	}
	if r.p.Contribution.Valid {
		return money(r.p.Contribution.Int64, currency), nil
	}
	flows, err := r.loadContribution(ctx)
	if err != nil {
		return Money{}, err
	}
	return money(flows, currency), nil
}

// buckets joins the allocation and the assumptions into the engine's
// buckets, weights filled in. A declared starting amount that differs
// from the net worth now is spread over the buckets in proportion.
func (r *PlanResolver) buckets(ctx context.Context) ([]plan.Bucket, error) {
	assumptions, err := r.loadAssumptions(ctx)
	if err != nil {
		return nil, err
	}
	alloc, _, err := r.loadAllocation(ctx)
	if err != nil {
		return nil, err
	}
	buckets := make([]plan.Bucket, 0, len(assumptions))
	for _, a := range assumptions {
		b := plan.Bucket{Kind: a.Bucket, Value: alloc[a.Bucket], Return: fromBp(a.ReturnBp), Volatility: fromBp(a.VolatilityBp)}
		if a.WeightBp.Valid {
			b.Weight = fromBp(a.WeightBp.Int64)
		}
		buckets = append(buckets, b)
	}
	buckets = plan.DefaultSplit(buckets)
	if r.p.StartingAmount.Valid {
		var total int64
		for _, b := range buckets {
			total += b.Value
		}
		if total != 0 && total != r.p.StartingAmount.Int64 {
			f := float64(r.p.StartingAmount.Int64) / float64(total)
			for i := range buckets {
				buckets[i].Value = int64(math.Round(float64(buckets[i].Value) * f))
			}
		} else if total == 0 {
			// Nothing held: the declared amount starts out as cash.
			for i := range buckets {
				if buckets[i].Kind == "CASH" {
					buckets[i].Value = r.p.StartingAmount.Int64
				}
			}
		}
	}
	return buckets, nil
}

// Assumptions resolves Plan.assumptions.
func (r *PlanResolver) Assumptions(ctx context.Context) ([]*PlanAssumptionResolver, error) {
	assumptions, err := r.loadAssumptions(ctx)
	if err != nil {
		return nil, err
	}
	buckets, err := r.buckets(ctx)
	if err != nil {
		return nil, err
	}
	alloc, currency, err := r.loadAllocation(ctx)
	if err != nil {
		return nil, err
	}
	weights := map[string]float64{}
	for _, b := range buckets {
		weights[b.Kind] = b.Weight
	}
	resolvers := make([]*PlanAssumptionResolver, 0, len(assumptions))
	for _, a := range assumptions {
		resolvers = append(resolvers, &PlanAssumptionResolver{a: a, value: money(alloc[a.Bucket], currency), weight: weights[a.Bucket]})
	}
	return resolvers, nil
}

// PlanAssumptionResolver resolves PlanAssumption.
type PlanAssumptionResolver struct {
	a      *persistence.PlanAssumption
	value  Money
	weight float64
}

func (r *PlanAssumptionResolver) Bucket() string        { return r.a.Bucket }
func (r *PlanAssumptionResolver) Value() Money          { return r.value }
func (r *PlanAssumptionResolver) Weight() float64       { return r.weight }
func (r *PlanAssumptionResolver) WeightDeclared() bool  { return r.a.WeightBp.Valid }
func (r *PlanAssumptionResolver) AnnualReturn() float64 { return fromBp(r.a.ReturnBp) }
func (r *PlanAssumptionResolver) Volatility() float64   { return fromBp(r.a.VolatilityBp) }

// project runs the engine once per resolver. The projection starts from
// what is held now (or the declared amount), dated at the plan's start:
// a plan that started months ago is re-based on today's books every time
// it is read, so the path stays a forecast rather than a record.
func (r *PlanResolver) project(ctx context.Context) ([]plan.Point, error) {
	var err error
	r.once.projection.Do(func() {
		var buckets []plan.Bucket
		if buckets, err = r.buckets(ctx); err != nil {
			return
		}
		contribution, cerr := r.Contribution(ctx)
		if cerr != nil {
			err = cerr
			return
		}
		h := fnv.New64a()
		h.Write([]byte(r.p.ID))
		r.projection = plan.Project(plan.Input{
			Start: r.p.StartDate, Months: int(r.p.HorizonMonths), Buckets: buckets,
			Contribution: int64(contribution.Amount), Cost: fromBp(r.p.CostBp), Paths: planPaths, Seed: h.Sum64(),
		})
	})
	return r.projection, err
}

// Projection resolves Plan.projection.
func (r *PlanResolver) Projection(ctx context.Context) ([]*PlanPointResolver, error) {
	points, err := r.project(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*PlanPointResolver, len(points))
	for i := range points {
		resolvers[i] = &PlanPointResolver{point: points[i], currency: r.currency}
	}
	return resolvers, nil
}

// PlanPointResolver resolves PlanPoint.
type PlanPointResolver struct {
	point    plan.Point
	currency string
}

func (r *PlanPointResolver) Month() graphql.Time { return graphql.Time{Time: r.point.Month} }
func (r *PlanPointResolver) Expected() Money     { return money(r.point.Expected, r.currency) }
func (r *PlanPointResolver) Contributed() Money  { return money(r.point.Contributed, r.currency) }
func (r *PlanPointResolver) P10() Money          { return money(r.point.P10, r.currency) }
func (r *PlanPointResolver) P25() Money          { return money(r.point.P25, r.currency) }
func (r *PlanPointResolver) P50() Money          { return money(r.point.P50, r.currency) }
func (r *PlanPointResolver) P75() Money          { return money(r.point.P75, r.currency) }
func (r *PlanPointResolver) P90() Money          { return money(r.point.P90, r.currency) }

// Actual resolves Plan.actual: the net worth at start, at every month
// end since, and now.
func (r *PlanResolver) Actual(ctx context.Context) ([]*NetWorthPointResolver, error) {
	if _, _, _, err := r.loadNetWorth(ctx); err != nil {
		return nil, err
	}
	now := time.Now()
	if r.p.StartDate.After(now) {
		return []*NetWorthPointResolver{}, nil
	}
	times := []time.Time{r.p.StartDate}
	for m := 1; ; m++ {
		end := r.p.StartDate.AddDate(0, m, 0).Add(-time.Second)
		if !end.Before(now) {
			break
		}
		times = append(times, end)
	}
	times = append(times, now)
	points := finance.NetWorthAt(r.txs, r.estimated, times, r.quote)
	resolvers := make([]*NetWorthPointResolver, len(points))
	for i := range points {
		resolvers[i] = &NetWorthPointResolver{point: points[i], currency: r.currency}
	}
	return resolvers, nil
}

// Status resolves Plan.status: the current month's expected value against
// the net worth now.
func (r *PlanResolver) Status(ctx context.Context) (*PlanStatusResolver, error) {
	_, nowNW, currency, err := r.loadNetWorth(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if r.p.StartDate.After(now) {
		return nil, nil
	}
	points, err := r.project(ctx)
	if err != nil {
		return nil, err
	}
	month := monthStart(now)
	i := sort.Search(len(points), func(i int) bool { return !points[i].Month.Before(month) })
	if i >= len(points) {
		i = len(points) - 1
	}
	return &PlanStatusResolver{month: points[i].Month, expected: points[i].Expected, actual: nowNW, currency: currency}, nil
}

// PlanStatusResolver resolves PlanStatus.
type PlanStatusResolver struct {
	month            time.Time
	expected, actual int64
	currency         string
}

func (r *PlanStatusResolver) Month() graphql.Time { return graphql.Time{Time: r.month} }
func (r *PlanStatusResolver) Expected() Money     { return money(r.expected, r.currency) }
func (r *PlanStatusResolver) Actual() Money       { return money(r.actual, r.currency) }
func (r *PlanStatusResolver) Gap() Money          { return money(r.actual-r.expected, r.currency) }
