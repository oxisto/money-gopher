package api

import (
	"context"
	"encoding/json"
	"testing"

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

	user, err := auth.EnsureDevUser(context.Background(), db)
	if err != nil {
		t.Fatalf("EnsureDevUser() error = %v", err)
	}

	updater := &quotes.Updater{DB: db, Registry: quotes.NewRegistry()}

	return NewSchema(db, updater), auth.WithUser(context.Background(), user)
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
		createSecurity(input: {
			displayName: "Vanguard FTSE All-World"
			identifiers: [{ kind: "ISIN", value: "IE00BK5BQT80" }]
			listings: [{ ticker: "VWCE", exchange: "XETR", currency: "EUR" }]
		}) { id identifiers { kind value } listings { ticker } }
	}`, nil)
	securityID := data["createSecurity"].(map[string]any)["id"].(string)

	// Deposit cash, then buy: 10 x 95.00 EUR + 1.50 fees.
	exec(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: DEPOSIT_CASH, time: "2026-01-02T10:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: 100000
		}) { id }
	}`, map[string]any{"account": accountID})

	data = exec(t, schema, ctx, `mutation ($portfolio: ID!, $security: ID!, $account: ID!) {
		createTransaction(input: {
			type: BUY, time: "2026-01-03T10:00:00Z", currency: "EUR"
			portfolioID: $portfolio, securityID: $security, cashAccountID: $account
			units: 10, price: 9500, fees: 150
		}) { id cashDelta { amount currency } }
	}`, map[string]any{"portfolio": portfolioID, "security": securityID, "account": accountID})
	buy := data["createTransaction"].(map[string]any)
	buyID := buy["id"].(string)
	if got := buy["cashDelta"].(map[string]any)["amount"].(float64); got != -95150 {
		t.Errorf("derived cashDelta = %v, want -95150", got)
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
		portfolio(id: $id) { transactions { type security { displayName } } }
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
	execExpectError(t, schema, ctx, `mutation ($portfolio: ID!, $security: ID!) {
		createTransaction(input: {
			type: BUY, time: "2026-01-03T10:00:00Z", currency: "EUR"
			portfolioID: $portfolio, securityID: $security
			units: 10, price: 9500
		}) { id }
	}`, map[string]any{"portfolio": portfolioID, "security": securityID})

	execExpectError(t, schema, ctx, `mutation ($account: ID!) {
		createTransaction(input: {
			type: WITHDRAW_CASH, time: "2026-01-04T10:00:00Z", currency: "EUR"
			cashAccountID: $account, cashDelta: 500
		}) { id }
	}`, map[string]any{"account": accountID})

	// A security referenced by transactions cannot be deleted.
	execExpectError(t, schema, ctx, `mutation ($id: ID!) { deleteSecurity(id: $id) }`,
		map[string]any{"id": securityID})

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

	// Another user sees an empty list and cannot touch the portfolio.
	other := auth.WithUser(context.Background(), &persistence.User{ID: "other"})
	if got := len(exec(t, schema, other, `{ portfolios { id } }`, nil)["portfolios"].([]any)); got != 0 {
		t.Errorf("other user portfolios = %v, want 0", got)
	}
	execExpectError(t, schema, other, `mutation ($id: ID!) { deletePortfolio(id: $id) }`,
		map[string]any{"id": portfolioID})
}
