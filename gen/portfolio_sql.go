package portfoliov1

import (
	"database/sql"
	"strings"
	"time"

	"github.com/oxisto/money-gopher/persistence"
	"golang.org/x/exp/slog"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

var _ persistence.StorageObject = &Portfolio{}

func (*Portfolio) InitTables(db *persistence.DB) (err error) {
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS portfolios (
name TEXT PRIMARY KEY,
display_name TEXT NOT NULL
);`)
	if err != nil {
		return err
	}

	slog.Info("Some test")

	return
}

func (*Portfolio) PrepareReplace(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`REPLACE INTO portfolios (name, display_name) VALUES (?,?,?);`)
}

func (*Portfolio) PrepareList(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, display_name FROM portfolios`)
}

func (*Portfolio) PrepareGet(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, display_name FROM portfolios WHERE name = ?`)
}

func (*Portfolio) PrepareUpdate(db *persistence.DB, columns []string) (stmt *sql.Stmt, err error) {
	// We need to make sure to quote columns here because they are potentially evil user input
	var (
		query string
		set   []string
	)

	set = make([]string, len(columns))
	for i, col := range columns {
		set[i] = persistence.Quote(col) + " = ?"
	}

	query += "UPDATE portfolios SET " + strings.Join(set, ", ") + " WHERE name = ?;"

	return db.Prepare(query)
}

func (*Portfolio) PrepareDelete(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`DELETE FROM portfolios WHERE name = ?`)
}

func (p *Portfolio) ReplaceIntoArgs() []any {
	return []any{p.Name, p.DisplayName}
}

func (p *Portfolio) UpdateArgs(columns []string) (args []any) {
	for _, col := range columns {
		switch col {
		case "name":
			args = append(args, p.Name)
		case "display_name":
			args = append(args, p.DisplayName)
		}
	}

	return args
}

func (*Portfolio) Scan(sc persistence.Scanner) (obj persistence.StorageObject, err error) {
	var (
		p Portfolio
	)

	err = sc.Scan(&p.Name, &p.DisplayName)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (*PortfolioEvent) InitTables(db *persistence.DB) (err error) {
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS listed_securities (
security_name TEXT,
ticker TEXT NOT NULL,
currency TEXT NOT NULL,
latest_quote REAL,
latest_quote_timestamp DATETIME,
PRIMARY KEY (security_name, ticker)
);`)
	if err != nil {
		return err
	}

	return
}

func (*ListedSecurity) PrepareReplace(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`REPLACE INTO listed_securities (security_name, ticker, currency, latest_quote, latest_quote_timestamp) VALUES (?,?,?,?,?);`)
}

func (*ListedSecurity) PrepareList(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT security_name, ticker, currency, latest_quote, latest_quote_timestamp FROM listed_securities WHERE security_name = ?`)
}

func (*ListedSecurity) PrepareGet(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT * FROM listed_securities WHERE security_name = ? AND ticker = ?`)
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

func (*ListedSecurity) PrepareDelete(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`DELETE FROM listed_securities WHERE security_name = ? AND ticker = ?`)
}

func (l *ListedSecurity) ReplaceIntoArgs() []any {
	var (
		pt *time.Time
		t  time.Time
	)

	if l.LatestQuoteTimestamp != nil {
		t = l.LatestQuoteTimestamp.AsTime()
		pt = &t
	}

	return []any{l.SecurityName, l.Ticker, l.Currency, l.LatestQuote, pt}
}

func (l *ListedSecurity) UpdateArgs(columns []string) (args []any) {
	for _, col := range columns {
		switch col {
		case "security_name":
			args = append(args, l.SecurityName)
		case "ticker":
			args = append(args, l.Ticker)
		case "currency":
			args = append(args, l.Currency)
		case "latest_quote":
			args = append(args, l.LatestQuote)
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

func (*ListedSecurity) Scan(sc persistence.Scanner) (obj persistence.StorageObject, err error) {
	var (
		l ListedSecurity
		t sql.NullTime
	)

	err = sc.Scan(&l.SecurityName, &l.Ticker, &l.Currency, &l.LatestQuote, &t)
	if err != nil {
		return nil, err
	}

	if t.Valid {
		l.LatestQuoteTimestamp = timestamppb.New(t.Time)
	}

	return &l, nil
}
