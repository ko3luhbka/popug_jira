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
	Repo struct {
		db *sqlx.DB
	}
	User struct {
		ID           string    `db:"id"`
		Username     string    `db:"username"`
		Password     string    `db:"password"`
		Role         string    `db:"role"`
		Email        string    `db:"email"`
		LastModified time.Time `db:"last_modified"`
	}
)

func NewRepo(db *sqlx.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) Create(ctx context.Context, u User) (*User, error) {
	stmt, err := r.db.PrepareNamedContext(ctx,
		`
		INSERT INTO "user"(
				username,
				password,
				role,
				email,
				last_modified)
		VALUES(:username,
				:password,
				:role,
				:email,
				CURRENT_TIMESTAMP)
		RETURNING
				id,
				username,
				role,
				email,
				last_modified`,
	)
	if err != nil {
		log.Printf("failed to prepare user create query: %v\n", err)
		return nil, err
	}
	err = stmt.GetContext(ctx, &u, u)
	if err != nil {
		log.Printf("failed to create user: %v\n", err)
		return nil, err
	}
	return &u, nil
}

func (r *Repo) GetByID(ctx context.Context, uuid string) (*User, error) {
	var u User
	err := r.db.GetContext(
		ctx, &u, `
		SELECT  id,
				id,
				username,
				role,
				email,
				last_modified
		FROM "user"
		WHERE id=$1`, uuid,
	)
	if err != nil {
		log.Printf("failed to get user with uuid %s: %v\n", uuid, err)
		return nil, err
	}
	return &u, nil
}

func (r *Repo) GetAll(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.SelectContext(
		ctx, &users, `
		SELECT 	id,
				id,
				username,
				role,
				email,
				last_modified
		FROM "user"`,
	)
	if err != nil {
		log.Printf("failed to get all users: %v\n", err)
		return nil, err
	}
	return users, nil
}

func (r *Repo) Update(ctx context.Context, u User) (*User, error) {
	stmt, err := r.db.PrepareNamedContext(ctx, buildUpdateQuery(&u))
	if err != nil {
		log.Printf("failed to prepare user udpate query: %v\n", err)
		return nil, err
	}
	if err = stmt.GetContext(ctx, &u, u); err != nil {
		log.Printf("failed to update user with uuid %s: %v\n", u.ID, err)
		return nil, err
	}

	return &u, nil
}

func buildUpdateQuery(u *User) string {
	var queryBuilder strings.Builder

	queryBuilder.WriteString(`UPDATE "user" SET `)
	if u.Username != "" {
		queryBuilder.WriteString(`username=:username, `)
	}
	if u.Password != "" {
		queryBuilder.WriteString(`password=:password, `)
	}
	if u.Email != "" {
		queryBuilder.WriteString(`email=:email, `)
	}
	if u.Role != "" {
		queryBuilder.WriteString(`role=:role, `)
	}
	queryBuilder.WriteString(`last_modified=CURRENT_TIMESTAMP `)
	queryBuilder.WriteString(`WHERE id=:id `)
	queryBuilder.WriteString(`RETURNING 
		id,
		username,
		role,
		email,
		last_modified`)
	return queryBuilder.String()
}

func (r *Repo) Delete(ctx context.Context, uuid string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM "user" WHERE id=$1;`, uuid)
	if err != nil {
		log.Printf("failed to delete user with id %s: %v\n", uuid, err)
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Printf("failed to get affected rows: %v\n", err)
		return err
	}
	if affected == 0 {
		return fmt.Errorf("no user found with uuid %s", uuid)
	}

	return nil
}
