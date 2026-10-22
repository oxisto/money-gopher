package api

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"
)

// newTestSchema returns a schema over a fresh in-memory database and a
// context authenticated as the dev user. Parsing the schema also verifies
// that the resolvers match api/schema.graphql.
func newTestSchema(t *testing.T) (*graphql.Schema, context.Context) {
	t.Helper()

	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	user, person, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatalf("EnsureDevUser() error = %v", err)
	}

	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry()}

	ctx := auth.WithPerson(auth.WithUser(context.Background(), user), person)
	return NewSchema(db, updater), ctx
}

// exec runs a GraphQL query and decodes the response data into a generic map,
// failing the test on any GraphQL error.
func exec(t *testing.T, schema *graphql.Schema, ctx context.Context, query string, vars map[string]any) map[string]any {
	t.Helper()

	resp := schema.Exec(ctx, query, "", vars)
	if len(resp.Errors) > 0 {
		t.Fatalf("query %q returned errors: %v", query, resp.Errors)
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	return data
}

// execExpectError runs a GraphQL query and asserts that it fails.
func execExpectError(t *testing.T, schema *graphql.Schema, ctx context.Context, query string, vars map[string]any) {
	t.Helper()

	if resp := schema.Exec(ctx, query, "", vars); len(resp.Errors) == 0 {
		t.Fatalf("query %q unexpectedly succeeded", query)
	}
}

func TestLifecycle(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `{ version me { displayName } }`, nil)
	if got := data["me"].(map[string]any)["displayName"]; got != "Development User" {
		t.Errorf("me.displayName = %v", got)
	}

	// Set up an account, a portfolio and a security with a listing.
	data = exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)

	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		createPortfolio(input: { displayName: "My Portfolio", cashAccountID: $account }) { id }
	}`, map[string]any{"account": accountID})
	portfolioID := data["createPortfolio"].(map[string]any)["id"].(string)

	data = exec(t, schema, ctx, `mutation {
		createAsset(input: {
			displayName: "Vanguard FTSE All-World"
			identifiers: [{ kind: "ISIN", value: "IE00BK5BQT80" }]
			listings: [{ ticker: "VWCE", exchange: "XETR", currency: "EUR" }]
		}) { id identifiers { kind value } listings { ticker } }
	}`, nil)
	assetID := data["createAsset"].(map[string]any)["id"].(string)

	// Deposit cash, then buy: 10 x 95.00 EUR + 1.50 fees.
	exec(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: DEPOSIT_CASH, time: "2026-01-02T10:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: 100000
		}) { id }
	}`, map[string]any{"account": accountID})

	data = exec(t, schema, ctx, `mutation ($portfolio: ID!, $asset: ID!, $account: ID!) {
		createTransaction(input: {
			type: BUY, time: "2026-01-03T10:00:00Z", currency: "EUR"
			portfolioID: $portfolio, assetID: $asset, cashAccountID: $account
			units: 10, price: "95.00", fees: 150
		}) { id cashDelta { amount currency } document { id } }
	}`, map[string]any{"portfolio": portfolioID, "asset": assetID, "account": accountID})
	buy := data["createTransaction"].(map[string]any)
	buyID := buy["id"].(string)
	if got := buy["cashDelta"].(map[string]any)["amount"].(float64); got != -95150 {
		t.Errorf("derived cashDelta = %v, want -95150", got)
	}
	// A manually created transaction has no source document.
	if buy["document"] != nil {
		t.Errorf("document = %v, want nil for a manual transaction", buy["document"])
	}

	// The account balance reflects both transactions; the portfolio only
	// sees the buy.
	data = exec(t, schema, ctx, `query ($id: ID!) {
		cashAccount(id: $id) { balance { amount } transactions { id } }
	}`, map[string]any{"id": accountID})
	account := data["cashAccount"].(map[string]any)
	if got := account["balance"].(map[string]any)["amount"].(float64); got != 4850 {
		t.Errorf("balance = %v, want 4850", got)
	}
	if got := len(account["transactions"].([]any)); got != 2 {
		t.Errorf("account transactions = %v, want 2", got)
	}

	data = exec(t, schema, ctx, `query ($id: ID!) {
		portfolio(id: $id) { transactions { type asset { displayName } } }
	}`, map[string]any{"id": portfolioID})
	transactions := data["portfolio"].(map[string]any)["transactions"].([]any)
	if len(transactions) != 1 {
		t.Fatalf("portfolio transactions = %v, want 1", len(transactions))
	}

	// Updating the units re-derives the cash delta.
	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		updateTransaction(id: $id, input: { units: 5 }) { cashDelta { amount } }
	}`, map[string]any{"id": buyID})
	if got := data["updateTransaction"].(map[string]any)["cashDelta"].(map[string]any)["amount"].(float64); got != -47650 {
		t.Errorf("re-derived cashDelta = %v, want -47650", got)
	}

	// Invalid transactions are rejected.
	execExpectError(t, schema, ctx, `mutation ($portfolio: ID!, $asset: ID!) {
		createTransaction(input: {
			type: BUY, time: "2026-01-03T10:00:00Z", currency: "EUR"
			portfolioID: $portfolio, assetID: $asset
			units: 10, price: "95.00"
		}) { id }
	}`, map[string]any{"portfolio": portfolioID, "asset": assetID})

	execExpectError(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: WITHDRAW_CASH, time: "2026-01-04T10:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: 500
		}) { id }
	}`, map[string]any{"account": accountID})

	// A security referenced by transactions cannot be deleted.
	execExpectError(t, schema, ctx, `mutation ($id: ID!) { deleteAsset(id: $id) }`,
		map[string]any{"id": assetID})

	// Deleting the transaction and portfolio works and restores the balance.
	exec(t, schema, ctx, `mutation ($id: ID!) { deleteTransaction(id: $id) }`,
		map[string]any{"id": buyID})
	exec(t, schema, ctx, `mutation ($id: ID!) { deletePortfolio(id: $id) }`,
		map[string]any{"id": portfolioID})

	data = exec(t, schema, ctx, `query ($id: ID!) { cashAccount(id: $id) { balance { amount } } }`,
		map[string]any{"id": accountID})
	if got := data["cashAccount"].(map[string]any)["balance"].(map[string]any)["amount"].(float64); got != 100000 {
		t.Errorf("balance after cleanup = %v, want 100000", got)
	}
}

// TestUpdateAsset_ListingReconciliation verifies that editing an
// existing listing (matched by ListingInput.id) preserves its quote
// history, while omitting a listing from the input set removes it (and its
// quotes) — the whole reason reconcileListings replaced a blind
// delete-all/recreate.
func TestUpdateAsset_ListingReconciliation(t *testing.T) {
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	user, person, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatalf("EnsureDevUser() error = %v", err)
	}
	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry()}
	schema := NewSchema(db, updater)
	ctx := auth.WithPerson(auth.WithUser(context.Background(), user), person)

	data := exec(t, schema, ctx, `mutation {
		createAsset(input: {
			displayName: "Test Corp"
			listings: [{ ticker: "TST", currency: "EUR", quoteProvider: "yf" }]
		}) { id listings { id } }
	}`, nil)
	sec := data["createAsset"].(map[string]any)
	assetID := sec["id"].(string)
	listingID := sec["listings"].([]any)[0].(map[string]any)["id"].(string)

	if err := db.CreateQuote(ctx, persistence.CreateQuoteParams{
		ListingID: listingID,
		Time:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Price:     persistence.NewDecimalFromInt64(10000).DivInt(100),
	}); err != nil {
		t.Fatalf("CreateQuote() error = %v", err)
	}

	// Editing the listing (same ID, new ticker) must keep its ID — and so
	// its quote history — intact, not delete-and-recreate it.
	data = exec(t, schema, ctx, `mutation ($id: ID!, $listingID: ID!) {
		updateAsset(id: $id, input: {
			listings: [{ id: $listingID, ticker: "TST2", currency: "EUR", quoteProvider: "yf" }]
		}) { listings { id ticker quotes(from: null, to: null) { price { amount } } } }
	}`, map[string]any{"id": assetID, "listingID": listingID})

	listings := data["updateAsset"].(map[string]any)["listings"].([]any)
	if len(listings) != 1 {
		t.Fatalf("len(listings) = %d, want 1", len(listings))
	}
	updated := listings[0].(map[string]any)
	if updated["id"] != listingID {
		t.Errorf("listing ID changed: got %v, want %v (quote history would be orphaned)", updated["id"], listingID)
	}
	if updated["ticker"] != "TST2" {
		t.Errorf("ticker = %v, want TST2", updated["ticker"])
	}
	if qs := updated["quotes"].([]any); len(qs) != 1 {
		t.Errorf("len(quotes) = %d, want 1 (quote history should survive an in-place edit)", len(qs))
	}

	// Dropping the listing from the input set entirely must remove it.
	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		updateAsset(id: $id, input: { listings: [] }) { listings { id } }
	}`, map[string]any{"id": assetID})
	if got := data["updateAsset"].(map[string]any)["listings"].([]any); len(got) != 0 {
		t.Errorf("listings after removal = %v, want empty", got)
	}
}

// TestTaxRefundEitherSign verifies that TAX_REFUND accepts a negative
// cashDelta as well as positive — a retroactive capital-gains loss offset
// can net to an additional small charge, not just a refund — but still
// rejects zero.
func TestTaxRefundEitherSign(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Extra-Konto", currency: "EUR" }) { id }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)

	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: TAX_REFUND, time: "2025-03-19T00:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: -2
		}) { id cashDelta { amount } }
	}`, map[string]any{"account": accountID})
	if got := data["createTransaction"].(map[string]any)["cashDelta"].(map[string]any)["amount"].(float64); got != -2 {
		t.Errorf("negative TAX_REFUND cashDelta = %v, want -2", got)
	}

	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: TAX_REFUND, time: "2025-05-06T00:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: 1262
		}) { id cashDelta { amount } }
	}`, map[string]any{"account": accountID})
	if got := data["createTransaction"].(map[string]any)["cashDelta"].(map[string]any)["amount"].(float64); got != 1262 {
		t.Errorf("positive TAX_REFUND cashDelta = %v, want 1262", got)
	}

	execExpectError(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: TAX_REFUND, time: "2025-01-01T00:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: 0
		}) { id }
	}`, map[string]any{"account": accountID})
}

// TestAssetUnderlying verifies linking a derivative asset (e.g. a
// subscription right) to the stock it derives from, both at creation and
// via a later update, and that the link resolves both ways round-trip
// (fetching it back by ID sees the same underlying).
func TestAssetUnderlying(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createAsset(input: { displayName: "Eutelsat Communications" }) { id }
	}`, nil)
	stockID := data["createAsset"].(map[string]any)["id"].(string)

	// Link at creation time.
	data = exec(t, schema, ctx, `mutation ($underlying: ID!) {
		createAsset(input: {
			displayName: "Eutelsat Communications Anrechte"
			underlyingAssetID: $underlying
		}) { id underlyingAsset { id displayName } }
	}`, map[string]any{"underlying": stockID})
	rights := data["createAsset"].(map[string]any)
	rightsID := rights["id"].(string)
	underlying := rights["underlyingAsset"].(map[string]any)
	if underlying["id"] != stockID {
		t.Errorf("underlyingAsset.id = %v, want %v", underlying["id"], stockID)
	}
	if underlying["displayName"] != "Eutelsat Communications" {
		t.Errorf("underlyingAsset.displayName = %v, want Eutelsat Communications", underlying["displayName"])
	}

	// A plain asset (no underlying) resolves it as null.
	data = exec(t, schema, ctx, `query ($id: ID!) { asset(id: $id) { underlyingAsset { id } } }`,
		map[string]any{"id": stockID})
	if got := data["asset"].(map[string]any)["underlyingAsset"]; got != nil {
		t.Errorf("stock underlyingAsset = %v, want nil", got)
	}

	// Link via a later update, e.g. fixing a security created without one —
	// the exact scenario a misattributed DELIVERY_INBOUND needs.
	data = exec(t, schema, ctx, `mutation {
		createAsset(input: { displayName: "Some Warrant" }) { id }
	}`, nil)
	warrantID := data["createAsset"].(map[string]any)["id"].(string)

	data = exec(t, schema, ctx, `mutation ($id: ID!, $underlying: ID!) {
		updateAsset(id: $id, input: { underlyingAssetID: $underlying }) {
			underlyingAsset { id }
		}
	}`, map[string]any{"id": warrantID, "underlying": rightsID})
	if got := data["updateAsset"].(map[string]any)["underlyingAsset"].(map[string]any)["id"]; got != rightsID {
		t.Errorf("updateAsset underlyingAsset.id = %v, want %v", got, rightsID)
	}
}

// TestUserIsolation ensures one user cannot see or modify another user's data.
func TestUserIsolation(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createCashAccount(input: { displayName: "Giro", currency: "EUR" }) { id }
	}`, nil)
	accountID := data["createCashAccount"].(map[string]any)["id"].(string)

	data = exec(t, schema, ctx, `mutation ($account: ID!) {
		createPortfolio(input: { displayName: "Mine", cashAccountID: $account }) { id }
	}`, map[string]any{"account": accountID})
	portfolioID := data["createPortfolio"].(map[string]any)["id"].(string)

	// An unauthenticated request cannot see anything.
	execExpectError(t, schema, context.Background(), `{ portfolios { id } }`, nil)

	// Another person sees an empty list and cannot touch the portfolio.
	other := auth.WithPerson(
		auth.WithUser(context.Background(), &persistence.User{ID: "other"}),
		&persistence.Person{ID: "other"},
	)
	if got := len(exec(t, schema, other, `{ portfolios { id } }`, nil)["portfolios"].([]any)); got != 0 {
		t.Errorf("other user portfolios = %v, want 0", got)
	}
	execExpectError(t, schema, other, `mutation ($id: ID!) { deletePortfolio(id: $id) }`,
		map[string]any{"id": portfolioID})
}
