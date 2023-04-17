// Copyright 2023 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file is part of The Money Gopher.

// package persistence contains our storage layer.
package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Options are database options
type Options struct {
	// UseInMemory forces our persistence layer to use an in-memory sqlite database
	UseInMemory bool

	// DSN contains the DSN, such as the file name of our sqlite database
	DSN string
}

// DB is a wrapper around [sql.DB]. This allows us to access all the
// functionalities of [sql.DB] as well as accessing the DB object in our
// internal functions.
type DB struct {
	*sql.DB

	log *log.Logger
}

type StorageObject interface {
	InitTables(db *DB) (err error)
	PrepareReplace(db *DB) (stmt *sql.Stmt, err error)
	PrepareList(db *DB) (stmt *sql.Stmt, err error)
	PrepareGet(db *DB) (stmt *sql.Stmt, err error)
	PrepareUpdate(db *DB, columns []string) (stmt *sql.Stmt, err error)
	PrepareDelete(db *DB) (stmt *sql.Stmt, err error)
	ReplaceIntoArgs() []any
	UpdateArgs([]string) []any
	Scan(sc Scanner) (StorageObject, error)
}

type StorageOperations[T StorageObject] interface {
	Replace(o StorageObject) (err error)
	List() (list []T, err error)
	Get(key any) (obj T, err error)
	Update(key any, in T, columns []string) (out T, err error)
	Delete(key any) error
}

type Scanner interface {
	Scan(args ...any) error
}

type ops[T StorageObject] struct {
	*DB
}

// OpenDB opens a connection to our database.
func OpenDB(opts Options) (db *DB, err error) {
	if opts.UseInMemory {
		opts.DSN = ":memory:?_pragma=foreign_keys(1)"
	} else if opts.DSN == "" {
		opts.DSN = "money.db"
	}

	inner, err := sql.Open("sqlite3", opts.DSN)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	db = &DB{
		DB:  inner,
		log: log.New(os.Stderr, "", log.Lmsgprefix|log.Ltime),
	}
	db.log.SetPrefix("[📄] ")
	db.initTables()

	db.log.Print("Successfully opened database connection")

	return
}

func Ops[T StorageObject](db *DB) StorageOperations[T] {
	return &ops[T]{DB: db}
}

func (ops *ops[T]) Replace(o StorageObject) (err error) {
	// TODO(oxisto): Move to db.initTables
	err = o.InitTables(ops.DB)
	if err != nil {
		return fmt.Errorf("could not init table: %w", err)
	}

	// TODO: "prepare" it somewhere else
	stmt, err := o.PrepareReplace(ops.DB)
	if err != nil {
		return fmt.Errorf("could not prepare query: %w", err)
	}

	res, err := stmt.Exec(o.ReplaceIntoArgs()...)
	if err != nil {
		return fmt.Errorf("could not execute query: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not fetch number of affected rows: %w", err)
	}
	ops.DB.log.Printf("%d row(s) affected by replace", rows)

	return nil
}

// List lists stuff. Because methods cannot have type parameters (unless the
// struct has one), we need to make this a function (for now) and pass the [DB]
// as the first parameter.
func (ops *ops[T]) List() (list []T, err error) {
	// We need to construct a pointer to the underlying type that fulfills T in
	// order to access some of the database functions.
	var t T

	list = make([]T, 0)

	// TODO(oxisto): prepare query at DB init
	stmt, err := t.PrepareList(ops.DB)
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("could not execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var obj StorageObject
		obj, err = t.Scan(rows)
		if err != nil {
			return nil, fmt.Errorf("could not scan object: %w", err)
		}

		list = append(list, obj.(T))
	}

	return
}

func (ops *ops[T]) Get(key any) (obj T, err error) {
	var (
		row *sql.Row
		tmp StorageObject
	)

	// TODO(oxisto): prepare query at DB init
	stmt, err := obj.PrepareGet(ops.DB)
	if err != nil {
		return obj, fmt.Errorf("could not prepare query: %w", err)
	}

	row = stmt.QueryRow(key)
	tmp, err = obj.Scan(row)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return obj, nil
	} else if err != nil {
		return obj, fmt.Errorf("could not scan object: %w", err)
	}

	obj = tmp.(T)

	return
}

func (ops *ops[T]) Update(key any, in T, columns []string) (out T, err error) {
	var (
		args []any
	)
	// TODO(oxisto): cache somehow
	stmt, err := in.PrepareUpdate(ops.DB, columns)
	if err != nil {
		return out, fmt.Errorf("could not prepare query: %w", err)
	}

	args = make([]any, 0, 1+len(columns))
	args = append(args, in.UpdateArgs(columns)...)
	args = append(args, key)

	res, err := stmt.Exec(args...)
	if err != nil {
		return out, fmt.Errorf("could not execute query: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return out, fmt.Errorf("could not fetch number of affected rows: %w", err)
	}

	ops.DB.log.Printf("%d row(s) affected by replace", rows)

	// Need to fetch it again
	return ops.Get(key)
}

func (ops *ops[T]) Delete(key any) (err error) {
	var (
		t T
	)

	// TODO(oxisto): cache somehow
	stmt, err := t.PrepareDelete(ops.DB)
	if err != nil {
		return fmt.Errorf("could not prepare query: %w", err)
	}

	res, err := stmt.Exec(key)
	if err != nil {
		return fmt.Errorf("could not execute query: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not fetch number of affected rows: %w", err)
	}

	ops.DB.log.Printf("%d row(s) affected by delete", rows)

	return
}

func Quote(in string) string {
	in = strings.ReplaceAll(in, string([]byte{0}), "")
	in = "\"" + strings.ReplaceAll(in, "\"", "\"\"") + "\""
	return in
}
