package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/critma/goblog/internal/store"
)

type ArticleStore struct {
	db *sql.DB
}

func (s *ArticleStore) GetLastTen(ctx context.Context) ([]*store.LatestArticle, error) {
	query := `
	SELECT * from latest_articles
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, store.ErrNotFound
		default:
			return nil, err
		}
	}
	var result []*store.LatestArticle
	for rows.Next() {
		art := &store.LatestArticle{}
		rows.Scan(
			&art.ID,
			&art.Title,
			&art.AuthorName,
			&art.Likes,
			&art.PublishedAt,
		)
		result = append(result, art)
	}
	return result, nil
}

// with author
func (s *ArticleStore) GetByID(ctx context.Context, id int) (*store.Article, error) {
	query := `
	SELECT 
		articles.*,
		users.id,
		users.username,
		users.email
	FROM articles
	JOIN users ON users.id = articles.author_id
	WHERE articles.id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	art := &store.Article{}

	if err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&art.ID,
		&art.Title,
		&art.Content,
		&art.AuthorID,
		&art.Likes,
		&art.PublishedAt,
		&art.UpdatedAt,

		&art.User.ID,
		&art.User.Username,
		&art.User.Email,
	); err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, store.ErrNotFound
		default:
			return nil, err
		}
	}
	return art, nil
}

// with count of likes
func (s *ArticleStore) GetByAuthor(ctx context.Context, UserId int, pq store.PaginatedQuery) ([]*store.Article, error) {
	query := `
		SELECT *
		FROM articles
		WHERE author_id = $1
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, UserId, pq.Limit, pq.Offset)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, store.ErrNotFound
		default:
			return nil, err
		}
	}

	result := make([]*store.Article, 0)
	for rows.Next() {
		art := &store.Article{}
		if err := rows.Scan(
			&art.ID,
			&art.Title,
			&art.Content,
			&art.AuthorID,
			&art.Likes,
			&art.PublishedAt,
			&art.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, art)
	}
	return result, nil
}

func (s *ArticleStore) Create(ctx context.Context, article *store.Article) (int, error) {
	if article.AuthorID == 0 {
		return 0, errors.New("author id is required")
	}
	query := `
		INSERT INTO articles (title, content, author_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	var id int
	if err := s.db.QueryRowContext(ctx, query, article.Title, article.Content, article.AuthorID).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *ArticleStore) Update(ctx context.Context, article *store.Article) (int, error) {
	if article.AuthorID == 0 {
		return 0, errors.New("author id is required")
	}
	if article.ID == 0 {
		return 0, errors.New("article id is required")
	}

	query := `
		UPDATE articles
		SET title = $1, content = $2
		WHERE articles.id = $3
		RETURNING id
	`

	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	var id int
	if err := s.db.QueryRowContext(ctx, query, article.Title, article.Content, article.ID).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *ArticleStore) Delete(ctx context.Context, id int) error {
	query := `
	DELETE FROM articles WHERE id = $1
	`

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return store.ErrNotFound
	}

	return nil
}

func (s *ArticleStore) GetComments(ctx context.Context, articleID int, pq store.PaginatedQuery) ([]*store.Comment, error) {
	query := `
	SELECT * FROM comments WHERE article_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, articleID, pq.Limit, pq.Offset)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, store.ErrNotFound
		default:
			return nil, err
		}
	}
	result := make([]*store.Comment, 0)
	for rows.Next() {
		comm := &store.Comment{}
		if err := rows.Scan(
			&comm.ID,
			&comm.ArticleID,
			&comm.UserID,
			&comm.Text,
			&comm.CreatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, comm)
	}
	return result, nil
}

func (s *ArticleStore) AddComment(ctx context.Context, comment *store.Comment) (int, error) {
	if comment.UserID == 0 || comment.ArticleID == 0 {
		return 0, errors.New("user or article id is required")
	}

	query := `
		INSERT INTO comments (article_id, user_id, text) VALUES ($1, $2, $3) RETURNING id
	`

	ctx, cancel := context.WithTimeout(ctx, store.QueryTimeoutDuration)
	defer cancel()

	commentID := 0
	if err := s.db.QueryRowContext(ctx, query, comment.ArticleID, comment.UserID, comment.Text).Scan(&comment.ID); err != nil {
		return 0, err
	}
	return commentID, nil
}

func (s *ArticleStore) AddLike(ctx context.Context, articleID, userID int) error {
	if articleID == 0 || userID == 0 {
		return errors.New("user or article id is required")
	}

	query := `
		INSERT INTO article_like (article_id, user_id) VALUES ($1, $2)
	`

	if _, err := s.db.ExecContext(ctx, query, articleID, userID); err != nil {
		switch err {
		case sql.ErrNoRows:
			return store.ErrNotFound
		default:
			return err
		}
	}

	return nil
}
