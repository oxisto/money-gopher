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
		db *persistence.DB
	)

	db = NewTestDB(t)
	ops = persistence.Ops[T](db)

	for _, init := range inits {
		init(ops)
	}

	return
}
