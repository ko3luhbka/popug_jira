package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type (
	TaskRepo struct {
		db *sqlx.DB
	}
	Task struct {
		ID          string    `db:"id"`
		Name        string    `db:"name"`
		Description string    `db:"description"`
		Status      string    `db:"status"`
		AssigneeID  string    `db:"assignee_id"`
		Created     time.Time `db:"created"`
	}
)

func NewTaskRepo(db *sqlx.DB) *TaskRepo {
	return &TaskRepo{
		db: db,
	}
}

func (r *TaskRepo) Create(ctx context.Context, t Task) (*Task, error) {
	stmt, err := r.db.PrepareNamedContext(ctx,
		`
		INSERT INTO task(
				name,
				description,
				status,
				assignee_id,
				created)
		VALUES(:name,
				:description,
				:status,
				:assignee_id,
				CURRENT_TIMESTAMP)
		RETURNING
				id,
				name,
				description,
				status,
				assignee_id,
				created`,
	)
	if err != nil {
		log.Printf("failed to prepare task create query: %v\n", err)
		return nil, err
	}
	err = stmt.GetContext(ctx, &t, t)
	if err != nil {
		log.Printf("failed to create task: %v\n", err)
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) GetByID(ctx context.Context, uuid string) (*Task, error) {
	var t Task
	err := r.db.GetContext(
		ctx, &t, `
		SELECT  id,
				name,
				description,
				status,
				assignee_id,
				created
		FROM task
		WHERE id=$1`, uuid,
	)
	if err != nil {
		log.Printf("failed to get task with uuid %s: %v\n", uuid, err)
		return nil, err
	}
	return &t, nil
}

func (r *TaskRepo) GetAll(ctx context.Context) ([]Task, error) {
	var tasks []Task
	err := r.db.SelectContext(
		ctx, &tasks, `
		SELECT 	id,
				name,
				description,
				status,
				assignee_id,
				created
		FROM task`,
	)
	if err != nil {
		log.Printf("failed to get all tasks: %v\n", err)
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, t Task) (*Task, error) {
	stmt, err := r.db.PrepareNamedContext(ctx, buildTaskUpdateQuery(&t))
	if err != nil {
		log.Printf("failed to prepare task udpate query: %v\n", err)
		return nil, err
	}
	if err = stmt.GetContext(ctx, &t, t); err != nil {
		log.Printf("failed to update task with uuid %s: %v\n", t.ID, err)
		return nil, err
	}

	return &t, nil
}

func buildTaskUpdateQuery(t *Task) string {
	var queryBuilder strings.Builder

	queryBuilder.WriteString(`UPDATE task SET `)
	if t.Name != "" {
		queryBuilder.WriteString(`name=:name, `)
	}
	if t.Description != "" {
		queryBuilder.WriteString(`description=:description, `)
	}
	if t.Status != "" {
		queryBuilder.WriteString(`status=:status, `)
	}
	if t.AssigneeID != "" {
		queryBuilder.WriteString(`assignee_id=:assignee_id, `)
	}
	queryBuilder.WriteString(`id=:id `)
	queryBuilder.WriteString(`WHERE id=:id `)
	queryBuilder.WriteString(`RETURNING 
		id,
		name,
		description,
		status,
		assignee_id,
		created`)
	return queryBuilder.String()
}

func (r *TaskRepo) Delete(ctx context.Context, uuid string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM task WHERE id=$1;`, uuid)
	if err != nil {
		log.Printf("failed to delete task with id %s: %v\n", uuid, err)
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Printf("failed to get affected rows: %v\n", err)
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no task found with uuid %s", uuid)
	}

	return nil
}
