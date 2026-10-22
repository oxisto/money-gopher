package schedule

import (
	"testing"
	"time"
)

func d(y int, m time.Month, day int) time.Time { return time.Date(y, m, day, 0, 0, 0, 0, time.UTC) }

func dates(occ []Occurrence) []string {
	var out []string
	for _, o := range occ {
		out = append(out, o.Date.Format("2006-01-02"))
	}
	return out
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestOccurrences(t *testing.T) {
	sparrate := Schedule{ID: "s", Recurrence: Monthly, DayOfMonth: 1, Start: d(2026, 1, 1), Steps: []Step{{d(2026, 1, 1), 53582}}}

	t.Run("monthly within a half-open window", func(t *testing.T) {
		got := dates(sparrate.Occurrences(d(2026, 1, 1), d(2026, 4, 1)))
		if want := []string{"2026-01-01", "2026-02-01", "2026-03-01"}; !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("a start after the day skips that month", func(t *testing.T) {
		s := sparrate
		s.Start = d(2026, 1, 15)
		got := dates(s.Occurrences(d(2026, 1, 1), d(2026, 3, 15)))
		if want := []string{"2026-02-01", "2026-03-01"}; !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("the end is inclusive", func(t *testing.T) {
		s := sparrate
		end := d(2026, 3, 1)
		s.End = &end
		got := dates(s.Occurrences(d(2026, 1, 1), d(2027, 1, 1)))
		if want := []string{"2026-01-01", "2026-02-01", "2026-03-01"}; !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("the 31st clamps to the month's end, leap year included", func(t *testing.T) {
		s := Schedule{Recurrence: Monthly, DayOfMonth: 31, Start: d(2028, 1, 1), Steps: []Step{{d(2028, 1, 1), 1}}}
		got := dates(s.Occurrences(d(2028, 1, 1), d(2028, 5, 1)))
		if want := []string{"2028-01-31", "2028-02-29", "2028-03-31", "2028-04-30"}; !equal(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("quarterly and yearly anchor at the start's month", func(t *testing.T) {
		q := Schedule{Recurrence: Quarterly, DayOfMonth: 15, Start: d(2026, 2, 1), Steps: []Step{{d(2026, 1, 1), 1}}}
		got := dates(q.Occurrences(d(2026, 1, 1), d(2027, 1, 1)))
		if want := []string{"2026-02-15", "2026-05-15", "2026-08-15", "2026-11-15"}; !equal(got, want) {
			t.Errorf("quarterly: got %v, want %v", got, want)
		}
		y := Schedule{Recurrence: Yearly, DayOfMonth: 2, Start: d(2026, 1, 2), Steps: []Step{{d(2026, 1, 1), 1}}}
		got = dates(y.Occurrences(d(2026, 1, 1), d(2029, 1, 1)))
		if want := []string{"2026-01-02", "2027-01-02", "2028-01-02"}; !equal(got, want) {
			t.Errorf("yearly: got %v, want %v", got, want)
		}
	})
	t.Run("steps change the amount and gate the start", func(t *testing.T) {
		salary := Schedule{Recurrence: Monthly, DayOfMonth: 28, Start: d(2025, 1, 1), Steps: []Step{
			{d(2026, 4, 1), 400000}, // the raise is listed first: order must not matter
			{d(2025, 6, 1), 350000},
		}}
		occ := salary.Occurrences(d(2025, 1, 1), d(2026, 6, 1))
		if len(occ) != 12 || occ[0].Date != d(2025, 6, 28) {
			t.Fatalf("got %d occurrences from %v, want 12 from 2025-06-28 (nothing before the first step)", len(occ), dates(occ))
		}
		if occ[9].Amount != 350000 || occ[10].Amount != 400000 {
			t.Errorf("March = %d, April = %d; want 350000 then 400000", occ[9].Amount, occ[10].Amount)
		}
	})
}

func TestNext(t *testing.T) {
	s := Schedule{Recurrence: Monthly, DayOfMonth: 1, Start: d(2026, 1, 1), Steps: []Step{{d(2026, 1, 1), 1}}}
	if n := s.Next(d(2026, 10, 22)); n == nil || n.Date != d(2026, 11, 1) {
		t.Errorf("Next(2026-10-22) = %v, want 2026-11-01", n)
	}
	if n := s.Next(d(2026, 11, 1)); n == nil || n.Date != d(2026, 12, 1) {
		t.Errorf("Next(on the day) = %v, want 2026-12-01", n)
	}
	end := d(2026, 11, 1)
	s.End = &end
	if n := s.Next(d(2026, 11, 1)); n != nil {
		t.Errorf("Next after the end = %v, want nil", n)
	}
}
