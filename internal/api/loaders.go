package api

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"

	"github.com/oxisto/money-gopher/internal/persistence"
)

// A list of transactions resolves the same cash account, portfolio,
// schedule and document links row after row. Loaders memoise those
// lookups for the duration of one request, and load the document links
// of a whole account in one query the first time any of its
// transactions asks. Without loaders in the context — tests calling the
// schema directly — the resolvers query as before.
type loaders struct {
	mu         sync.Mutex
	accounts   map[string]*persistence.CashAccount
	portfolios map[string]*persistence.Portfolio
	securities map[string]*persistence.Security
	schedules  map[string]*persistence.Schedule
	// docs by cash account: transaction ID → its documents, oldest link first.
	docs map[string]map[string][]*persistence.Document
}

type loadersKey struct{}

func withLoaders(ctx context.Context) context.Context {
	return context.WithValue(ctx, loadersKey{}, &loaders{
		accounts:   map[string]*persistence.CashAccount{},
		portfolios: map[string]*persistence.Portfolio{},
		securities: map[string]*persistence.Security{},
		schedules:  map[string]*persistence.Schedule{},
		docs:       map[string]map[string][]*persistence.Document{},
	})
}

func loadersFrom(ctx context.Context) *loaders {
	l, _ := ctx.Value(loadersKey{}).(*loaders)
	return l
}

// withRequestLoaders gives every request its own loaders.
func withRequestLoaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(withLoaders(r.Context())))
	})
}

// memo runs load once per key for the request, when loaders are present.
func memo[T any](ctx context.Context, cache func(*loaders) map[string]T, key string, load func() (T, error)) (T, error) {
	l := loadersFrom(ctx)
	if l == nil {
		return load()
	}
	l.mu.Lock()
	m := cache(l)
	if v, ok := m[key]; ok {
		l.mu.Unlock()
		return v, nil
	}
	l.mu.Unlock()
	v, err := load()
	if err != nil {
		return v, err
	}
	l.mu.Lock()
	m[key] = v
	l.mu.Unlock()
	return v, nil
}

func cashAccountByID(ctx context.Context, db *persistence.DB, id, personID string) (*persistence.CashAccount, error) {
	return memo(ctx, func(l *loaders) map[string]*persistence.CashAccount { return l.accounts }, id+"|"+personID, func() (*persistence.CashAccount, error) {
		return db.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: id, PersonID: personID})
	})
}

func portfolioByID(ctx context.Context, db *persistence.DB, id, personID string) (*persistence.Portfolio, error) {
	return memo(ctx, func(l *loaders) map[string]*persistence.Portfolio { return l.portfolios }, id+"|"+personID, func() (*persistence.Portfolio, error) {
		return db.GetPortfolio(ctx, persistence.GetPortfolioParams{ID: id, PersonID: personID})
	})
}

func securityByID(ctx context.Context, db *persistence.DB, id string) (*persistence.Security, error) {
	return memo(ctx, func(l *loaders) map[string]*persistence.Security { return l.securities }, id, func() (*persistence.Security, error) {
		return db.GetSecurity(ctx, id)
	})
}

func scheduleByID(ctx context.Context, db *persistence.DB, id, personID string) (*persistence.Schedule, error) {
	return memo(ctx, func(l *loaders) map[string]*persistence.Schedule { return l.schedules }, id+"|"+personID, func() (*persistence.Schedule, error) {
		return db.GetSchedule(ctx, persistence.GetScheduleParams{ID: id, PersonID: personID})
	})
}

// lightDocument is a document row without its file and extracted text.
func lightDocument(id, filename, contentType, state string, detectedBank, errText, iban sql.NullString, txDate sql.NullTime, createdAt time.Time, personID string, hash sql.NullString) *persistence.Document {
	return &persistence.Document{
		ID: id, Filename: filename, ContentType: contentType, State: state,
		DetectedBank: detectedBank, Error: errText, SettlementIban: iban,
		TransactionDate: txDate, CreatedAt: createdAt, PersonID: personID, ContentHash: hash,
	}
}

// transactionDocuments is the documents describing tx. With loaders, the
// first transaction of an account loads the links of the whole account.
func transactionDocuments(ctx context.Context, db *persistence.DB, tx *persistence.Transaction) ([]*persistence.Document, error) {
	l := loadersFrom(ctx)
	if l == nil || !tx.CashAccountID.Valid {
		rows, err := db.ListTransactionDocuments(ctx, tx.ID)
		if err != nil {
			return nil, err
		}
		out := make([]*persistence.Document, 0, len(rows))
		for _, r := range rows {
			out = append(out, lightDocument(r.ID, r.Filename, r.ContentType, r.State, r.DetectedBank, r.Error, r.SettlementIban, r.TransactionDate, r.CreatedAt, r.PersonID, r.ContentHash))
		}
		return out, nil
	}
	account := tx.CashAccountID.String
	l.mu.Lock()
	byTx, ok := l.docs[account]
	l.mu.Unlock()
	if !ok {
		rows, err := db.ListTransactionDocumentsByCashAccount(ctx, tx.CashAccountID)
		if err != nil {
			return nil, err
		}
		byTx = map[string][]*persistence.Document{}
		for _, r := range rows {
			byTx[r.TransactionID] = append(byTx[r.TransactionID], lightDocument(r.ID, r.Filename, r.ContentType, r.State, r.DetectedBank, r.Error, r.SettlementIban, r.TransactionDate, r.CreatedAt, r.PersonID, r.ContentHash))
		}
		l.mu.Lock()
		l.docs[account] = byTx
		l.mu.Unlock()
	}
	return byTx[tx.ID], nil
}
