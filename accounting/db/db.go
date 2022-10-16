package db

import (
	"database/sql"
	"fmt"
	"io/fs"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

const (
	dbUser = "postgres"
	dbPass = "mysecretpassword"
	dbName = "postgres"
	dbHost = "127.0.0.1"
	// dbHost              = "172.17.0.2"
	// dbHost              = "db"
	dbPort              = 5432
	driverName          = "pgx"
	dsnFormat           = "postgres://%s:%s@%s:%d/%s?search_path=accounting"
	migrationsDirectory = "."
)

func NewConnection() (*sqlx.DB, error) {
	dsn := fmt.Sprintf(dsnFormat, dbUser, dbPass, dbHost, dbPort, dbName)
	conn, err := sqlx.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}
	return conn, nil
}

func RunMigrations(conn *sql.DB, migrationFiles fs.FS) error {
	goose.SetBaseFS(migrationFiles)
	if err := goose.Up(conn, migrationsDirectory); err != nil {
		return fmt.Errorf("failed to perform migrations: %v", err)
	}
	return nil
}
