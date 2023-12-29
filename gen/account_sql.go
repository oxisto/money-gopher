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

	"github.com/oxisto/money-gopher/persistence"
)

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
