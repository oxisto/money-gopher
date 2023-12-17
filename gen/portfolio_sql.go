// Copyright 2023 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file is part of The Money Gopher.
package portfoliov1

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/oxisto/money-gopher/persistence"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

var _ persistence.StorageObject = &Portfolio{}

func (*Portfolio) InitTables(db *persistence.DB) (err error) {
	_, err1 := db.Exec(`CREATE TABLE IF NOT EXISTS portfolios (
name TEXT PRIMARY KEY,
display_name TEXT NOT NULL
);`)
	err2 := (&PortfolioEvent{}).InitTables(db)

	return errors.Join(err1, err2)
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
name TEXT PRIMARY KEY,
type INTEGER NOT NULL,
time DATETIME NOT NULL,
portfolio_name TEXT NOT NULL, 
security_name TEXT NOT NULL,
amount REAL,
price INTEGER,
fees INTEGER,
taxes INTEGER
);`)
	if err != nil {
		return err
	}

	return
}

func (*PortfolioEvent) PrepareReplace(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`REPLACE INTO portfolio_events
(name, type, time, portfolio_name, security_name, amount, price, fees, taxes)
VALUES (?,?,?,?,?,?,?,?,?);`)
}

func (*PortfolioEvent) PrepareList(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, type, time, portfolio_name, security_name, amount, price, fees, taxes
FROM portfolio_events WHERE portfolio_name = ? ORDER BY time ASC`)
}

func (*PortfolioEvent) PrepareGet(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT * FROM portfolio_events WHERE name = ?`)
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

	query += "UPDATE portfolio_events SET " + strings.Join(set, ", ") + " WHERE name = ?;"

	return db.Prepare(query)
}

func (*PortfolioEvent) PrepareDelete(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`DELETE FROM portfolio_events WHERE name = ?`)
}

func (e *PortfolioEvent) ReplaceIntoArgs() []any {
	return []any{
		e.Name,
		e.Type,
		e.Time.AsTime(),
		e.PortfolioName,
		e.SecurityName,
		e.Amount,
		e.Price.GetValue(),
		e.Fees.GetValue(),
		e.Taxes.GetValue(),
	}
}

func (e *PortfolioEvent) UpdateArgs(columns []string) (args []any) {
	for _, col := range columns {
		switch col {
		case "name":
			args = append(args, e.Name)
		case "type":
			args = append(args, e.Type)
		case "time":
			args = append(args, e.Time.AsTime())
		case "portfolio_name":
			args = append(args, e.PortfolioName)
		case "security_name":
			args = append(args, e.SecurityName)
		case "amount":
			args = append(args, e.Amount)
		case "price":
			args = append(args, e.Price.GetValue())
		case "fees":
			args = append(args, e.Fees.GetValue())
		case "taxes":
			args = append(args, e.Taxes.GetValue())
		}
	}

	return args
}

func (*PortfolioEvent) Scan(sc persistence.Scanner) (obj persistence.StorageObject, err error) {
	var (
		e PortfolioEvent
		t time.Time
	)

	e.Price = Zero()
	e.Fees = Zero()
	e.Taxes = Zero()

	err = sc.Scan(
		&e.Name,
		&e.Type,
		&t,
		&e.PortfolioName,
		&e.SecurityName,
		&e.Amount,
		&e.Price.Value,
		&e.Fees.Value,
		&e.Taxes.Value,
	)
	if err != nil {
		return nil, err
	}

	e.Time = timestamppb.New(t)

	return &e, nil
}

func (*BankAccount) InitTables(db *persistence.DB) (err error) {
	_, err1 := db.Exec(`CREATE TABLE IF NOT EXISTS bank_accounts (
name TEXT PRIMARY KEY,
display_name TEXT NOT NULL
);`)
	err2 := (&PortfolioEvent{}).InitTables(db)

	return errors.Join(err1, err2)
}

func (*BankAccount) PrepareReplace(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`REPLACE INTO bank_accounts (name, display_name) VALUES (?,?);`)
}

func (*BankAccount) PrepareList(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, display_name FROM bank_accounts`)
}

func (*BankAccount) PrepareGet(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, display_name FROM bank_accounts WHERE name = ?`)
}

func (*BankAccount) PrepareUpdate(db *persistence.DB, columns []string) (stmt *sql.Stmt, err error) {
	// We need to make sure to quote columns here because they are potentially evil user input
	var (
		query string
		set   []string
	)

	set = make([]string, len(columns))
	for i, col := range columns {
		set[i] = persistence.Quote(col) + " = ?"
	}

	query += "UPDATE bank_accounts SET " + strings.Join(set, ", ") + " WHERE name = ?;"

	return db.Prepare(query)
}

func (*BankAccount) PrepareDelete(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`DELETE FROM bank_accounts WHERE name = ?`)
}

func (p *BankAccount) ReplaceIntoArgs() []any {
	return []any{p.Name, p.DisplayName}
}

func (p *BankAccount) UpdateArgs(columns []string) (args []any) {
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

func (*BankAccount) Scan(sc persistence.Scanner) (obj persistence.StorageObject, err error) {
	var (
		acc BankAccount
	)

	err = sc.Scan(&acc.Name, &acc.DisplayName)
	if err != nil {
		return nil, err
	}

	return &acc, nil
}
