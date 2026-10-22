package persistence

import (
	"context"
	"database/sql"
	"testing"
)

func TestNormalizeIBAN(t *testing.T) {
	for in, want := range map[string]string{
		"DE46 5001 0517 5802 2452 63": "DE46500105175802245263",
		" de46500105175802245263 ":    "DE46500105175802245263",
		"":                            "",
	} {
		if got := NormalizeIBAN(in); got != want {
			t.Errorf("NormalizeIBAN(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestNormalizeIBANs: accounts stored with spaces, from before IBANs were
// normalized on the way in, are rewritten once; the rest are left alone.
func TestNormalizeIBANs(t *testing.T) {
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	if _, err := db.CreatePerson(ctx, CreatePersonParams{ID: "p", DisplayName: "P"}); err != nil {
		t.Fatal(err)
	}
	for id, iban := range map[string]string{"spaced": "DE46 5001 0517 5802 2452 63", "clean": "DE02123456780000000042"} {
		if _, err := db.CreateCashAccount(ctx, CreateCashAccountParams{
			ID: id, PersonID: "p", DisplayName: id, Currency: "EUR", Iban: sql.NullString{String: iban, Valid: true},
		}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.CreateCashAccount(ctx, CreateCashAccountParams{ID: "none", PersonID: "p", DisplayName: "none", Currency: "EUR"}); err != nil {
		t.Fatal(err)
	}

	if n, err := db.NormalizeIBANs(ctx); err != nil || n != 1 {
		t.Fatalf("NormalizeIBANs() = %d, %v; want 1, nil", n, err)
	}
	got, err := db.GetCashAccountByIBAN(ctx, GetCashAccountByIBANParams{Iban: sql.NullString{String: "DE46500105175802245263", Valid: true}, PersonID: "p"})
	if err != nil || got.ID != "spaced" {
		t.Errorf("lookup by the normalized IBAN = %v, %v; want the spaced account", got, err)
	}
	if n, err := db.NormalizeIBANs(ctx); err != nil || n != 0 {
		t.Errorf("second NormalizeIBANs() = %d, %v; want 0, nil", n, err)
	}
}
