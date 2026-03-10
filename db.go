package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	log.Println("connected to database")
	createSchema()
}

func createSchema() {
	query := `
	CREATE TABLE IF NOT EXISTS pastes (
		id          TEXT PRIMARY KEY,
		content     TEXT NOT NULL,
		language    TEXT NOT NULL DEFAULT 'plaintext',
		expires_at  TIMESTAMPTZ,
		view_limit  INT,
		view_count  INT NOT NULL DEFAULT 0,
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);`

	if _, err := db.Exec(query); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	log.Println("schema ready")
}

func insertPaste(p *Paste) error {
	query := `
	INSERT INTO pastes (id, content, language, expires_at, view_limit, created_at)
	VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := db.Exec(query,
		p.ID,
		p.Content,
		p.Language,
		p.ExpiresAt,
		p.ViewLimit,
		p.CreatedAt,
	)
	return err
}

func getPaste(id string) (*Paste, error) {
	query := `
	SELECT id, content, language, expires_at, view_limit, view_count, created_at
	FROM pastes WHERE id = $1`

	p := &Paste{}
	err := db.QueryRow(query, id).Scan(
		&p.ID,
		&p.Content,
		&p.Language,
		&p.ExpiresAt,
		&p.ViewLimit,
		&p.ViewCount,
		&p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func incrementViewCount(id string) error {
	_, err := db.Exec(`UPDATE pastes SET view_count = view_count + 1 WHERE id = $1`, id)
	return err
}

func deletePaste(id string) error {
	_, err := db.Exec(`DELETE FROM pastes WHERE id = $1`, id)
	return err
}

func deleteExpiredPastes() (int64, error) {
	res, err := db.Exec(`DELETE FROM pastes WHERE expires_at IS NOT NULL AND expires_at < $1`, time.Now())
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}