package schedule

import (
	"testing"
	"time"
)

func TestSuggest(t *testing.T) {
	var ms []Movement
	// A salary on the 28th, with a raise from April 2026; one month's
	// statement is missing. Amounts drift by a few cents.
	for m := 1; m <= 12; m++ {
		if m == 7 {
			continue
		}
		amount := int64(485012)
		if m >= 4 {
			amount = 505000
		}
		ms = append(ms, Movement{Time: d(2026, time.Month(m), 28), Amount: amount + int64(m%3), Note: "Gehalt/Rente Fraunhofer-Gesellschaft zur Forderung der angewandten Forschung einge · 3440 Lohn/Gehalt 10010826/2026" + string(rune('0'+m))})
	}
	// A card bill around the 3rd, a different amount every month.
	for m := 1; m <= 12; m++ {
		ms = append(ms, Movement{Time: d(2026, time.Month(m), 3), Amount: -int64(80000 + 7000*m), Note: "Lastschrift VISA Card · Kreditkartenabrechnung 10/2026"})
	}
	// An insurance once a year, three years running.
	for y := 2024; y <= 2026; y++ {
		ms = append(ms, Movement{Time: d(y, 1, 15), Amount: -45000, Note: "Lastschrift Hausrat Versicherung AG · Police 123"})
	}
	// Two one-offs and an unlabelled row: not recurring.
	ms = append(ms, Movement{Time: d(2026, 5, 2), Amount: -129900, Note: "Ueberweisung Fahrradladen · Rechnung 77"})
	ms = append(ms, Movement{Time: d(2026, 9, 2), Amount: -129900, Note: "Ueberweisung Fahrradladen · Rechnung 99"})
	ms = append(ms, Movement{Time: d(2026, 6, 1), Amount: 100000})

	got := Suggest(ms, 3)
	if len(got) != 3 {
		for _, s := range got {
			t.Logf("%+v", s)
		}
		t.Fatalf("got %d suggestions, want 3 (salary, card, insurance)", len(got))
	}

	salary := got[0]
	if !salary.Incoming || salary.Recurrence != Monthly || salary.DayOfMonth != 28 || salary.Count != 11 || salary.Variable {
		t.Errorf("salary = %+v", salary)
	}
	if salary.Name != "Fraunhofer-Gesellschaft zur Forderung der angewandten Forschung einge" {
		t.Errorf("salary name = %q", salary.Name)
	}
	if len(salary.Steps) != 2 || salary.Steps[0].From != d(2026, 1, 28) || salary.Steps[1].From != d(2026, 4, 28) || salary.Steps[1].Amount < 504990 || salary.Steps[1].Amount > 505010 {
		t.Errorf("salary steps = %+v, want 4,850 from January and 5,050 from April", salary.Steps)
	}

	card := got[1]
	if card.Incoming || card.Recurrence != Monthly || card.DayOfMonth != 3 || !card.Variable || len(card.Steps) != 1 {
		t.Errorf("card = %+v, want a variable monthly withdrawal with one average", card)
	}
	if card.Steps[0].Amount != 125500 {
		t.Errorf("card average = %d, want 125500", card.Steps[0].Amount)
	}

	insurance := got[2]
	if insurance.Recurrence != Yearly || insurance.DayOfMonth != 15 || insurance.Steps[0].Amount != 45000 || insurance.Name != "Hausrat Versicherung AG" {
		t.Errorf("insurance = %+v", insurance)
	}
}
