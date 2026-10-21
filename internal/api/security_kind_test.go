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
