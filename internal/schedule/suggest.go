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
	// Extra counts occurrences beyond one per month — a bonus, a refund
	// from the same party — which the rhythm and the steps leave out.
	Extra int
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
		all := g.rows
		sort.Slice(all, func(i, j int) bool { return all[i].Time.Before(all[j].Time) })
		// One occurrence per month stands for the rhythm: the one closest
		// to the party's typical amount. A bonus or a refund in the same
		// month is an extra, not a second beat.
		rows := representatives(all)
		if len(rows) < minCount {
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
			Extra:      len(all) - len(rows),
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

// representatives keeps one occurrence per month: the one whose amount
// is closest to the median of all. Input and output are sorted by time.
func representatives(rows []Movement) []Movement {
	amounts := make([]int64, 0, len(rows))
	for _, r := range rows {
		a := r.Amount
		if a < 0 {
			a = -a
		}
		amounts = append(amounts, a)
	}
	sort.Slice(amounts, func(i, j int) bool { return amounts[i] < amounts[j] })
	typical := amounts[len(amounts)/2]
	distance := func(r Movement) int64 {
		a := r.Amount
		if a < 0 {
			a = -a
		}
		if a > typical {
			return a - typical
		}
		return typical - a
	}
	var out []Movement
	for _, r := range rows {
		month := r.Time.UTC().Format("2006-01")
		if n := len(out); n > 0 && out[n-1].Time.UTC().Format("2006-01") == month {
			if distance(r) < distance(out[n-1]) {
				out[n-1] = r
			}
			continue
		}
		out = append(out, r)
	}
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
	// A lone outlier between two runs of the same amount — a month whose
	// only payment was a bonus — does not end a step; the runs around it
	// are one.
	var merged []run
	for i, ru := range runs {
		avg := func(r run) int64 { return r.total / r.n }
		if n := len(merged); n > 0 {
			if ru.n == 1 && i < len(runs)-1 && same(avg(runs[i+1]), avg(merged[n-1])) {
				continue
			}
			if same(avg(ru), avg(merged[n-1])) {
				merged[n-1].total += ru.total
				merged[n-1].n += ru.n
				continue
			}
		}
		merged = append(merged, ru)
	}
	var steps []Step
	for i, ru := range merged {
		if ru.n >= 2 || i == len(merged)-1 && len(steps) > 0 {
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
