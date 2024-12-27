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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/oxisto/money-gopher/persistence/sql/migrations"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
)

// Options are database options
type Options struct {
	// UseInMemory forces our persistence layer to use an in-memory sqlite database
	UseInMemory bool

	// DSN contains the DSN, such as the file name of our sqlite database
	DSN string
}

// LogValue implements slog.LogValuer.
func (o Options) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Bool("in-memory", o.UseInMemory),
		slog.String("dsn", o.DSN))
}

// DB is a type alias around [sql.DB] to avoid importing the [database/sql] package.
type DB = sql.DB

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
	List(args ...any) (list []T, err error)
	Get(key any) (obj T, err error)
	Update(key any, in T, columns []string) (out T, err error)
	Delete(key any) (err error)
}

type Scanner interface {
	Scan(args ...any) error
}

type ops[T StorageObject] struct {
	*DB
}

// OpenDB opens a connection to our database.
func OpenDB(opts Options) (db *DB, q *Queries, err error) {
	if opts.UseInMemory {
		opts.DSN = ":memory:?_pragma=foreign_keys(1)"
	} else if opts.DSN == "" {
		opts.DSN = "money.db"
	}

	db, err = sql.Open("sqlite3", opts.DSN)
	if err != nil {
		return nil, nil, fmt.Errorf("could not open database: %w", err)
	}

	slog.Info("Successfully opened database connection", "opts", opts)

	// Prepare database migrations with goose
	provider, err := goose.NewProvider(database.DialectSQLite3, db, migrations.Embed)
	if err != nil {
		log.Fatal(err)
	}

	// Apply all migrations
	results, err := provider.Up(context.Background())
	if err != nil {
		return nil, nil, err
	}

	for _, result := range results {
		slog.Debug("Applied migration.", "migration", result)
	}

	// Create a new query object
	q = New(db)

	return
}

func Ops[T StorageObject](db *DB) StorageOperations[T] {
	return &ops[T]{DB: db}
}

func Relationship[T StorageObject, S StorageObject](op StorageOperations[S]) StorageOperations[T] {
	return &ops[T]{DB: op.(*ops[S]).DB}
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

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not fetch number of affected rows: %w", err)
	}

	return nil
}

// List lists stuff. Because methods cannot have type parameters (unless the
// struct has one), we need to make this a function (for now) and pass the [DB]
// as the first parameter.
func (ops *ops[T]) List(args ...any) (list []T, err error) {
	// We need to construct a pointer to the underlying type that fulfills T in
	// order to access some of the database functions.
	var t T

	list = make([]T, 0)

	// TODO(oxisto): prepare query at DB init
	stmt, err := t.PrepareList(ops.DB)
	if err != nil {
		return nil, fmt.Errorf("could not prepare query: %w", err)
	}

	rows, err := stmt.Query(args...)
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
		row  *sql.Row
		tmp  StorageObject
		args []any
		ok   bool
	)

	// TODO(oxisto): prepare query at DB init
	stmt, err := obj.PrepareGet(ops.DB)
	if err != nil {
		return obj, fmt.Errorf("could not prepare query: %w", err)
	}

	// split composite keys
	if args, ok = key.([]any); !ok {
		// otherwise, use single key
		args = []any{key}
	}

	row = stmt.QueryRow(args...)
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
		keys []any
		ok   bool
	)
	// TODO(oxisto): cache somehow
	stmt, err := in.PrepareUpdate(ops.DB, columns)
	if err != nil {
		return out, fmt.Errorf("could not prepare query: %w", err)
	}

	args = make([]any, 0, 1+len(columns))
	args = append(args, in.UpdateArgs(columns)...)

	if keys, ok = key.([]any); ok {
		args = append(args, keys...)
	} else {
		args = append(args, key)
	}

	res, err := stmt.Exec(args...)
	if err != nil {
		return out, fmt.Errorf("could not execute query: %w", err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return out, fmt.Errorf("could not fetch number of affected rows: %w", err)
	}

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

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not fetch number of affected rows: %w", err)
	}

	return
}

func Quote(in string) string {
	in = strings.ReplaceAll(in, string([]byte{0}), "")
	in = "\"" + strings.ReplaceAll(in, "\"", "\"\"") + "\""
	return in
}
