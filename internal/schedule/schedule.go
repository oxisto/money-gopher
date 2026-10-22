// Package schedule expands recurring cash flows — a salary, a Sparrate, a
// loan's monthly interest, the rent an office room brings in — into dated
// occurrences. It knows nothing about accounts or signs: the API decides
// what an occurrence means for which account; this package only answers
// "when, and how much".
package schedule

import (
	"sort"
	"time"
)

// Recurrence is how often a schedule occurs.
type Recurrence string

const (
	Monthly   Recurrence = "MONTHLY"
	Quarterly Recurrence = "QUARTERLY"
	Yearly    Recurrence = "YEARLY"
)

// Months is the gap between two occurrences.
func (r Recurrence) Months() int {
	switch r {
	case Quarterly:
		return 3
	case Yearly:
		return 12
	default:
		return 1
	}
}

// Valid reports whether r is one of the known recurrences.
func (r Recurrence) Valid() bool {
	return r == Monthly || r == Quarterly || r == Yearly
}

// A Step is the amount that applies from a date on: a raise, a changed
// Sparrate. Amounts are unsigned minor units.
type Step struct {
	From   time.Time
	Amount int64
}

// A Schedule occurs on DayOfMonth every Recurrence, anchored at Start's
// month, from Start to End inclusive (End nil: forever).
type Schedule struct {
	ID         string
	Recurrence Recurrence
	DayOfMonth int
	Start      time.Time
	End        *time.Time
	Steps      []Step
}

// An Occurrence is one dated instance of a schedule.
type Occurrence struct {
	ScheduleID string
	Date       time.Time
	Amount     int64
}

// Day is midnight UTC on the given day, clamped to the month's length: the
// 31st of April is the 30th, the 30th of February the 28th or 29th.
func Day(year int, month time.Month, day int) time.Time {
	last := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > last {
		day = last
	}
	if day < 1 {
		day = 1
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// Midnight truncates t to its UTC day.
func Midnight(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// AmountAt is the step in force on t: the last step whose From is on or
// before t. False when t lies before the first step.
func (s Schedule) AmountAt(t time.Time) (int64, bool) {
	steps := s.sortedSteps()
	var amount int64
	found := false
	for _, st := range steps {
		if st.From.After(t) {
			break
		}
		amount, found = st.Amount, true
	}
	return amount, found
}

func (s Schedule) sortedSteps() []Step {
	steps := append([]Step(nil), s.Steps...)
	sort.Slice(steps, func(i, j int) bool { return steps[i].From.Before(steps[j].From) })
	return steps
}

// Occurrences lists the schedule's occurrences with from <= Date < to,
// within [Start, End]. Occurrences before the first step carry no amount
// and are left out.
func (s Schedule) Occurrences(from, to time.Time) []Occurrence {
	var out []Occurrence
	start := Midnight(s.Start)
	step := s.Recurrence.Months()
	for n := 0; ; n++ {
		// n-th month after the anchor; Go normalises month overflow.
		anchor := time.Date(start.Year(), start.Month()+time.Month(n*step), 1, 0, 0, 0, 0, time.UTC)
		date := Day(anchor.Year(), anchor.Month(), s.DayOfMonth)
		if !date.Before(to) {
			break
		}
		if s.End != nil && date.After(Midnight(*s.End)) {
			break
		}
		if date.Before(start) || date.Before(from) {
			continue
		}
		amount, ok := s.AmountAt(date)
		if !ok {
			continue
		}
		out = append(out, Occurrence{ScheduleID: s.ID, Date: date, Amount: amount})
	}
	return out
}

// Next is the first occurrence strictly after t, or nil when the schedule
// has ended.
func (s Schedule) Next(t time.Time) *Occurrence {
	from := Midnight(t).AddDate(0, 0, 1)
	// Two full cycles cover any clamping; the first hit wins.
	to := from.AddDate(0, 2*s.Recurrence.Months()+1, 0)
	occ := s.Occurrences(from, to)
	if len(occ) == 0 {
		return nil
	}
	return &occ[0]
}
