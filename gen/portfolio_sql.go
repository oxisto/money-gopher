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
	return db.Prepare(`REPLACE INTO portfolios (name, display_name) VALUES (?,?);`)
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
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS portfolio_events (
id INTEGER PRIMARY KEY AUTOINCREMENT,
time DATETIME NOT NULL,
portfolio_name TEXT NOT NULL, 
security_name TEXT NOT NULL,
amount REAL,
price REAL,
fees REAL,
taxes REAL
);`)
	if err != nil {
		return err
	}

	return
}

func (*PortfolioEvent) PrepareReplace(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`REPLACE INTO portfolio_events
(id, time, portfolio_name, security_name, amount, price, fees, taxes)
VALUES (?,?,?,?,?,?,?,?);`)
}

func (*PortfolioEvent) PrepareList(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT id, time, portfolio_name, security_name, amount, price, fees, taxes
FROM portfolio_events WHERE portfolio_name = ?`)
}

func (*PortfolioEvent) PrepareGet(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT * FROM portfolio_events WHERE id = ?`)
}

func (*PortfolioEvent) PrepareUpdate(db *persistence.DB, columns []string) (stmt *sql.Stmt, err error) {
	// We need to make sure to quote columns here because they are potentially evil user input
	var (
		query string
		set   []string
	)

	set = make([]string, len(columns))
	for i, col := range columns {
		set[i] = persistence.Quote(col) + " = ?"
	}

	query += "UPDATE portfolio_events SET " + strings.Join(set, ", ") + " WHERE id = ?;"

	return db.Prepare(query)
}

func (*PortfolioEvent) PrepareDelete(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`DELETE FROM portfolio_events WHERE id = ?`)
}

func (e *PortfolioEvent) ReplaceIntoArgs() []any {
	return []any{
		e.Id,
		e.Time.AsTime(),
		e.PortfolioName,
		e.SecurityName,
		e.Amount,
		e.Price,
		e.Fees,
		e.Taxes,
	}
}

func (e *PortfolioEvent) UpdateArgs(columns []string) (args []any) {
	for _, col := range columns {
		switch col {
		case "id":
			args = append(args, e.Id)
		case "time":
			args = append(args, e.Time.AsTime())
		case "portfolio_name":
			args = append(args, e.PortfolioName)
		case "security_name":
			args = append(args, e.SecurityName)
		case "amount":
			args = append(args, e.Amount)
		case "price":
			args = append(args, e.Price)
		case "fees":
			args = append(args, e.Fees)
		case "taxes":
			args = append(args, e.Taxes)
		}
	}

	return args
}

func (*PortfolioEvent) Scan(sc persistence.Scanner) (obj persistence.StorageObject, err error) {
	var (
		e PortfolioEvent
		t time.Time
	)

	err = sc.Scan(
		&e.Id,
		&t,
		&e.PortfolioName,
		&e.SecurityName,
		&e.Amount,
		&e.Price,
		&e.Fees,
		&e.Taxes,
	)
	if err != nil {
		return nil, err
	}

	e.Time = timestamppb.New(t)

	return &e, nil
}
