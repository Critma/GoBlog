package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/critma/goblog/internal/store"
	"github.com/go-chi/chi/v5"
)

type articleKey string

const articleCtx articleKey = "article"

type CreateArticlePayload struct {
	Title   string `json:"title" validate:"required,max=100"`
	Content string `json:"content" validate:"required,max=100"`
}

func (app *application) getLatestArticlesHandler(w http.ResponseWriter, r *http.Request) {
	latests, err := app.store.Articles.GetLastTen(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusOK, latests); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getArticleByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Articles.GetByID(r.Context(), int(id))
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getArticlesByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}
	pq := store.PaginatedQuery{}
	if pq, err = pq.Parse(r); err != nil {
		app.badRequestResponse(w, r, err)
	}

	articles, err := app.store.Articles.GetByAuthor(r.Context(), int(userID), pq)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}
	if err := app.jsonResponse(w, http.StatusOK, articles); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) createArticleHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateArticlePayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	article := &store.Article{
		Title:    payload.Title,
		Content:  payload.Content,
		AuthorID: user.ID,
	}

	ctx := r.Context()
	id, err := app.store.Articles.Create(ctx, article)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	article.ID = id
	if err := app.jsonResponse(w, http.StatusCreated, article); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type UpdateArticlePayload struct {
	Title   string `json:"title" validate:"omitempty,max=100"`
	Content string `json:"content" validate:"omitempty,max=1000"`
}

func (app *application) updateArticleHandler(w http.ResponseWriter, r *http.Request) {
	article := getArticleFromCtx(r)

	var payload UpdateArticlePayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Title != "" {
		article.Title = payload.Title
	}
	if payload.Content != "" {
		article.Content = payload.Content
	}

	ctx := r.Context()
	id, err := app.store.Articles.Update(ctx, article)
	app.logger.Infow("info", "art", article)
	if err != nil {
		app.internalServerError(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusOK, id); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) deleteArticleHandler(w http.ResponseWriter, r *http.Request) {
	article := getArticleFromCtx(r)

	ctx := r.Context()

	if err := app.store.Articles.Delete(ctx, article.ID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) getArticleCommentsHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	pq := store.PaginatedQuery{}
	pq, err = pq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if comments, err := app.store.Articles.GetComments(r.Context(), int(id), pq); err != nil {
		app.internalServerError(w, r, err)
		return
	} else {
		app.jsonResponse(w, http.StatusOK, comments)
		return
	}
}

func (app *application) createArticleCommentHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	comm := &store.Comment{}

	err = readJSON(w, r, comm)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	comm.ArticleID = int(id)
	user := getUserFromContext(r)
	comm.UserID = user.ID

	commID, err := app.store.Articles.AddComment(r.Context(), comm)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	app.jsonResponse(w, http.StatusCreated, commID)
}

func (app *application) createLikeOnArticle(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	if err := app.store.Articles.AddLike(r.Context(), int(id), user.ID); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) articleContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		ctx := r.Context()

		article, err := app.store.Articles.GetByID(ctx, int(id))
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		ctx = context.WithValue(ctx, articleCtx, article)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getArticleFromCtx(r *http.Request) *store.Article {
	art, _ := r.Context().Value(articleCtx).(*store.Article)
	return art
}
