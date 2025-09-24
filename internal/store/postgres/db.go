package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/critma/goblog/internal/store"
)

func NewConnection(addr string, maxOpenConns, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)

	duration, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func NewStorage(db *sql.DB) store.Storage {
	return store.Storage{
		Users:    &UserStore{db},
		Articles: &ArticleStore{db},
	}
}
