package main

import (
	"net/http"
	"strconv"

	"github.com/critma/goblog/internal/store"
	"github.com/go-chi/chi/v5"
)

type userKey string

const userCtx userKey = "user"

// @description	Get user by ID
// @summary		Get user by ID
// @Tags			users
// @Accept			json
// @Produce		json
// @Param			id	path		int	true	"User ID"
// @Success		200	{object}	store.User
// @Failure		400	{object}	error
// @Failure		404	{object}	error
// @Router			/users/{id} [get]
func (app *application) getUserByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// user, err := app.getUser(r.Context(), userID)
	user, err := app.store.Users.GetByID(r.Context(), int(userID))
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

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
