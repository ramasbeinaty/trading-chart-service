package db

import (
	"context"
	"database/sql"
	_ "github.com/lib/pq"
	"fmt"
	"log"
	"sort"
	"time"

	"go.uber.org/zap"
)

func InitializeDB(
	cfg *DBConfigs,
) (*sql.DB, error) {
	connectionStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
		return nil, err
	}

	// verify db connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err
	}

	log.Println("Successfully connected to database")

	return db, nil
}

func RunMigrations(
	ctx context.Context,
	lgr *zap.Logger,
	db *sql.DB,
	migrations []MigrationScript,
) error {
	var entityExists bool
	var previousMigrations []migrationEntity

	dbtx, err := db.BeginTx(ctx, nil)
	if err != nil {
		panic("RunMigrations: Failed to begin a db transaction")
	}

	err = dbtx.QueryRow(queryCheckMigrationsExist).Scan(&entityExists)
	if err != nil {
		panic(fmt.Errorf("Failed to check if migrations exist - %w", err))
	}

	if !entityExists {
		lgr.Info("Creating migrations table")
		dbtx.Exec(migrationsTable.up)
		previousMigrations = []migrationEntity{}
	} else {
		lgr.Info("Fetching migrations history")

		rows, err := dbtx.Query(queryAllMigrations)
		if err != nil {
			panic(fmt.Errorf("Failed to fetch migrations history - %w", err))
		}
		defer rows.Close()

		for rows.Next() {
			var m migrationEntity
			if err := rows.Scan(&m.Index, &m.Key, &m.CreatedAt); err != nil {
				panic(fmt.Errorf("Failed to scan migration: %w", err))
			}
			previousMigrations = append(previousMigrations, m)
		}
		if err := rows.Err(); err != nil {
			panic(fmt.Errorf("Error in row iteration: %w", err))
		}
	}

	sort.Slice(previousMigrations, func(i, j int) bool {
		return previousMigrations[i].Index < previousMigrations[j].Index
	})

	previousMigrationsLength := len(previousMigrations)

	for idx, m := range migrations {
		if idx < previousMigrationsLength {
			if m.key != previousMigrations[idx].Key {
				panic(fmt.Errorf("Error: migration key is mismatched"))
			}
		} else {
			lgr.Info("Running migration", zap.String("migration", m.key))

			_, err := dbtx.Exec(m.up)
			if err != nil {
				panic(fmt.Errorf("Failed to run migration %s - %w", m.key, err))
			}

			_, err = dbtx.Exec(queryAddMigration, m.key)
			if err != nil {
				panic(fmt.Errorf("Failed to add migration %s - %w", m.key, err))
			}
		}
	}

	return dbtx.Commit()
}

type migrationEntity struct {
	Index     int        `db:"index"`
	Key       string     `db:"key"`
	CreatedAt *time.Time `db:"created_at"`
}

type MigrationScript struct {
	key  string
	up   string
	down string
}

var migrationsTable = MigrationScript{
	up: `
		SET TIMEZONE='UTC';

		CREATE TABLE migrations (
				index SERIAL,
				key text PRIMARY KEY,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
		);
	`,
	down: `
		DROP TABLE migrations;
	`,
}

const (
	queryCheckMigrationsExist = `
		SELECT EXISTS(
			SELECT * FROM pg_tables
			WHERE schemaname = 'public' AND tablename = 'migrations'
		) as exists
	`
	queryAllMigrations = `
		SELECT index, key, created_at FROM migrations
	`
	queryAddMigration = `
		INSERT INTO migrations(key) VALUES ($1)
	`
)
