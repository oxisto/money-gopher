package pdf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	portfoliov1 "github.com/oxisto/money-gopher/gen"
)

type expectedTx struct {
	Type     string  `json:"type"`
	ISIN     string  `json:"isin,omitempty"`
	Amount   float64 `json:"amount,omitempty"`
	PriceEUR float64 `json:"price_eur,omitempty"`
	Valuta   string  `json:"valuta,omitempty"`
}

type expectedSpec struct {
	ExpectedTxCount int          `json:"expected_tx_count"`
	ExpectedTxs      []expectedTx `json:"expected_txs"`
}

func TestImportPDF_Fixtures(t *testing.T) {
	cases := []struct{
		text string
		spec string
	}{
		{"flatex-verkauf-2026-01.txt", "flatex-verkauf-2026-01.expected.json"},
		{"ing-dividende-2025-12.txt", "ing-dividende-2025-12.expected.json"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.text, func(t *testing.T) {
			textPath := filepath.Join("..", "..", "internal", "testdata", c.text)
			specPath := filepath.Join("..", "..", "internal", "testdata", c.spec)

			btxt, err := os.ReadFile(textPath)
			if err != nil {
				t.Fatalf("open text fixture: %v", err)
			}

			var spec expectedSpec
			b, err := os.ReadFile(specPath)
			if err != nil {
				t.Fatalf("read spec: %v", err)
			}
			if err := json.Unmarshal(b, &spec); err != nil {
				t.Fatalf("unmarshal spec: %v", err)
			}

			txs, _, err := parseText(string(btxt), "test-portfolio")
			if err != nil {
				t.Fatalf("parseText returned error: %v", err)
			}

			if len(txs) != spec.ExpectedTxCount {
				t.Fatalf("expected %d transactions, got %d", spec.ExpectedTxCount, len(txs))
			}

			for i, want := range spec.ExpectedTxs {
				if i >= len(txs) {
					t.Fatalf("missing tx #%d", i)
				}
				tx := txs[i]

				// type
				switch want.Type {
				case "SELL":
					if tx.Type != portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_SELL {
						t.Fatalf("tx %d: wanted SELL, got %s", i, tx.Type.String())
					}
				case "BUY":
					if tx.Type != portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_BUY {
						t.Fatalf("tx %d: wanted BUY, got %s", i, tx.Type.String())
					}
				case "DIVIDEND":
					if tx.Type != portfoliov1.PortfolioEventType_PORTFOLIO_EVENT_TYPE_DIVIDEND {
						t.Fatalf("tx %d: wanted DIVIDEND, got %s", i, tx.Type.String())
					}
				default:
					t.Fatalf("unknown expected type %q", want.Type)
				}

				// ISIN
				if want.ISIN != "" && tx.SecurityId != want.ISIN {
					t.Fatalf("tx %d: expected isin %s, got %s", i, want.ISIN, tx.SecurityId)
				}

				// amount
				if want.Amount != 0 {
					if tx.Amount == 0 || (tx.Amount-want.Amount) > 0.0001 {
						t.Fatalf("tx %d: expected amount %v, got %v", i, want.Amount, tx.Amount)
					}
				}

				// price
				if want.PriceEUR != 0 {
					if tx.Price == nil {
						t.Fatalf("tx %d: expected price %v EUR, got nil", i, want.PriceEUR)
					}
					wantCents := int32((want.PriceEUR * 100) + 0.5)
					if tx.Price.GetValue() != wantCents {
						t.Fatalf("tx %d: expected price %v EUR (%d), got %d", i, want.PriceEUR, wantCents, tx.Price.GetValue())
					}
				}

				// valuta
				if want.Valuta != "" {
					if tx.Time == nil {
						t.Fatalf("tx %d: expected valuta %s, got nil time", i, want.Valuta)
					}
					if tx.Time.AsTime().In(time.Local).Format("02.01.2006") != want.Valuta {
						t.Fatalf("tx %d: expected valuta %s, got %s", i, want.Valuta, tx.Time.AsTime().In(time.Local).Format("02.01.2006"))
					}
				}
			}
		})
	}
}

func TestImportPDF_FromPDFFixtures(t *testing.T) {
	cases := []struct{
		pdf  string
		spec string
	}{
		{"flatex-verkauf-2026-01.pdf", "flatex-verkauf-2026-01.expected.json"},
		{"ing-dividende-2025-12.pdf", "ing-dividende-2025-12.expected.json"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.pdf, func(t *testing.T) {
			pdfPath := filepath.Join("..", "..", "internal", "testdata", c.pdf)
			specPath := filepath.Join("..", "..", "internal", "testdata", c.spec)

			f, err := os.Open(pdfPath)
			if err != nil {
				t.Fatalf("open pdf: %v", err)
			}
			defer f.Close()

			var spec expectedSpec
			b, err := os.ReadFile(specPath)
			if err != nil {
				t.Fatalf("read spec: %v", err)
			}
			if err := json.Unmarshal(b, &spec); err != nil {
				t.Fatalf("unmarshal spec: %v", err)
			}

			txs, _, err := Import(f, "test-portfolio")
			if err != nil {
				t.Fatalf("Import returned error: %v", err)
			}

			if len(txs) != spec.ExpectedTxCount {
				t.Fatalf("expected %d transactions, got %d", spec.ExpectedTxCount, len(txs))
			}
		})
	}
}
