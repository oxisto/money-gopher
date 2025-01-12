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

package internal

import (
	"testing"

	"github.com/oxisto/money-gopher/persistence"
)

func NewTestDB(t *testing.T, inits ...func(db *persistence.DB)) (db *persistence.DB) {
	var (
		err error
	)

	db, err = persistence.OpenDB(persistence.Options{UseInMemory: true})
	if err != nil {
		t.Fatalf("Could not create test DB: %v", err)
	}

	for _, init := range inits {
		init(db)
	}

	return
}

func NewTestDBOps[T persistence.StorageObject](t *testing.T, inits ...func(ops persistence.StorageOperations[T])) (ops persistence.StorageOperations[T]) {
	var (
		db = NewTestDB(t)
	)

	ops = persistence.Ops[T](db)

	for _, init := range inits {
		init(ops)
	}

	return
}

type errorOp[T any] struct {
	err error
}

// ErrReader creates an [persistence.StorageOperations] that returns the
// specified error in all calls.
func ErrOps[T persistence.StorageObject](err error) persistence.StorageOperations[T] {
	return &errorOp[T]{err: err}
}

func (e *errorOp[T]) Replace(o persistence.StorageObject) (err error) {
	return e.err
}

func (e *errorOp[T]) List(args ...any) (list []T, err error) {
	return nil, e.err
}

func (e *errorOp[T]) Get(key any) (obj T, err error) {
	return obj, e.err
}

func (e *errorOp[T]) Update(key any, in T, columns []string) (out T, err error) {
	return out, e.err
}

func (e *errorOp[T]) Delete(key any) (err error) {
	return e.err
}
