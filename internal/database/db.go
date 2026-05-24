package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	models "example.com/inga/internal/database/models"
	"example.com/inga/internal/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Connected to database successfully")
	if err := runMigrations(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}

func runMigrations(ctx context.Context, db *sql.DB) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	migrationsDir := filepath.Join(wd, "migrations")

	driver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: postgres.DefaultMigrationsTable,
	})
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"postgres", driver)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Running migrations.....", zap.String("dir", migrationsDir))
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info(ctx, "No new migrations to apply")
			return nil
		}
		return err
	}
	v, dirty, _ := m.Version()
	logger.Info(ctx, "Migrations applied successfully", zap.Uint("version", v), zap.Bool("dirty", dirty))
	return nil
}

type Querier interface {
	models.Querier
	WithTx(tx *sql.Tx) Querier
	DB() *sql.DB
}

type Queries struct {
	*models.Queries
	db *sql.DB
}

func (q *Queries) DB() *sql.DB {
	return q.db
}

func (q *Queries) WithTx(tx *sql.Tx) Querier {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
	}
}

func NewQuerier(q *models.Queries, db *sql.DB) Querier {
	return &Queries{
		Queries: q,
		db:      db,
	}
}
