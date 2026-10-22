package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

// A transfer between two own accounts is one event with two legs: the
// money leaves one account as a WITHDRAW_CASH and reaches the other as a
// DEPOSIT_CASH, both carrying the same transfer ID. Every balance and the
// net worth see two ordinary transactions; only the two of them together
// say "moved", not "spent" or "earned".

// CreateTransferInput is the input for Mutation.createTransfer.
type CreateTransferInput struct {
	Time              graphql.Time
	FromCashAccountID graphql.ID
	ToCashAccountID   graphql.ID
	Amount            int32
	Currency          string
}

// TransferResolver resolves the Transfer GraphQL type from its two legs.
type TransferResolver struct {
	db       *persistence.DB
	from, to *persistence.Transaction
}

// CreateTransfer resolves Mutation.createTransfer.
func (r *RootResolver) CreateTransfer(ctx context.Context, args struct{ Input CreateTransferInput }) (*TransferResolver, error) {
	person, err := auth.PersonFromContext(ctx)
	if err != nil {
		return nil, err
	}
	in := args.Input

	if in.FromCashAccountID == in.ToCashAccountID {
		return nil, errors.New("a transfer needs two different accounts")
	}
	if in.Amount <= 0 {
		return nil, errors.New("the amount must be positive")
	}
	from := string(in.FromCashAccountID)
	to := string(in.ToCashAccountID)
	for _, id := range []string{from, to} {
		account, err := r.db.GetCashAccount(ctx, persistence.GetCashAccountParams{ID: id, PersonID: person.ID})
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cash account %q not found", id)
		} else if err != nil {
			return nil, err
		}
		if account.Currency != in.Currency {
			return nil, fmt.Errorf("%s is in %s, not %s", account.DisplayName, account.Currency, in.Currency)
		}
	}

	legs := [2]txData{
		{Type: "WITHDRAW_CASH", Time: in.Time.Time.UTC(), CashAccountID: &from, CashDelta: -int64(in.Amount), Currency: in.Currency},
		{Type: "DEPOSIT_CASH", Time: in.Time.Time.UTC(), CashAccountID: &to, CashDelta: int64(in.Amount), Currency: in.Currency},
	}
	for i := range legs {
		if err := legs[i].validate(ctx, r.db, person); err != nil {
			return nil, err
		}
	}

	transferID := sql.NullString{String: uuid.NewString(), Valid: true}
	var created [2]*persistence.Transaction
	err = r.db.Tx(ctx, func(q *persistence.Queries) error {
		for i, d := range legs {
			tx, err := q.CreateTransaction(ctx, persistence.CreateTransactionParams{
				ID:            uuid.NewString(),
				Type:          d.Type,
				Time:          d.Time,
				CashAccountID: nullString(d.CashAccountID),
				CashDelta:     d.CashDelta,
				Currency:      d.Currency,
				Source:        "MANUAL",
				TransferID:    transferID,
			})
			if err != nil {
				return err
			}
			created[i] = tx
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &TransferResolver{db: r.db, from: created[0], to: created[1]}, nil
}

func (r *TransferResolver) ID() graphql.ID { return graphql.ID(r.from.TransferID.String) }
func (r *TransferResolver) Time() graphql.Time {
	return graphql.Time{Time: r.from.Time}
}
func (r *TransferResolver) Amount() Money {
	return money(r.to.CashDelta, r.to.Currency)
}
func (r *TransferResolver) From() *TransactionResolver {
	return &TransactionResolver{db: r.db, transaction: r.from}
}
func (r *TransferResolver) To() *TransactionResolver {
	return &TransactionResolver{db: r.db, transaction: r.to}
}

// TransferID resolves Transaction.transferID.
func (r *TransactionResolver) TransferID() *graphql.ID {
	if !r.transaction.TransferID.Valid {
		return nil
	}
	id := graphql.ID(r.transaction.TransferID.String)
	return &id
}

// Counterpart resolves Transaction.counterpart: the other leg of a transfer.
func (r *TransactionResolver) Counterpart(ctx context.Context) (*TransactionResolver, error) {
	if !r.transaction.TransferID.Valid {
		return nil, nil
	}
	other, err := r.db.GetTransferCounterpart(ctx, persistence.GetTransferCounterpartParams{
		TransferID: r.transaction.TransferID,
		ID:         r.transaction.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &TransactionResolver{db: r.db, transaction: other}, nil
}
