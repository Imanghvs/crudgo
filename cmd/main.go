package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Imanghvs/crudgo/internal/adapter/httpapi"
	"github.com/jackc/pgx/v5/pgxpool"
)

const port = 8080

func main() {
	if err := run(); err != nil {
		log.Fatalf("starting the server failed: %v", err)
	}
}

func run() error {
	log.Print("connecting to postgres")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is not provided.")
	}

	pool, err := pgxpool.New(
		context.Background(),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("cannot connect to postgres: %w", err)
	}
	defer pool.Close()

	pingCtx, pingCancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer pingCancel()
	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("pgx ping failed: %w", err)
	}

	log.Printf("Starting the server on port %d", port)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", httpapi.Health)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		return fmt.Errorf("cannot start the server: %w", err)
	}

	return nil
}
