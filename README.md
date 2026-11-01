# CRUDGo

My reverse AI experiment to practice coding. AI tells me what to do.
It gives me the ticket. It helps me with the documentation. I do the coding!

## Prerequisites

- Go (version specified in `go.mod`)
- Docker with Docker Compose
- [Goose](https://github.com/pressly/goose) for database migrations

Install Goose if it is not already installed:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Ensure the installed `goose` executable is available in your `PATH`.

## Local setup

### 1. Configure environment variables

Create a local environment file:

```bash
cp .env.example .env
```

Review `.env` and ensure the PostgreSQL password matches across `POSTGRES_PASSWORD`, `GOOSE_DBSTRING`, `DATABASE_URL`, and `TEST_DATABASE_URL`.

The `.env` file is gitignored and must not be committed.

### 2. Start PostgreSQL

```bash
docker compose up -d
```

### 3. Export environment variables

```bash
set -a
source .env
set +a
```

Go does not automatically load `.env` files. Exporting these variables makes them available to the application, Goose, and integration tests.

### 4. Apply database migrations

```bash
goose up
```

This applies the migrations to the development database configured by `GOOSE_DBSTRING`.

### 5. Run the application

```bash
go run ./cmd
```

The application listens on port `8080`.

Check its health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:

```text
OK
```

## Integration tests

Repository integration tests use a separate PostgreSQL database named `crudgo_test`.

### 1. Create the test database

Run the following command once:

```bash
docker compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "CREATE DATABASE crudgo_test;"
```

Skip this step if `crudgo_test` already exists.

### 2. Apply migrations to the test database

Make sure environment variables have been exported:

```bash
set -a
source .env
set +a
```

Then run:

```bash
GOOSE_DBSTRING="$TEST_DATABASE_URL" goose up
```

This applies migrations to `crudgo_test` without changing the development database configuration.

### 3. Run the integration tests

```bash
go test -v -count=1 ./internal/adapter/postgres
```

To run all project tests:

```bash
go test ./...
```

### Test isolation and cleanup

The integration tests:

- Read the database connection string from `TEST_DATABASE_URL`.
- Verify that the connected database is named `crudgo_test` before writing data.
- Insert test records into PostgreSQL and verify persistence.
- Delete records created by the tests during cleanup.

Tests fail when `TEST_DATABASE_URL` is missing or points to a database with an unexpected name.

### Important notes

- The test database must exist and have migrations applied before running integration tests.
- Integration tests require a running PostgreSQL instance.
