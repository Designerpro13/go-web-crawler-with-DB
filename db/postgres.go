package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

func PostgresInit(ctx context.Context, connStr string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Println("LOGGED HERE: PostgreSQL connection successful")
	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS word_index (
		id SERIAL PRIMARY KEY,
		word VARCHAR(255) NOT NULL,
		domain VARCHAR(255) NOT NULL,
		frequency INTEGER DEFAULT 1,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(word, domain)
	);

	CREATE INDEX IF NOT EXISTS idx_word ON word_index(word);
	CREATE INDEX IF NOT EXISTS idx_domain ON word_index(domain);
	`

	if _, err := p.DB.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("LOGGED HERE: PostgreSQL schema initialized")
	return nil
}

func (p *PostgresDB) BatchUpsertWords(ctx context.Context, postings []WordPosting) error {
	if len(postings) == 0 {
		return nil
	}

	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt := `
		INSERT INTO word_index (word, domain, frequency, last_updated)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (word, domain)
		DO UPDATE SET
			frequency = word_index.frequency + EXCLUDED.frequency,
			last_updated = NOW()
	`

	prepStmt, err := tx.PrepareContext(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer prepStmt.Close()

	for _, p := range postings {
		if _, err := prepStmt.ExecContext(ctx, p.Word, p.Domain, p.Frequency); err != nil {
			//log.Printf("LOGGED HERE: Failed to upsert word=%s domain=%s: %v", p.Word, p.Domain, err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	//log.Printf("LOGGED HERE: Batch upserted %d word postings to PostgreSQL", len(postings))
	return nil
}

func (p *PostgresDB) Close() error {
	log.Println("LOGGED HERE: Closing PostgreSQL connection")
	return p.DB.Close()
}

func ExtractDomain(urlStr string) string {
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")
	
	parts := strings.Split(urlStr, "/")
	if len(parts) > 0 {
		domain := parts[0]
		domainParts := strings.Split(domain, ".")
		if len(domainParts) >= 2 {
			return domainParts[len(domainParts)-2] + "." + domainParts[len(domainParts)-1]
		}
		return domain
	}
	return urlStr
}
