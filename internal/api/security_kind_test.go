package api

import "testing"

// TestSecurityKind round-trips the kind: set on creation, changed on
// update, and null when nobody said.
func TestSecurityKind(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		bond: createSecurity(input: { displayName: "4,125 % United States of America", kind: BOND }) { id kind }
		unknown: createSecurity(input: { displayName: "Something" }) { id kind }
	}`, nil)
	bond := data["bond"].(map[string]any)
	if bond["kind"] != "BOND" {
		t.Errorf("created kind = %v, want BOND", bond["kind"])
	}
	if kind := data["unknown"].(map[string]any)["kind"]; kind != nil {
		t.Errorf("kind without input = %v, want null", kind)
	}

	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		updateSecurity(id: $id, input: { kind: ETF }) { kind displayName }
	}`, map[string]any{"id": bond["id"]})
	updated := data["updateSecurity"].(map[string]any)
	if updated["kind"] != "ETF" || updated["displayName"] != "4,125 % United States of America" {
		t.Errorf("updated = %v, want kind ETF with the name kept", updated)
	}

	// A name-only update leaves the kind alone.
	data = exec(t, schema, ctx, `mutation ($id: ID!) {
		updateSecurity(id: $id, input: { displayName: "Renamed" }) { kind }
	}`, map[string]any{"id": bond["id"]})
	if kind := data["updateSecurity"].(map[string]any)["kind"]; kind != "ETF" {
		t.Errorf("kind after a name-only update = %v, want ETF", kind)
	}
}

// TestValuedByHand: an asset nobody quotes — the office room — gets a
// listing without a provider and is priced through recordQuote.
func TestValuedByHand(t *testing.T) {
	schema, ctx := newTestSchema(t)

	data := exec(t, schema, ctx, `mutation {
		createSecurity(input: {
			displayName: "Office room", kind: PROPERTY
			listings: [{ ticker: "OFFICE-ROOM", currency: "EUR" }]
		}) { id kind listings { id quoteProvider latestQuote { price { amount } } } }
	}`, nil)
	room := data["createSecurity"].(map[string]any)
	listing := room["listings"].([]any)[0].(map[string]any)
	if room["kind"] != "PROPERTY" || listing["quoteProvider"] != nil || listing["latestQuote"] != nil {
		t.Fatalf("created = %v, want a PROPERTY with an unquoted, providerless listing", room)
	}

	record := func(day, price string) map[string]any {
		data := exec(t, schema, ctx, `mutation ($listing: ID!, $time: Time!, $price: Decimal!) {
			recordQuote(listingID: $listing, time: $time, price: $price) { time price { amount currency } }
		}`, map[string]any{"listing": listing["id"], "time": day + "T12:00:00Z", "price": price})
		return data["recordQuote"].(map[string]any)
	}
	if got := record("2026-01-15", "40000")["price"].(map[string]any); got["amount"] != "40000.00000000" || got["currency"] != "EUR" {
		t.Errorf("recorded price = %v, want 40000 EUR", got)
	}
	record("2026-10-01", "42000")
	// Same time again replaces rather than duplicates.
	record("2026-10-01", "42500")

	latest := func() (string, int) {
		data := exec(t, schema, ctx, `query ($id: ID!) {
			security(id: $id) { listings { latestQuote { price { amount } } quotes { time } } }
		}`, map[string]any{"id": room["id"]})
		l := data["security"].(map[string]any)["listings"].([]any)[0].(map[string]any)
		return l["latestQuote"].(map[string]any)["price"].(map[string]any)["amount"].(string), len(l["quotes"].([]any))
	}
	if price, n := latest(); price != "42500.00000000" || n != 2 {
		t.Errorf("latest = %s with %d quotes, want 42500 with 2", price, n)
	}

	execExpectError(t, schema, ctx, `mutation ($listing: ID!) {
		recordQuote(listingID: $listing, time: "2026-10-02T12:00:00Z", price: "0") { time }
	}`, map[string]any{"listing": listing["id"]})

	data = exec(t, schema, ctx, `mutation ($listing: ID!) {
		gone: deleteQuote(listingID: $listing, time: "2026-10-01T12:00:00Z")
		again: deleteQuote(listingID: $listing, time: "2026-10-01T12:00:00Z")
	}`, map[string]any{"listing": listing["id"]})
	if data["gone"] != true || data["again"] != false {
		t.Errorf("deleteQuote = %v / %v, want true then false", data["gone"], data["again"])
	}
	if price, n := latest(); price != "40000.00000000" || n != 1 {
		t.Errorf("after delete: latest = %s with %d quotes, want 40000 with 1", price, n)
	}
}
