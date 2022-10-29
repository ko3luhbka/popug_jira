package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type (
	AccountRepo struct {
		db *sqlx.DB
	}
	Account struct {
		ID         int       `db:"id"`
		AssigneeID string    `db:"assignee_id"`
		Credit     int       `db:"credit"`
		Debit      int       `db:"debit"`
		Created    time.Time `db:"created"`
	}
	AssigneeBalance struct {
		AssigneeID string `db:"assignee_id"`
		Balance    int    `db:"balance"`
	}
)

func NewAccountRepo(db *sqlx.DB) *AccountRepo {
	return &AccountRepo{
		db: db,
	}
}

func (r *AccountRepo) CreateRecord(ctx context.Context, acc *Account) (*Account, error) {
	stmt, err := r.db.PrepareNamedContext(ctx,
		`
		INSERT INTO account(
				assignee_id,
				credit,
				debit,
				created)
		VALUES(:assignee_id,
				:credit,
				:debit,
				CURRENT_TIMESTAMP)
		RETURNING
				id,
				assignee_id,
				credit,
				debit,
				created`,
	)
	if err != nil {
		log.Printf("failed to prepare account create query: %v\n", err)
		return nil, err
	}
	err = stmt.GetContext(ctx, acc, acc)
	if err != nil {
		log.Printf("failed to create account: %v\n", err)
		return nil, err
	}
	return acc, nil
}

func (r *AccountRepo) GetUserBalance(ctx context.Context, uuid string) (balance int, err error) {
	// TODO: check that records with such uuid are actually extist in the table
	err = r.db.GetContext(
		ctx, &balance, `
		SELECT coalesce(SUM(credit) + SUM(debit), 0)
		FROM account
		WHERE assignee_id=$1`, uuid,
	)
	if err != nil {
		log.Printf("failed to get balance of user %s: %v\n", uuid, err)
		return 0, err
	}
	return balance, nil
}

func (r *AccountRepo) GetUsersBalances(ctx context.Context) ([]AssigneeBalance, error) {
	var balances []AssigneeBalance
	err := r.db.SelectContext(
		ctx, &balances,
		`SELECT assignee_id, coalesce(SUM(credit) + SUM(debit), 0) AS balance
		FROM account
		WHERE created::date = now()::date
		GROUP BY assignee_id`,
	)
	if err != nil {
		log.Printf("failed to get users' balances: %v\n", err)
		return nil, err
	}
	return balances, nil
}

func (r *AccountRepo) GetManagementTodayIncome(ctx context.Context) (income int, err error) {
	err = r.db.GetContext(
		ctx, &income, `
		SELECT 	coalesce(-(SUM(credit) + SUM(debit)), 0)
		FROM account
		WHERE created::date = now()::date`,
	)
	if err != nil {
		log.Printf("failed to get management income: %v\n", err)
		return 0, err
	}
	return income, nil
}

func (r *AccountRepo) DeleteByUser(ctx context.Context, uuid string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM account WHERE assignee_id=$1;`, uuid)
	if err != nil {
		log.Printf("failed to delete account rows with assignee_id %s: %v\n", uuid, err)
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Printf("failed to get affected rows: %v\n", err)
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no account rows found with assignee_id %s", uuid)
	}

	return nil
}
