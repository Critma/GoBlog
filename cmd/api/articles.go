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

// @Summary		Get latest articles
// @Description	Get latest articles
// @Tags			articles
// @Accept			json
// @Produce		json
// @Success		200	{object}	[]store.LatestArticle
// @Failure		400	{object}	error
// @Failure		500	{object}	error
// @Router			/articles [get]
func (app *application) getLatestArticlesHandler(w http.ResponseWriter, r *http.Request) {
	latests, err := app.store.Articles.GetLastTen(r.Context())
	if err != nil {
		app.internalServerError(w, r, err)
	}

	if err := app.jsonResponse(w, http.StatusOK, latests); err != nil {
		app.internalServerError(w, r, err)
	}
}

// @Summary		Get article by id
// @Description	Get article by id
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"Article ID"
// @Success		200	{object}	store.Article
// @Failure		400	{object}	error
// @Failure		404	{object}	error
// @Failure		500	{object}	error
// @Security		ApiKeyAuth
// @Router			/articles/{id} [get]
func (app *application) getArticleByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	article, err := app.store.Articles.GetByID(r.Context(), int(id))
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

	if err := app.jsonResponse(w, http.StatusOK, article); err != nil {
		app.internalServerError(w, r, err)
	}
}

// @Summary		Get articles by user id
// @Description	Get articles by user id
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id		path		int	true	"User ID"
// @Param			offset	query		int	true	"Offset"
// @Param			limit	query		int	true	"limit"
// @Success		200		{object}	[]store.Article
// @Failure		400		{object}	error
// @Failure		404		{object}	error
// @Failure		500		{object}	error
// @Security		ApiKeyAuth
// @Router			/articles/author/{id} [get]
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

// @Summary		Create article
// @Description	Create article
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			article	body		CreateArticlePayload	true	"Article"
// @Success		201		{object}	store.Article
// @Failure		400		{object}	error
// @Failure		500		{object}	error
// @Security		ApiKeyAuth
// @Router			/articles [post]
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

// @Summary		Update article
// @Description	Update article
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id		path		int						true	"Article ID"
// @Param			article	body		UpdateArticlePayload	true	"Article"
// @Success		200		{object}	int
// @Failure		400		{object}	error
// @Failure		500		{object}	error
// @Security		ApiKeyAuth
// @Router			/articles/{id} [patch]
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

// @Summary		Delete article
// @Description	Delete article
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id	path	int	true	"Article ID"
// @Success		204
// @Failure		400	{object}	error
// @Failure		404	{object}	error
// @Failure		500	{object}	error
// @Security		ApiKeyAuth
// @Router			/articles/{id} [delete]
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

// @Summary		Get comments of article by id
// @Description	Get comments of article by id
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id		path		int	true	"Article ID"
// @Param			offset	query		int	true	"Offset"
// @Param			limit	query		int	true	"Limit"
// @Success		200		{object}	[]store.Comment
// @Failure		400		{object}	error
// @Failure		404		{object}	error
// @Failure		500		{object}	error
// @Security		ApiKeyAuth
// @Router			/articles/{id}/comments [get]
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

type CommentOnlyText struct {
	Text string `json:"text"`
}

// @Summary		Create comment
// @Description	Create comment
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id		path		int				true	"Article ID"
// @Param			comment	body		CommentOnlyText	true	"Comment"
// @Success		201		{object}	int
// @Failure		400		{object}	error
// @Failure		500		{object}	error
// @Security		ApiKeyAuth
// @Router			/articles/{id}/comments [post]
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

// @Summary		set like on article
// @Description	set like on article
// @Tags			articles
// @Accept			json
// @Produce		json
// @Param			id	path	int	true	"Article ID"
// @Success		201
// @Failure		400	{object}	error
// @Failure		500	{object}	error
// @Security		ApiKeyAuth
// @Router			/articles/{id}/like [post]
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
