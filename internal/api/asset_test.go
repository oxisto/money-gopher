package api

import (
	"context"
	"database/sql"
	"testing"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// TestImportMatchesVisibleISIN: staging matches a document's ISIN against
// the shared securities and the person's own assets, never against another
// person's.
func TestImportMatchesVisibleISIN(t *testing.T) {
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	_, person, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	svc := importer.NewService(db)

	// Someone else owns an asset that happens to carry Pfizer's ISIN.
	other, err := db.CreatePerson(ctx, persistence.CreatePersonParams{ID: "other", DisplayName: "Other"})
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := db.CreateSecurity(ctx, persistence.CreateSecurityParams{
		ID: "their-pfizer", DisplayName: "Their Pfizer", PersonID: sql.NullString{String: other.ID, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSecurityIdentifier(ctx, persistence.CreateSecurityIdentifierParams{SecurityID: theirs.ID, Kind: "ISIN", Value: "US7170811035"}); err != nil {
		t.Fatal(err)
	}

	upload := func(personID string) string {
		t.Helper()
		doc, err := svc.Upload(ctx, personID, "buy.txt", "text/plain", []byte(ingBuyText))
		if err != nil {
			t.Fatal(err)
		}
		return doc.ID
	}
	stage := func(docID, personID string) sql.NullString {
		t.Helper()
		if err := svc.Process(ctx, docID, personID); err != nil {
			t.Fatal(err)
		}
		staged, err := db.ListStagedTransactions(ctx, docID)
		if err != nil || len(staged) == 0 {
			t.Fatalf("staged = %v, %v", staged, err)
		}
		return staged[0].SecurityID
	}

	devDoc := upload(person.ID)
	if got := stage(devDoc, person.ID); got.Valid {
		t.Errorf("dev's document matched %q, want no match: that asset is someone else's", got.String)
	}
	if got := stage(upload(other.ID), other.ID); !got.Valid || got.String != theirs.ID {
		t.Errorf("the owner's document matched %v, want their own asset", got)
	}

	// An ISIN names one asset on the whole server (unique index on the
	// identifiers), so the shared Pfizer can only appear once theirs is
	// gone — and then everyone's documents match it.
	if _, err := db.DeleteSecurity(ctx, persistence.DeleteSecurityParams{ID: theirs.ID, PersonID: sql.NullString{String: other.ID, Valid: true}}); err != nil {
		t.Fatal(err)
	}
	shared, err := db.CreateSecurity(ctx, persistence.CreateSecurityParams{ID: "pfizer", DisplayName: "Pfizer Inc."})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateSecurityIdentifier(ctx, persistence.CreateSecurityIdentifierParams{SecurityID: shared.ID, Kind: "ISIN", Value: "US7170811035"}); err != nil {
		t.Fatal(err)
	}
	if got := stage(devDoc, person.ID); !got.Valid || got.String != shared.ID {
		t.Errorf("dev's re-analyzed document matched %v, want the shared security", got)
	}
}

// TestBookAsset books the office room: one mutation makes the asset, its
// price series, the transfer-in at cost into an "Other assets" portfolio
// created on the spot, and today's valuation — and a second asset reuses
// the portfolio.
func TestBookAsset(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)

	// The portfolio does not exist yet, so the account is needed.
	execExpectError(t, schema, ctx, `mutation {
		bookAsset(input: { displayName: "Office room", kind: PROPERTY, currency: "EUR", acquiredAt: "2024-03-01T12:00:00Z", cost: "40000" }) { id }
	}`, nil)
	execExpectError(t, schema, ctx, `mutation ($account: ID!) {
		bookAsset(input: { displayName: "Office room", kind: PROPERTY, currency: "EUR", acquiredAt: "2024-03-01T12:00:00Z", cost: "0", cashAccountID: $account }) { id }
	}`, map[string]any{"account": accountID})

	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		bookAsset(input: {
			displayName: "Office room", kind: PROPERTY, currency: "EUR"
			acquiredAt: "2024-03-01T12:00:00Z", cost: "40000", valueToday: "42000"
			cashAccountID: $account
		}) {
			id kind
			owner { id }
			listings { id ticker quoteProvider latestQuote { price { amount currency } } }
		}
	}`, map[string]any{"account": accountID})
	room := data["bookAsset"].(map[string]any)
	if room["kind"] != "PROPERTY" || room["owner"].(map[string]any)["id"] != "dev" {
		t.Errorf("booked = %v, want a PROPERTY owned by dev", room)
	}
	listing := room["listings"].([]any)[0].(map[string]any)
	if listing["ticker"] != "OFFICE-ROOM" || listing["quoteProvider"] != nil {
		t.Errorf("listing = %v, want OFFICE-ROOM without a provider", listing)
	}
	if price := listing["latestQuote"].(map[string]any)["price"].(map[string]any); price["amount"] != "42000.00000000" || price["currency"] != "EUR" {
		t.Errorf("valuation = %v, want 42000 EUR", price)
	}

	// Listed among the person's assets, not among the shared securities.
	data = exec(t, schema, ctx, `{ assets { id } securities { id } }`, nil)
	if n := len(data["assets"].([]any)); n != 1 {
		t.Errorf("assets = %d, want 1", n)
	}
	if n := len(data["securities"].([]any)); n != 0 {
		t.Errorf("securities = %d, want 0 (an owned asset is not shared)", n)
	}

	// Held in "Other assets": one unit, cost basis 40000, worth 42000.
	portfolios := func() []any {
		data := exec(t, schema, ctx, `{ portfolios { displayName snapshot { positions { security { id } units purchaseValue { amount } marketValue { amount } } } } }`, nil)
		return data["portfolios"].([]any)
	}
	ps := portfolios()
	if len(ps) != 1 || ps[0].(map[string]any)["displayName"] != "Other assets" {
		t.Fatalf("portfolios = %v, want one named Other assets", ps)
	}
	positions := ps[0].(map[string]any)["snapshot"].(map[string]any)["positions"].([]any)
	if len(positions) != 1 {
		t.Fatalf("positions = %d, want 1", len(positions))
	}
	pos := positions[0].(map[string]any)
	if pos["units"] != 1.0 || pos["purchaseValue"].(map[string]any)["amount"] != 4000000.0 || pos["marketValue"].(map[string]any)["amount"] != 4200000.0 {
		t.Errorf("position = %v, want 1 unit, 4000000 cost, 4200000 value", pos)
	}

	// A second asset, no value given and no account named: the portfolio
	// exists, so neither is needed, and it is worth what it cost.
	exec(t, schema, ctx, `mutation {
		bookAsset(input: { displayName: "Car", kind: VEHICLE, currency: "EUR", acquiredAt: "2025-06-01T12:00:00Z", cost: "18000" }) { id }
	}`, nil)
	ps = portfolios()
	if len(ps) != 1 {
		t.Fatalf("portfolios after the second asset = %d, want still 1", len(ps))
	}
	positions = ps[0].(map[string]any)["snapshot"].(map[string]any)["positions"].([]any)
	if len(positions) != 2 {
		t.Fatalf("positions = %d, want 2", len(positions))
	}
	for _, p := range positions {
		pos := p.(map[string]any)
		if pos["purchaseValue"].(map[string]any)["amount"] == 1800000.0 && pos["marketValue"].(map[string]any)["amount"] != 1800000.0 {
			t.Errorf("car = %v, want valued at cost", pos)
		}
	}
}

// TestAssetIsolation: an owned asset is invisible to and untouchable by
// another person, while a shared security stays everyone's.
func TestAssetIsolation(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
		shared: createSecurity(input: { displayName: "Vanguard FTSE All-World", listings: [{ ticker: "VWCE", currency: "EUR" }] }) { id listings { id } }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)
	sharedID := data["shared"].(map[string]any)["id"].(string)
	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		bookAsset(input: { displayName: "Office room", kind: PROPERTY, currency: "EUR", acquiredAt: "2024-03-01T12:00:00Z", cost: "40000", cashAccountID: $account }) {
			id listings { id }
		}
	}`, map[string]any{"account": accountID})
	room := data["bookAsset"].(map[string]any)
	roomID := room["id"].(string)
	roomListing := room["listings"].([]any)[0].(map[string]any)["id"].(string)

	// Another person, present only in the context — enough for scoping.
	other := auth.WithPerson(auth.WithUser(context.Background(), &persistence.User{ID: "other"}), &persistence.Person{ID: "other"})

	data = exec(t, schema, other, `query ($room: ID!, $shared: ID!) {
		assets { id }
		room: security(id: $room) { id }
		shared: security(id: $shared) { id }
	}`, map[string]any{"room": roomID, "shared": sharedID})
	if n := len(data["assets"].([]any)); n != 0 {
		t.Errorf("the other person's assets = %d, want 0", n)
	}
	if data["room"] != nil {
		t.Errorf("the other person can see the room: %v", data["room"])
	}
	if data["shared"] == nil {
		t.Errorf("the other person cannot see the shared security")
	}

	for name, q := range map[string]string{
		"update":  `mutation ($id: ID!) { updateSecurity(id: $id, input: { displayName: "Mine now" }) { id } }`,
		"delete":  `mutation ($id: ID!) { deleteSecurity(id: $id) }`,
		"refresh": `mutation ($id: ID!) { triggerQuoteUpdate(securityIDs: [$id]) { updatedListings } }`,
	} {
		t.Run("other cannot "+name, func(t *testing.T) {
			execExpectError(t, schema, other, q, map[string]any{"id": roomID})
		})
	}
	t.Run("other cannot value it", func(t *testing.T) {
		execExpectError(t, schema, other, `mutation ($listing: ID!) {
			recordQuote(listingID: $listing, time: "2026-10-22T12:00:00Z", price: "1") { time }
		}`, map[string]any{"listing": roomListing})
		data := exec(t, schema, other, `mutation ($listing: ID!) {
			deleteQuote(listingID: $listing, time: "2026-10-22T12:00:00Z")
		}`, map[string]any{"listing": roomListing})
		if data["deleteQuote"] != false {
			t.Errorf("deleteQuote by the other person = %v, want false", data["deleteQuote"])
		}
	})

	// The owner can do all of it, and the shared security is editable by both.
	exec(t, schema, ctx, `mutation ($id: ID!) { updateSecurity(id: $id, input: { displayName: "Office room, 25 m²" }) { id } }`, map[string]any{"id": roomID})
	exec(t, schema, other, `mutation ($id: ID!) { updateSecurity(id: $id, input: { displayName: "Vanguard FTSE All-World (Acc)" }) { id } }`, map[string]any{"id": sharedID})
	exec(t, schema, ctx, `mutation ($id: ID!) { updateSecurity(id: $id, input: { displayName: "Vanguard FTSE All-World" }) { id } }`, map[string]any{"id": sharedID})
}
