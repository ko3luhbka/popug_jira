package db

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type (
	AuditRepo struct {
		db *sqlx.DB
	}
	Audit struct {
		ID         int       `db:"id"`
		EventName  string    `db:"event_name"`
		AssigneeID string    `db:"assignee_id"`
		TaskID     string    `db:"task_id"`
		TaskTitle  string    `db:"task_title"`
		JiraID     string    `db:"jira_id"`
		Amount     int       `db:"amount"`
		Created    time.Time `db:"created"`
	}
)

func NewAuditRepo(db *sqlx.DB) *AuditRepo {
	return &AuditRepo{
		db: db,
	}
}

func (r *AuditRepo) Create(ctx context.Context, aud *Audit) (*Audit, error) {
	stmt, err := r.db.PrepareNamedContext(ctx,
		`
		INSERT INTO audit (
				event_name,
				assignee_id,
				task_id,
				task_title,
				jira_id,
				amount,
				created)
		VALUES(:event_name,
				:assignee_id,
				:task_id,
				:task_title,
				:jira_id,
				:amount,
				CURRENT_TIMESTAMP)
		RETURNING
				id,
				event_name,
				assignee_id,
				task_id,
				task_title,
				jira_id,
				amount,
				created`,
	)
	if err != nil {
		log.Printf("failed to prepare audit create query: %v\n", err)
		return nil, err
	}
	err = stmt.GetContext(ctx, aud, aud)
	if err != nil {
		log.Printf("failed to create audit record: %v\n", err)
		return nil, err
	}
	return aud, nil
}

func (r *AuditRepo) GetUserAudit(ctx context.Context, uuid string) ([]Audit, error) {
	var audits []Audit
	err := r.db.SelectContext(
		ctx, &audits, `
		SELECT 	id,
				event_name,
				assignee_id,
				task_id,
				task_title,
				jira_id,
				amount,
				created
		FROM audit
		WHERE assignee_id=$1`, uuid,
	)
	if err != nil {
		log.Printf("failed to get audit records for user %s: %v\n", uuid, err)
		return nil, err
	}
	return audits, nil
}
