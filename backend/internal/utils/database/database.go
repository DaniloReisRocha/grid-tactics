package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type TxFunc func(*sqlx.Tx) error
type TxFuncContext func(context.Context, *sql.Tx) error

func WithinTransaction(db *sqlx.DB, txFunc TxFunc) (err error) {

	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("unknown panic")
			}
		}
		if err != nil {
			e2 := tx.Rollback()
			if e2 != nil {
				err = errors.Wrap(err, "rollback: "+e2.Error())
			}
		}
	}()

	if err := txFunc(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func WithinTransactionContext(db *sqlx.DB, ctx context.Context, opts *sql.TxOptions, txFunc TxFuncContext) (err error) {

	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("panic recover")
		}
		if err != nil {
			e2 := tx.Rollback()
			if e2 != nil {
				err = errors.Wrap(err, "rollback: "+e2.Error())
			}
		}
	}()

	if err := txFunc(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func QueryLog(query string, args []interface{}) {
	q, e := json.Marshal(args)
	fmt.Println("\nQuery: ", query)
	fmt.Println("Args: ", string(q), e)
}
