package graph

import (
	"database/sql"

	"github.com/oxisto/money-gopher/db"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	Queries *db.Queries
	DB      *sql.DB
}

func withTx[T any](r *Resolver, f func(qtx *db.Queries) (*T, error)) (res *T, err error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := r.Queries.WithTx(tx)
	res, err = f(qtx)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return res, nil
}
