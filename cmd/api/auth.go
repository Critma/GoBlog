package main

import (
	"net/http"
	"time"

	"github.com/critma/goblog/internal/store"
	"github.com/golang-jwt/jwt/v5"
)

type ToRegisterPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=7,max=72"`
}

// @Summary		Register user
// @Description	Register user
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			user	body		ToRegisterPayload	true	"User"
// @Success		204		{object}	nil
// @Failure		400		{object}	error
// @Failure		500		{object}	error
// @Router			/auth/reg [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload ToRegisterPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	ctx := r.Context()

	if err := app.store.Users.Create(ctx, user); err != nil {
		app.internalServerError(w, r, err)
	}

	app.jsonResponse(w, http.StatusNoContent, nil)
}

type ToLoginPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=7,max=72"`
}

// @Summary		Login user
// @Description	Login user
// @Tags			auth
// @Accept			json
// @Produce		json
// @Param			user	body		ToLoginPayload	true	"User"
// @Success		202		{object}	string
// @Failure		400		{object}	error
// @Failure		404		{object}	error
// @Failure		500		{object}	error
// @Router			/auth/log [post]
func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload ToLoginPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}

	if err := user.Password.CompareWithHash(payload.Password); err != nil {
		app.unauthorizedErrorResponse(w, r, err)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.issuer,
		"aud": app.config.auth.issuer,
	}

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusAccepted, token); err != nil {
		app.internalServerError(w, r, err)
	}
}
