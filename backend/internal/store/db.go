package store

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	DB *sql.DB
}

func NewFormDB(db *sql.DB) *Store { return &Store{DB: db} }
