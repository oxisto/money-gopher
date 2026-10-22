package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/schedule"
)

// A schedule is a recurring cash flow the person expects: a salary, a
// Sparrate, a loan's monthly interest, the rent an office room brings in.
// It lands on one cash account; a TRANSFER is paid from a second one. Its
// occurrences are expectations for the cash-flow view, and, when the
// schedule books, real transactions once their date has passed. A
// statement that later lists the same movement confirms the booking
// instead of duplicating it (see ConfirmImport).

var scheduleTypes = map[string]bool{
	"DEPOSIT_CASH": true, "WITHDRAW_CASH": true, "INTEREST": true,
	"ACCOUNT_FEES": true, "DIVIDEND": true, "TRANSFER": true,
}

// scheduleSign is how an occurrence hits the schedule's cash account:
// money in is positive. Interest on a loan is charged, fees and
// withdrawals leave. A TRANSFER lands here and leaves its counterpart.
func scheduleSign(typ, accountKind string) int64 {
	switch typ {
	case "WITHDRAW_CASH", "ACCOUNT_FEES":
		return -1
	case "INTEREST":
		if accountKind == "LOAN" {
			return -1
		}
	}
	return 1
}

// ScheduleStepInput is one amount step of a schedule.
type ScheduleStepInput struct {
	From   graphql.Time
	Amount int32
}

// CreateScheduleInput is the input for Mutation.createSchedule.
type CreateScheduleInput struct {
	DisplayName          string
	Type                 string
	Interval             string
	DayOfMonth           int32
	Start                graphql.Time
	End                  *graphql.Time
	CashAccountID        graphql.ID
	CounterpartAccountID *graphql.ID
	AssetID              *graphql.ID
	Steps                []ScheduleStepInput
	Books                *bool
}

// UpdateScheduleInput is the input for Mutation.updateSchedule.
type UpdateScheduleInput struct {
	DisplayName          *string
	Interval             *string
	DayOfMonth           *int32
	Start                *graphql.Time
	End                  *graphql.Time
	ClearEnd             *bool
	CounterpartAccountID *graphql.ID
	Steps                *[]ScheduleStepInput
	Books                *bool
}

func stepsFromInput(in []ScheduleStepInput) []*persistence.ScheduleStep {
	out := make([]*persistence.ScheduleStep, 0, len(in))
	for _, s := range in {
		out = append(out, &persistence.ScheduleStep{FromDate: schedule.Midnight(s.From.Time), Amount: int64(s.Amount)})
	}
	return out
}

// validateSchedule checks a schedule and its steps against the person's
// accounts and assets.
func validateSchedule(ctx context.Context, db *persistence.DB, person *persistence.Person, s *persistence.Schedule, steps []*persistence.ScheduleStep) error {
	if s.DisplayName == "" {
		return errors.New("a schedule needs a name")
	}
	if !scheduleTypes[s.Type] {
		return fmt.Errorf("unknown schedule type %q", s.Type)
	}
	if !schedule.Recurrence(s.Recurrence).Valid() {
		return fmt.Errorf("unknown interval %q", s.Recurrence)
	}
	if s.DayOfMonth < 1 || s.DayOfMonth > 31 {
		return errors.New("the day of month must be between 1 and 31")
	}
	if s.EndDate.Valid && s.EndDate.Time.Before(s.StartDate) {
		return errors.New("the end lies before the start")
	}

	account, err := db.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: s.CashAccountID, PersonID: person.ID})
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("cash account %q not found", s.CashAccountID)
	} else if err != nil {
		return err
	}

	switch s.Type {
	case "TRANSFER":
		if !s.CounterpartAccountID.Valid {
			return errors.New("a transfer needs the account it is paid from")
		}
		if s.CounterpartAccountID.String == s.CashAccountID {
			return errors.New("a transfer needs two different accounts")
		}
		other, err := db.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: s.CounterpartAccountID.String, PersonID: person.ID})
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("cash account %q not found", s.CounterpartAccountID.String)
		} else if err != nil {
			return err
		}
		if other.Currency != account.Currency {
			return fmt.Errorf("%s is in %s, %s in %s: a transfer does not convert", other.DisplayName, other.Currency, account.DisplayName, account.Currency)
		}
	case "DIVIDEND":
		if !s.SecurityID.Valid {
			return errors.New("a dividend schedule needs the asset that pays it")
		}
		if _, err := db.GetSecurityVisibleTo(ctx, persistence.GetSecurityVisibleToParams{ID: s.SecurityID.String, PersonID: ownedBy(person)}); errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("asset %q not found", s.SecurityID.String)
		} else if err != nil {
			return err
		}
		if s.Books == 1 {
			return errors.New("a dividend schedule is an expectation only; its confirmations book it")
		}
	default:
		if s.CounterpartAccountID.Valid || s.SecurityID.Valid {
			return fmt.Errorf("a %s schedule takes neither a counterpart account nor an asset", s.Type)
		}
	}
	if s.Type != "TRANSFER" && s.CounterpartAccountID.Valid {
		return errors.New("only a transfer has a counterpart account")
	}

	if len(steps) == 0 {
		return errors.New("a schedule needs at least one amount")
	}
	seen := map[time.Time]bool{}
	for _, st := range steps {
		if st.Amount <= 0 {
			return errors.New("every amount must be positive")
		}
		if seen[st.FromDate] {
			return fmt.Errorf("two amounts from %s", st.FromDate.Format("2006-01-02"))
		}
		seen[st.FromDate] = true
	}
	return nil
}

func scheduleModel(s *persistence.Schedule, steps []*persistence.ScheduleStep) schedule.Schedule {
	m := schedule.Schedule{
		ID:         s.ID,
		Recurrence: schedule.Recurrence(s.Recurrence),
		DayOfMonth: int(s.DayOfMonth),
		Start:      s.StartDate,
	}
	if s.EndDate.Valid {
		end := s.EndDate.Time
		m.End = &end
	}
	for _, st := range steps {
		m.Steps = append(m.Steps, schedule.Step{From: st.FromDate, Amount: st.Amount})
	}
	return m
}

// ScheduleResolver resolves the Schedule GraphQL type.
type ScheduleResolver struct {
	db    *persistence.DB
	s     *persistence.Schedule
	steps []*persistence.ScheduleStep // loaded on first use
}

func (r *ScheduleResolver) loadSteps(ctx context.Context) ([]*persistence.ScheduleStep, error) {
	if r.steps == nil {
		steps, err := r.db.ListScheduleSteps(ctx, r.s.ID)
		if err != nil {
			return nil, err
		}
		r.steps = steps
	}
	return r.steps, nil
}

func (r *ScheduleResolver) model(ctx context.Context) (schedule.Schedule, error) {
	steps, err := r.loadSteps(ctx)
	if err != nil {
		return schedule.Schedule{}, err
	}
	return scheduleModel(r.s, steps), nil
}

func (r *ScheduleResolver) account(ctx context.Context) (*persistence.CashAccount, error) {
	return r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: r.s.CashAccountID, PersonID: r.s.PersonID})
}

func (r *ScheduleResolver) ID() graphql.ID      { return graphql.ID(r.s.ID) }
func (r *ScheduleResolver) DisplayName() string { return r.s.DisplayName }
func (r *ScheduleResolver) Type() string        { return r.s.Type }
func (r *ScheduleResolver) Interval() string    { return r.s.Recurrence }
func (r *ScheduleResolver) DayOfMonth() int32   { return int32(r.s.DayOfMonth) }
func (r *ScheduleResolver) Start() graphql.Time { return graphql.Time{Time: r.s.StartDate} }
func (r *ScheduleResolver) Books() bool         { return r.s.Books == 1 }
func (r *ScheduleResolver) End() *graphql.Time {
	if !r.s.EndDate.Valid {
		return nil
	}
	return &graphql.Time{Time: r.s.EndDate.Time}
}

func (r *ScheduleResolver) CashAccount(ctx context.Context) (*CashAccountResolver, error) {
	account, err := r.account(ctx)
	if err != nil {
		return nil, err
	}
	return &CashAccountResolver{db: r.db, account: account}, nil
}

func (r *ScheduleResolver) CounterpartAccount(ctx context.Context) (*CashAccountResolver, error) {
	if !r.s.CounterpartAccountID.Valid {
		return nil, nil
	}
	account, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: r.s.CounterpartAccountID.String, PersonID: r.s.PersonID})
	if err != nil {
		return nil, err
	}
	return &CashAccountResolver{db: r.db, account: account}, nil
}

func (r *ScheduleResolver) Asset(ctx context.Context) (*AssetResolver, error) {
	if !r.s.SecurityID.Valid {
		return nil, nil
	}
	sec, err := r.db.GetSecurityVisibleTo(ctx, persistence.GetSecurityVisibleToParams{ID: r.s.SecurityID.String, PersonID: sql.NullString{String: r.s.PersonID, Valid: true}})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &AssetResolver{db: r.db, security: sec}, nil
}

// ScheduleStepResolver resolves ScheduleStep.
type ScheduleStepResolver struct {
	step     *persistence.ScheduleStep
	currency string
}

func (r *ScheduleStepResolver) From() graphql.Time { return graphql.Time{Time: r.step.FromDate} }
func (r *ScheduleStepResolver) Amount() Money      { return money(r.step.Amount, r.currency) }

func (r *ScheduleResolver) Steps(ctx context.Context) ([]*ScheduleStepResolver, error) {
	steps, err := r.loadSteps(ctx)
	if err != nil {
		return nil, err
	}
	account, err := r.account(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*ScheduleStepResolver, 0, len(steps))
	for _, st := range steps {
		out = append(out, &ScheduleStepResolver{step: st, currency: account.Currency})
	}
	return out, nil
}

// CurrentAmount is the amount in force today, signed as it hits the
// schedule's cash account; zero before the first step.
func (r *ScheduleResolver) CurrentAmount(ctx context.Context) (Money, error) {
	m, err := r.model(ctx)
	if err != nil {
		return Money{}, err
	}
	account, err := r.account(ctx)
	if err != nil {
		return Money{}, err
	}
	amount, _ := m.AmountAt(time.Now())
	return money(scheduleSign(r.s.Type, kindOf(account))*amount, account.Currency), nil
}

func (r *ScheduleResolver) NextDate(ctx context.Context) (*graphql.Time, error) {
	m, err := r.model(ctx)
	if err != nil {
		return nil, err
	}
	next := m.Next(time.Now())
	if next == nil {
		return nil, nil
	}
	return &graphql.Time{Time: next.Date}, nil
}

// Booked counts the occurrences the schedule has written as transactions.
func (r *ScheduleResolver) Booked(ctx context.Context) (int32, error) {
	txs, err := r.db.ListTransactionsBySchedule(ctx, sql.NullString{String: r.s.ID, Valid: true})
	if err != nil {
		return 0, err
	}
	dates := map[time.Time]bool{}
	for _, tx := range txs {
		if tx.ScheduledFor.Valid {
			dates[tx.ScheduledFor.Time] = true
		}
	}
	return int32(len(dates)), nil
}

// Unconfirmed lists the schedule's bookings on its cash account that no
// statement has confirmed yet.
func (r *ScheduleResolver) Unconfirmed(ctx context.Context) ([]*TransactionResolver, error) {
	txs, err := r.db.ListTransactionsBySchedule(ctx, sql.NullString{String: r.s.ID, Valid: true})
	if err != nil {
		return nil, err
	}
	var out []*TransactionResolver
	for _, tx := range txs {
		if tx.DocumentID.Valid || !tx.CashAccountID.Valid || tx.CashAccountID.String != r.s.CashAccountID {
			continue
		}
		out = append(out, &TransactionResolver{db: r.db, transaction: tx})
	}
	return out, nil
}

// ExpectedFlowResolver resolves ExpectedFlow: one occurrence of a schedule.
type ExpectedFlowResolver struct {
	schedule *ScheduleResolver
	date     time.Time
	amount   Money
	booked   bool
}

func (r *ExpectedFlowResolver) Schedule() *ScheduleResolver { return r.schedule }
func (r *ExpectedFlowResolver) Date() graphql.Time          { return graphql.Time{Time: r.date} }
func (r *ExpectedFlowResolver) Amount() Money               { return r.amount }
func (r *ExpectedFlowResolver) Booked() bool                { return r.booked }

// Flows expands the schedule between from (inclusive) and to (exclusive),
// signed as it hits the schedule's cash account.
func (r *ScheduleResolver) Flows(ctx context.Context, args struct{ From, To graphql.Time }) ([]*ExpectedFlowResolver, error) {
	m, err := r.model(ctx)
	if err != nil {
		return nil, err
	}
	account, err := r.account(ctx)
	if err != nil {
		return nil, err
	}
	booked, err := r.db.ListTransactionsBySchedule(ctx, sql.NullString{String: r.s.ID, Valid: true})
	if err != nil {
		return nil, err
	}
	bookedFor := map[time.Time]bool{}
	for _, tx := range booked {
		if tx.ScheduledFor.Valid {
			bookedFor[tx.ScheduledFor.Time] = true
		}
	}
	sign := scheduleSign(r.s.Type, kindOf(account))
	var out []*ExpectedFlowResolver
	for _, occ := range m.Occurrences(args.From.Time, args.To.Time) {
		out = append(out, &ExpectedFlowResolver{
			schedule: r,
			date:     occ.Date,
			amount:   money(sign*occ.Amount, account.Currency),
			booked:   bookedFor[occ.Date],
		})
	}
	return out, nil
}

// Schedules resolves Query.schedules.
func (r *RootResolver) Schedules(ctx context.Context) ([]*ScheduleResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	schedules, err := r.db.ListSchedules(ctx, person.ID)
	if err != nil {
		return nil, err
	}
	out := make([]*ScheduleResolver, 0, len(schedules))
	for _, s := range schedules {
		out = append(out, &ScheduleResolver{db: r.db, s: s})
	}
	return out, nil
}

// Schedule resolves Query.schedule.
func (r *RootResolver) Schedule(ctx context.Context, args struct{ ID graphql.ID }) (*ScheduleResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	s, err := r.db.GetSchedule(ctx, persistence.GetScheduleParams{ID: string(args.ID), PersonID: person.ID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &ScheduleResolver{db: r.db, s: s}, nil
}

// Schedules resolves CashAccount.schedules: those landing on the account
// and those paid from it.
func (r *CashAccountResolver) Schedules(ctx context.Context) ([]*ScheduleResolver, error) {
	schedules, err := r.db.ListSchedulesByCashAccount(ctx, persistence.ListSchedulesByCashAccountParams{PersonID: r.account.PersonID, CashAccountID: r.account.ID, CounterpartAccountID: sql.NullString{String: r.account.ID, Valid: true}})
	if err != nil {
		return nil, err
	}
	out := make([]*ScheduleResolver, 0, len(schedules))
	for _, s := range schedules {
		out = append(out, &ScheduleResolver{db: r.db, s: s})
	}
	return out, nil
}

// ExpectedFlows resolves Query.expectedFlows: every schedule's occurrences
// in the window, by date then name.
func (r *RootResolver) ExpectedFlows(ctx context.Context, args struct{ From, To graphql.Time }) ([]*ExpectedFlowResolver, error) {
	schedules, err := r.Schedules(ctx)
	if err != nil {
		return nil, err
	}
	var out []*ExpectedFlowResolver
	for _, s := range schedules {
		flows, err := s.Flows(ctx, args)
		if err != nil {
			return nil, err
		}
		out = append(out, flows...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].date.Equal(out[j].date) {
			return out[i].date.Before(out[j].date)
		}
		return out[i].schedule.s.DisplayName < out[j].schedule.s.DisplayName
	})
	return out, nil
}

// CreateSchedule resolves Mutation.createSchedule. A schedule that books
// writes its past occurrences right away.
func (r *RootResolver) CreateSchedule(ctx context.Context, args struct{ Input CreateScheduleInput }) (*ScheduleResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	in := args.Input
	s := &persistence.Schedule{
		ID:                   uuid.NewString(),
		PersonID:             person.ID,
		DisplayName:          in.DisplayName,
		Type:                 in.Type,
		CashAccountID:        string(in.CashAccountID),
		CounterpartAccountID: nullString(idPtr(in.CounterpartAccountID)),
		SecurityID:           nullString(idPtr(in.AssetID)),
		Recurrence:           in.Interval,
		DayOfMonth:           int64(in.DayOfMonth),
		StartDate:            schedule.Midnight(in.Start.Time),
	}
	if in.End != nil {
		s.EndDate = sql.NullTime{Time: schedule.Midnight(in.End.Time), Valid: true}
	}
	if in.Books != nil && *in.Books {
		s.Books = 1
	}
	steps := stepsFromInput(in.Steps)
	if err := validateSchedule(ctx, r.db, person, s, steps); err != nil {
		return nil, err
	}

	var created *persistence.Schedule
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		var err error
		created, err = q.CreateSchedule(ctx, persistence.CreateScheduleParams{
			ID: s.ID, PersonID: s.PersonID, DisplayName: s.DisplayName, Type: s.Type,
			CashAccountID: s.CashAccountID, CounterpartAccountID: s.CounterpartAccountID, SecurityID: s.SecurityID,
			Recurrence: s.Recurrence, DayOfMonth: s.DayOfMonth, StartDate: s.StartDate, EndDate: s.EndDate,
			Books: s.Books, CreatedAt: time.Now().UTC(),
		})
		if err != nil {
			return err
		}
		for _, st := range steps {
			if err := q.CreateScheduleStep(ctx, persistence.CreateScheduleStepParams{ScheduleID: s.ID, FromDate: st.FromDate, Amount: st.Amount}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if created.Books == 1 {
		if _, err := bookSchedule(ctx, r.db, created, time.Now()); err != nil {
			return nil, err
		}
	}
	return &ScheduleResolver{db: r.db, s: created}, nil
}

// UpdateSchedule resolves Mutation.updateSchedule. Steps, when given,
// replace the existing ones. Type, cash account and asset are fixed:
// those make it a different schedule.
func (r *RootResolver) UpdateSchedule(ctx context.Context, args struct {
	ID    graphql.ID
	Input UpdateScheduleInput
}) (*ScheduleResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	current, err := r.db.GetSchedule(ctx, persistence.GetScheduleParams{ID: string(args.ID), PersonID: person.ID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("schedule %q not found", args.ID)
	} else if err != nil {
		return nil, err
	}
	in := args.Input
	s := *current
	if in.DisplayName != nil {
		s.DisplayName = *in.DisplayName
	}
	if in.Interval != nil {
		s.Recurrence = *in.Interval
	}
	if in.DayOfMonth != nil {
		s.DayOfMonth = int64(*in.DayOfMonth)
	}
	if in.Start != nil {
		s.StartDate = schedule.Midnight(in.Start.Time)
	}
	if in.End != nil {
		s.EndDate = sql.NullTime{Time: schedule.Midnight(in.End.Time), Valid: true}
	}
	if in.ClearEnd != nil && *in.ClearEnd {
		s.EndDate = sql.NullTime{}
	}
	if in.CounterpartAccountID != nil {
		s.CounterpartAccountID = nullString(idPtr(in.CounterpartAccountID))
	}
	if in.Books != nil {
		s.Books = 0
		if *in.Books {
			s.Books = 1
		}
	}
	var steps []*persistence.ScheduleStep
	if in.Steps != nil {
		steps = stepsFromInput(*in.Steps)
	} else if steps, err = r.db.ListScheduleSteps(ctx, s.ID); err != nil {
		return nil, err
	}
	if err := validateSchedule(ctx, r.db, person, &s, steps); err != nil {
		return nil, err
	}

	var updated *persistence.Schedule
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		var err error
		updated, err = q.UpdateSchedule(ctx, persistence.UpdateScheduleParams{
			DisplayName: s.DisplayName, CounterpartAccountID: s.CounterpartAccountID, Recurrence: s.Recurrence,
			DayOfMonth: s.DayOfMonth, StartDate: s.StartDate, EndDate: s.EndDate, Books: s.Books,
			ID: s.ID, PersonID: person.ID,
		})
		if err != nil {
			return err
		}
		if in.Steps != nil {
			if err := q.DeleteScheduleSteps(ctx, s.ID); err != nil {
				return err
			}
			for _, st := range steps {
				if err := q.CreateScheduleStep(ctx, persistence.CreateScheduleStepParams{ScheduleID: s.ID, FromDate: st.FromDate, Amount: st.Amount}); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if updated.Books == 1 {
		if _, err := bookSchedule(ctx, r.db, updated, time.Now()); err != nil {
			return nil, err
		}
	}
	return &ScheduleResolver{db: r.db, s: updated}, nil
}

// DeleteSchedule resolves Mutation.deleteSchedule. Transactions it booked
// stay, as plain transactions.
func (r *RootResolver) DeleteSchedule(ctx context.Context, args struct{ ID graphql.ID }) (graphql.ID, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return "", err
	}
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		if err := q.UnlinkScheduleTransactions(ctx, sql.NullString{String: string(args.ID), Valid: true}); err != nil {
			return err
		}
		rows, err := q.DeleteSchedule(ctx, persistence.DeleteScheduleParams{ID: string(args.ID), PersonID: person.ID})
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("schedule %q not found", args.ID)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return args.ID, nil
}

// BookSchedules resolves Mutation.bookSchedules: writes every due
// occurrence of the person's booking schedules, up to until (default now),
// and returns how many were written.
func (r *RootResolver) BookSchedules(ctx context.Context, args struct{ Until *graphql.Time }) (int32, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return 0, err
	}
	until := time.Now()
	if args.Until != nil {
		until = args.Until.Time
	}
	schedules, err := r.db.ListSchedules(ctx, person.ID)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, s := range schedules {
		if s.Books != 1 {
			continue
		}
		n, err := bookSchedule(ctx, r.db, s, until)
		if err != nil {
			return int32(total), err
		}
		total += n
	}
	return int32(total), nil
}

// bookSchedule writes the schedule's occurrences up to and including
// until's day that have no transaction yet. Idempotent: an occurrence is
// identified by (schedule, date). Returns how many were written.
func bookSchedule(ctx context.Context, db *persistence.DB, s *persistence.Schedule, until time.Time) (int, error) {
	if s.Type == "DIVIDEND" {
		return 0, nil
	}
	steps, err := db.ListScheduleSteps(ctx, s.ID)
	if err != nil {
		return 0, err
	}
	person := &persistence.Person{ID: s.PersonID}
	account, err := db.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: s.CashAccountID, PersonID: s.PersonID})
	if err != nil {
		return 0, err
	}
	sign := scheduleSign(s.Type, kindOf(account))
	scheduleID := sql.NullString{String: s.ID, Valid: true}

	n := 0
	for _, occ := range scheduleModel(s, steps).Occurrences(s.StartDate, schedule.Midnight(until).AddDate(0, 0, 1)) {
		existing, err := db.ListTransactionsByScheduleOccurrence(ctx, persistence.ListTransactionsByScheduleOccurrenceParams{ScheduleID: scheduleID, ScheduledFor: sql.NullTime{Time: occ.Date, Valid: true}})
		if err != nil {
			return n, err
		}
		if len(existing) > 0 {
			continue
		}
		// A transaction already in the books for this occurrence — imported
		// from a statement or entered by hand before the schedule existed —
		// becomes the schedule's own instead of being booked a second time.
		if adopted, err := adoptOccurrence(ctx, db, s, account, occ.Date, sign*occ.Amount, scheduleID); err != nil {
			return n, err
		} else if adopted {
			n++
			continue
		}

		var legs []txData
		var transferID sql.NullString
		if s.Type == "TRANSFER" {
			from := s.CounterpartAccountID.String
			to := s.CashAccountID
			legs = []txData{
				{Type: "WITHDRAW_CASH", Time: occ.Date, CashAccountID: &from, CashDelta: -occ.Amount, Currency: account.Currency},
				{Type: "DEPOSIT_CASH", Time: occ.Date, CashAccountID: &to, CashDelta: occ.Amount, Currency: account.Currency},
			}
			transferID = sql.NullString{String: uuid.NewString(), Valid: true}
		} else {
			to := s.CashAccountID
			legs = []txData{{Type: s.Type, Time: occ.Date, CashAccountID: &to, CashDelta: sign * occ.Amount, Currency: account.Currency}}
		}
		for i := range legs {
			if err := legs[i].validate(ctx, db, person); err != nil {
				return n, fmt.Errorf("schedule %q on %s: %w", s.DisplayName, occ.Date.Format("2006-01-02"), err)
			}
		}
		err = db.Tx(ctx, func(q *persistence.Queries) error {
			for _, d := range legs {
				if _, err := q.CreateTransaction(ctx, persistence.CreateTransactionParams{
					ID:            uuid.NewString(),
					Type:          d.Type,
					Time:          d.Time,
					CashAccountID: nullString(d.CashAccountID),
					CashDelta:     d.CashDelta,
					Currency:      d.Currency,
					Source:        "schedule",
					TransferID:    transferID,
					ScheduleID:    scheduleID,
					ScheduledFor:  sql.NullTime{Time: occ.Date, Valid: true},
				}); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// adoptOccurrence links an existing transaction on the schedule's account
// (same type and amount, within a week of the occurrence, not yet any
// schedule's) to the occurrence, both legs when it is a transfer.
func adoptOccurrence(ctx context.Context, db *persistence.DB, s *persistence.Schedule, account *persistence.CashAccount, date time.Time, cashDelta int64, scheduleID sql.NullString) (bool, error) {
	typ := s.Type
	if typ == "TRANSFER" {
		typ = "DEPOSIT_CASH"
	}
	candidates, err := db.ListAdoptableTransactions(ctx, persistence.ListAdoptableTransactionsParams{
		CashAccountID: sql.NullString{String: account.ID, Valid: true},
		Type:          typ,
		CashDelta:     cashDelta,
		Time:          date.AddDate(0, 0, -7),
		Time_2:        date.AddDate(0, 0, 7),
	})
	if err != nil || len(candidates) == 0 {
		return false, err
	}
	tx := candidates[0]
	for _, c := range candidates { // the closest one
		if abs(c.Time.Sub(date)) < abs(tx.Time.Sub(date)) {
			tx = c
		}
	}
	return true, db.Tx(ctx, func(q *persistence.Queries) error {
		for _, id := range []string{tx.ID} {
			if err := q.LinkTransactionToSchedule(ctx, persistence.LinkTransactionToScheduleParams{ScheduleID: scheduleID, ScheduledFor: sql.NullTime{Time: date, Valid: true}, ID: id}); err != nil {
				return err
			}
		}
		if tx.TransferID.Valid {
			other, err := q.GetTransferCounterpart(ctx, persistence.GetTransferCounterpartParams{TransferID: tx.TransferID, ID: tx.ID})
			if err == nil {
				return q.LinkTransactionToSchedule(ctx, persistence.LinkTransactionToScheduleParams{ScheduleID: scheduleID, ScheduledFor: sql.NullTime{Time: date, Valid: true}, ID: other.ID})
			} else if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}
		return nil
	})
}

func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

// Bookkeeper writes due occurrences of every person's booking schedules,
// at startup and then every interval, so a Sparrate is in the books on the
// day it is pulled without anyone opening the app.
type Bookkeeper struct {
	DB *persistence.DB
}

// BookAll runs one pass over all booking schedules.
func (b *Bookkeeper) BookAll(ctx context.Context) (int, error) {
	schedules, err := b.DB.ListBookingSchedules(ctx)
	if err != nil {
		return 0, err
	}
	total := 0
	var errs error
	for _, s := range schedules {
		n, err := bookSchedule(ctx, b.DB, s, time.Now())
		total += n
		errs = errors.Join(errs, err)
	}
	return total, errs
}

// Run books now and then every interval until ctx ends.
func (b *Bookkeeper) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if n, err := b.BookAll(ctx); err != nil {
			slog.Error("booking schedules failed", "err", err, "booked", n)
		} else if n > 0 {
			slog.Info("booked scheduled transactions", "transactions", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// ScheduleID resolves Transaction.scheduleID.
func (r *TransactionResolver) ScheduleID() *graphql.ID {
	if !r.transaction.ScheduleID.Valid {
		return nil
	}
	id := graphql.ID(r.transaction.ScheduleID.String)
	return &id
}

// Schedule resolves Transaction.schedule.
func (r *TransactionResolver) Schedule(ctx context.Context) (*ScheduleResolver, error) {
	if !r.transaction.ScheduleID.Valid {
		return nil, nil
	}
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	s, err := r.db.GetSchedule(ctx, persistence.GetScheduleParams{ID: r.transaction.ScheduleID.String, PersonID: person.ID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &ScheduleResolver{db: r.db, s: s}, nil
}

// ScheduledFor resolves Transaction.scheduledFor.
func (r *TransactionResolver) ScheduledFor() *graphql.Time {
	if !r.transaction.ScheduledFor.Valid {
		return nil
	}
	return &graphql.Time{Time: r.transaction.ScheduledFor.Time}
}
