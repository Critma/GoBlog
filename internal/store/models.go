package store

import (
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int      `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Password  password `json:"-"`
	CreatedAt string   `json:"created_at,omitempty"`
}

type password struct {
	Text *string
	Hash []byte
}

func (p *password) Set(text string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(text), 12)
	if err != nil {
		return err
	}

	p.Text = &text
	p.Hash = hash

	return nil
}

func (p *password) CompareWithHash(text string) error {
	return bcrypt.CompareHashAndPassword(p.Hash, []byte(text))
}

type Article struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	AuthorID    int       `json:"author_id"`
	Likes       int       `json:"likes"`
	PublishedAt time.Time `json:"published_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	User User `json:"user"`
}

type LatestArticle struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	AuthorName  string    `json:"author_name"`
	Likes       int       `json:"likes"`
	PublishedAt time.Time `json:"published_at"`
}

type PaginatedQuery struct {
	Limit  int    `json:"limit" validate:"gte=1,lte=10"`
	Offset int    `json:"offset" validate:"gte=0"`
	Search string `json:"search" validate:"max=90"`
}

func (pq PaginatedQuery) Parse(r *http.Request) (PaginatedQuery, error) {
	q := r.URL.Query()

	limit := q.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return pq, err
		}
		pq.Limit = l
	}

	offset := q.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return pq, err
		}
		pq.Offset = o
	}

	search := q.Get("search")
	if search != "" {
		pq.Search = search
	}
	return pq, nil
}

type Comment struct {
	ID        int    `json:"id"`
	ArticleID int    `json:"article_id"`
	UserID    int    `json:"user_id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}
