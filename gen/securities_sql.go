package portfoliov1

import (
	"database/sql"
	"strings"
	"time"

	"github.com/oxisto/money-gopher/persistence"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ persistence.StorageObject = &Security{}

func (*Security) InitTables(db *persistence.DB) (err error) {
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS securities (
name TEXT PRIMARY KEY,
display_name TEXT NOT NULL,
quote_provider TEXT
);`)
	if err != nil {
		return err
	}

	return
}

func (*ListedSecurity) InitTables(db *persistence.DB) (err error) {
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS listed_securities (
security_name TEXT,
ticker TEXT NOT NULL,
currency TEXT NOT NULL,
latest_quote INTEGER,
latest_quote_timestamp DATETIME,
PRIMARY KEY (security_name, ticker)
);`)
	if err != nil {
		return err
	}

	return
}

func (*Security) PrepareReplace(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`REPLACE INTO securities (name, display_name, quote_provider) VALUES (?,?,?);`)
}

func (*ListedSecurity) PrepareReplace(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`REPLACE INTO listed_securities (security_name, ticker, currency, latest_quote, latest_quote_timestamp) VALUES (?,?,?,?,?);`)
}

func (*Security) PrepareList(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, display_name, quote_provider FROM securities`)
}

func (*ListedSecurity) PrepareList(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT security_name, ticker, currency, latest_quote, latest_quote_timestamp FROM listed_securities WHERE security_name = ?`)
}

func (*Security) PrepareGet(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, display_name, quote_provider FROM securities WHERE name = ?`)
}

func (*ListedSecurity) PrepareGet(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT * FROM listed_securities WHERE security_name = ? AND ticker = ?`)
}

func (*Security) PrepareUpdate(db *persistence.DB, columns []string) (stmt *sql.Stmt, err error) {
	// We need to make sure to quote columns here because they are potentially evil user input
	var (
		query string
		set   []string
	)

	set = make([]string, len(columns))
	for i, col := range columns {
		set[i] = persistence.Quote(col) + " = ?"
	}

	query += "UPDATE securities SET " + strings.Join(set, ", ") + " WHERE name = ?;"

	return db.Prepare(query)
}

func (*ListedSecurity) PrepareUpdate(db *persistence.DB, columns []string) (stmt *sql.Stmt, err error) {
	// We need to make sure to quote columns here because they are potentially evil user input
	var (
		query string
		set   []string
	)

	set = make([]string, len(columns))
	for i, col := range columns {
		set[i] = persistence.Quote(col) + " = ?"
	}

	query += "UPDATE listed_securities SET " + strings.Join(set, ", ") + " WHERE security_name = ? AND ticker = ?;"

	return db.Prepare(query)
}

func (*Security) PrepareDelete(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`DELETE FROM securities WHERE name = ?`)
}

func (*ListedSecurity) PrepareDelete(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`DELETE FROM listed_securities WHERE security_name = ? AND ticker = ?`)
}

func (s *Security) ReplaceIntoArgs() []any {
	return []any{s.Name, s.DisplayName, s.QuoteProvider}
}

func (l *ListedSecurity) ReplaceIntoArgs() []any {
	var (
		pt    *time.Time
		t     time.Time
		value sql.NullInt32
	)

	if l.LatestQuoteTimestamp != nil {
		t = l.LatestQuoteTimestamp.AsTime()
		pt = &t
	}

	if l.LatestQuote != nil {
		value.Int32 = l.LatestQuote.Value
		value.Valid = true
	}

	return []any{l.SecurityName, l.Ticker, l.Currency, value, pt}
}

func (s *Security) UpdateArgs(columns []string) (args []any) {
	for _, col := range columns {
		switch col {
		case "name":
			args = append(args, s.Name)
		case "display_name":
			args = append(args, s.DisplayName)
		case "quote_provider":
			args = append(args, s.QuoteProvider)
		}
	}

	return args
}

func (l *ListedSecurity) UpdateArgs(columns []string) (args []any) {
	for _, col := range columns {
		switch col {
		case "security_name":
			args = append(args, l.SecurityName)
		case "ticker":
			args = append(args, l.Ticker)
		case "currency":
			args = append(args, l.LatestQuote.GetSymbol())
		case "latest_quote":
			args = append(args, l.LatestQuote.GetValue())
		case "latest_quote_timestamp":
			if l.LatestQuoteTimestamp != nil {
				args = append(args, l.LatestQuoteTimestamp.AsTime())
			} else {
				args = append(args, nil)
			}
		}
	}

	return args
}

func (*Security) Scan(sc persistence.Scanner) (obj persistence.StorageObject, err error) {
	var (
		s Security
	)

	err = sc.Scan(&s.Name, &s.DisplayName, &s.QuoteProvider)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (*ListedSecurity) Scan(sc persistence.Scanner) (obj persistence.StorageObject, err error) {
	var (
		l     ListedSecurity
		t     sql.NullTime
		value sql.NullInt32
	)

	err = sc.Scan(&l.SecurityName, &l.Ticker, &l.Currency, &value, &t)
	if err != nil {
		return nil, err
	}

	if t.Valid {
		l.LatestQuoteTimestamp = timestamppb.New(t.Time)
	}

	if value.Valid {
		l.LatestQuote = Value(value.Int32)
		l.LatestQuote.Symbol = l.Currency
	}

	return &l, nil
}
