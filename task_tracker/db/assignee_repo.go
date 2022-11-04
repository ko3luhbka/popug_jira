package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
)

type (
	AssigneeRepo struct {
		db *sqlx.DB
	}
	Assignee struct {
		ID       string `db:"id"`
		Username string `db:"username"`
		Role string `db:"role"`
	}
)

func NewAssigneeepo(db *sqlx.DB) *AssigneeRepo {
	return &AssigneeRepo{
		db: db,
	}
}

func (r *AssigneeRepo) Create(ctx context.Context, a Assignee) (*Assignee, error) {
	stmt, err := r.db.PrepareNamedContext(ctx,
		`
		INSERT INTO assignee (
				id,
				username)
		VALUES(:id,
				:username)
		RETURNING
				id,
				username`,
	)
	if err != nil {
		log.Printf("failed to prepare assignee create query: %v\n", err)
		return nil, err
	}
	err = stmt.GetContext(ctx, &a, a)
	if err != nil {
		log.Printf("failed to create assignee: %v\n", err)
		return nil, err
	}
	return &a, nil
}

func (r *AssigneeRepo) GetAll(ctx context.Context) ([]Assignee, error) {
	var assignees []Assignee
	err := r.db.SelectContext(
		ctx, &assignees, `
		SELECT 	id,
				username
		FROM assignee`,
	)
	if err != nil {
		log.Printf("failed to get all users: %v\n", err)
		return nil, err
	}
	return assignees, nil
}

func (r *AssigneeRepo) Update(ctx context.Context, a Assignee) (*Assignee, error) {
	stmt, err := r.db.PrepareNamedContext(ctx, buildUpdateQuery(&a))
	if err != nil {
		log.Printf("failed to prepare assignee udpate query: %v\n", err)
		return nil, err
	}
	if err = stmt.GetContext(ctx, &a, a); err != nil {
		log.Printf("failed to update assignee with uuid %s: %v\n", a.ID, err)
		return nil, err
	}

	return &a, nil
}

func buildUpdateQuery(a *Assignee) string {
	var queryBuilder strings.Builder

	queryBuilder.WriteString(`UPDATE assignee SET `)
	if a.Username != "" {
		queryBuilder.WriteString(`username=:username, `)
	}
	queryBuilder.WriteString(`id=:id `)
	queryBuilder.WriteString(`WHERE id=:id `)
	queryBuilder.WriteString(`RETURNING
		id,
		username`)
	return queryBuilder.String()
}

func (r *AssigneeRepo) Delete(ctx context.Context, uuid string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM assignee WHERE id=$1;`, uuid)
	if err != nil {
		log.Printf("failed to delete assignee with id %s: %v\n", uuid, err)
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Printf("failed to get affected rows: %v\n", err)
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no assignee found with uuid %s", uuid)
	}

	return nil
}
