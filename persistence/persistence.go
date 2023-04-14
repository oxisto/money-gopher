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
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
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

// OpenDB opens a connection to our database.
func OpenDB(opts Options) (db *DB, err error) {
	if opts.UseInMemory {
		opts.DSN = ":memory:?_pragma=foreign_keys(1)"
	} else if opts.DSN == "" {
		opts.DSN = "money.db"
	}

	inner, err := sql.Open("sqlite", opts.DSN)
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
