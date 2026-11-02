package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupDB() (*pgxpool.Pool, error) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("TEST_DATABASE_URL not provided")
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to the database: %w", err)
	}

	queryCtx, queryCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer queryCancel()
	var databaseName string
	err = pool.QueryRow(
		queryCtx,
		"SELECT current_database()",
	).Scan(&databaseName)

	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("query current database: %w", err)
	}

	if databaseName != "crudgo_test" {
		pool.Close()
		return nil, fmt.Errorf(
			"unexpected database %q, expected %q",
			databaseName,
			"crudgo_test",
		)
	}

	return pool, nil
}
