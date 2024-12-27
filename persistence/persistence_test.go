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

package persistence

import (
	"path"
	"testing"

	"github.com/mattn/go-sqlite3"
)

func TestOpenDB(t *testing.T) {
	type args struct {
		opts Options
	}
	tests := []struct {
		name    string
		args    args
		wantDb  func(t *testing.T, db *DB) bool
		wantErr bool
	}{
		{
			name: "Happy path with in-memory",
			args: args{Options{UseInMemory: true}},
			wantDb: func(t *testing.T, db *DB) bool {
				_, ok := db.Driver().(*sqlite3.SQLiteDriver)
				return ok
			},
			wantErr: false,
		},
		{
			name: "Happy path with in-memory",
			args: args{Options{DSN: path.Join(t.TempDir(), "money.db")}},
			wantDb: func(t *testing.T, db *DB) bool {
				_, ok := db.Driver().(*sqlite3.SQLiteDriver)
				return ok
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDb, err := OpenDB(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantDb(t, gotDb) {
				t.Errorf("OpenDB() = %v, not expected", gotDb)
			}
		})
	}
}
