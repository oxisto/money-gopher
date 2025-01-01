package graph

import (
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/securities/quote"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB           *persistence.DB
	QuoteUpdater quote.QuoteUpdater
}

func withTx[T any](r *Resolver, f func(qtx *persistence.Queries) (*T, error)) (res *T, err error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := r.DB.WithTx(tx)
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
