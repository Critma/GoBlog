package postgres

import (
	"context"
	"database/sql"

	"github.com/critma/goblog/internal/store"
)

type UserStore struct {
	db *sql.DB
}

func (s *UserStore) GetByID(ctx context.Context, id int) (*store.User, error) {
	query := `
	SELECT * FROM users WHERE id = $1
	`
	user := &store.User{}
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Password.Hash,
		&user.Email,
		&user.CreatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, store.ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*store.User, error) {
	query := `
	SELECT * FROM users WHERE email = $1
	`
	user := &store.User{}

	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Password.Hash,
		&user.Email,
		&user.CreatedAt)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, store.ErrNotFound
		default:
			return nil, err
		}
	}
	return user, nil
}

func (s *UserStore) Create(ctx context.Context, user *store.User) error {
	query := `
	INSERT INTO users (username, password_hash, email)
	VALUES ($1, $2, $3) RETURNING id, created_at
	`

	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Password.Hash,
		user.Email,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
