package schedule

import (
	"regexp"
	"sort"
	"strings"
	"time"
)

// Suggestions: what recurs on an account, read off its transactions. A
// salary arriving on the 28th, a card bill on the 3rd, an insurance once
// a year. Groups form by counterparty, qualify by rhythm, and their
// amounts split into steps where they change — a raise, a new Sparrate.
// The result is a proposal for a schedule, never a schedule itself.

// A Movement is one transaction as the detector sees it.
type Movement struct {
	Time   time.Time
	Amount int64 // signed minor units
	// Note is the statement's wording: "Buchungsart counterparty · purpose".
	Note string
}

// A Suggestion is one recurring movement the detector found.
type Suggestion struct {
	// Name is the counterparty as the statement prints it.
	Name string
	// Incoming is true for money arriving.
	Incoming   bool
	Recurrence Recurrence
	DayOfMonth int
	Start      time.Time
	LastSeen   time.Time
	Count      int
	// Steps are the amounts in force from a date on, unsigned.
	Steps []Step
	// Variable is true when the amount changes from one occurrence to
	// the next, as a card bill does; Steps then holds one average.
	Variable bool
}

var (
	suggestDigits = regexp.MustCompile(`\d+`)
	suggestSpaces = regexp.MustCompile(`\s+`)
)

// counterparty reduces a note to the party it names: the part before
// " · " without the leading Buchungsart, lower case, digits dropped.
func counterparty(note string) (key, display string) {
	head, _, _ := strings.Cut(note, " · ")
	head = strings.TrimSpace(head)
	if head == "" {
		return "", ""
	}
	if _, rest, ok := strings.Cut(head, " "); ok {
		display = strings.TrimSpace(rest)
	}
	if display == "" {
		display = head
	}
	key = strings.ToLower(suggestDigits.ReplaceAllString(display, ""))
	key = strings.TrimSpace(suggestSpaces.ReplaceAllString(key, " "))
	return key, display
}

// Suggest finds the recurring movements among ms. A group needs at least
// minCount occurrences in as many different months, with a typical gap
// between neighbours that fits a monthly, quarterly or yearly rhythm.
func Suggest(ms []Movement, minCount int) []Suggestion {
	if minCount < 2 {
		minCount = 2
	}
	type group struct {
		display string
		rows    []Movement
	}
	groups := map[string]*group{}
	for _, m := range ms {
		key, display := counterparty(m.Note)
		if key == "" || m.Amount == 0 {
			continue
		}
		// Money in and money out from the same party are two things.
		if m.Amount < 0 {
			key = "-" + key
		}
		g := groups[key]
		if g == nil {
			g = &group{display: display}
			groups[key] = g
		}
		g.rows = append(g.rows, m)
	}

	var out []Suggestion
	for _, g := range groups {
		rows := g.rows
		sort.Slice(rows, func(i, j int) bool { return rows[i].Time.Before(rows[j].Time) })
		months := map[string]bool{}
		for _, r := range rows {
			months[r.Time.Format("2006-01")] = true
		}
		if len(rows) < minCount || len(months) < minCount {
			continue
		}
		recurrence, ok := rhythm(rows)
		if !ok {
			continue
		}
		steps, variable := amountSteps(rows)
		if len(steps) == 0 {
			continue
		}
		out = append(out, Suggestion{
			Name:       g.display,
			Incoming:   rows[0].Amount > 0,
			Recurrence: recurrence,
			DayOfMonth: medianDay(rows),
			Start:      Midnight(rows[0].Time),
			LastSeen:   Midnight(rows[len(rows)-1].Time),
			Count:      len(rows),
			Steps:      steps,
			Variable:   variable,
		})
	}
	// Largest first: the salary before the streaming subscription.
	sort.Slice(out, func(i, j int) bool {
		ai, aj := out[i].Steps[len(out[i].Steps)-1].Amount, out[j].Steps[len(out[j].Steps)-1].Amount
		if ai != aj {
			return ai > aj
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// rhythm reads the recurrence off the typical gap between neighbours.
// Statements may be missing for whole months, so the median gap is
// used, not every gap.
func rhythm(rows []Movement) (Recurrence, bool) {
	var gaps []float64
	for i := 1; i < len(rows); i++ {
		gaps = append(gaps, rows[i].Time.Sub(rows[i-1].Time).Hours()/24)
	}
	if len(gaps) == 0 {
		return "", false
	}
	sort.Float64s(gaps)
	median := gaps[len(gaps)/2]
	switch {
	case median >= 25 && median <= 36:
		return Monthly, true
	case median >= 80 && median <= 100:
		return Quarterly, true
	case median >= 350 && median <= 380:
		return Yearly, true
	}
	return "", false
}

func medianDay(rows []Movement) int {
	days := make([]int, 0, len(rows))
	for _, r := range rows {
		days = append(days, r.Time.UTC().Day())
	}
	sort.Ints(days)
	return days[len(days)/2]
}

// amountSteps splits the amounts into runs of the same value (within
// 3 %) and makes a step of each run of two or more, or of a final run
// of one — the newest amount, seen once so far. When no run reaches two,
// the amount varies and one average stands for all.
func amountSteps(rows []Movement) ([]Step, bool) {
	abs := func(a int64) int64 {
		if a < 0 {
			return -a
		}
		return a
	}
	same := func(a, b int64) bool {
		d := a - b
		if d < 0 {
			d = -d
		}
		return d*100 <= 3*b
	}
	type run struct {
		from  time.Time
		total int64
		n     int64
	}
	var runs []run
	for _, r := range rows {
		a := abs(r.Amount)
		if len(runs) > 0 {
			last := &runs[len(runs)-1]
			if same(a, last.total/last.n) {
				last.total += a
				last.n++
				continue
			}
		}
		runs = append(runs, run{from: Midnight(r.Time), total: a, n: 1})
	}
	var steps []Step
	for i, ru := range runs {
		if ru.n >= 2 || i == len(runs)-1 && len(steps) > 0 {
			steps = append(steps, Step{From: ru.from, Amount: ru.total / ru.n})
		}
	}
	if len(steps) > 0 {
		return steps, false
	}
	// Nothing stable: an average, from the first occurrence.
	var total int64
	for _, r := range rows {
		total += abs(r.Amount)
	}
	return []Step{{From: Midnight(rows[0].Time), Amount: total / int64(len(rows))}}, true
}
