package store

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("res not found")
	ErrExists            = errors.New("res already exists")
	QueryTimeoutDuration = time.Second * 10
)

type Storage struct {
	Users interface {
		GetByID(context.Context, int) (*User, error)
		GetByEmail(ctx context.Context, email string) (*User, error)
		Create(context.Context, *User) error
	}
	Articles interface {
		GetLastTen(context.Context) ([]*LatestArticle, error)
		GetByID(context.Context, int) (*Article, error)
		GetByAuthor(ctx context.Context, UserId int, pq PaginatedQuery) ([]*Article, error)
		Create(ctx context.Context, article *Article) (int, error)
		Update(ctx context.Context, article *Article) (int, error)
		Delete(ctx context.Context, id int) error
		GetComments(ctx context.Context, articleID int, pq PaginatedQuery) ([]*Comment, error)
		AddComment(ctx context.Context, comment *Comment) (int, error)
		// DeleteComment(ctx context.Context, id int) error
		AddLike(ctx context.Context, articleID, userID int) error
	}
}
