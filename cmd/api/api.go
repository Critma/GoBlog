package main

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/critma/goblog/docs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/reg", app.registerUserHandler)
			r.Post("/log", app.loginUserHandler)
		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", app.getUserByIDHandler)
			})
		})

		r.Route("/articles", func(r chi.Router) {
			r.Get("/", app.getLatestArticlesHandler)
			r.Group(func(r chi.Router) { // with middleware
				r.Use(app.AuthTokenMiddleware)
				r.Post("/", app.createArticleHandler)
				r.Route("/{id}", func(r chi.Router) {
					r.Use(app.articleContextMiddleware)
					r.Get("/", app.getArticleByID)

					r.Route("/comments", func(r chi.Router) {
						r.Get("/", app.getArticleCommentsHandler)
						r.Post("/", app.createArticleCommentHandler)
					})
					r.Post("/like", app.createLikeOnArticle)

					r.Group(func(r chi.Router) {
						r.Use(app.CheckArticleOwnershipMiddleware)
						r.Delete("/", app.deleteArticleHandler)
						r.Patch("/", app.updateArticleHandler)
					})
				})
				r.Get("/author/{id}", app.getArticlesByUserID)
			})
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Minute,
	}

	app.logger.Infow("server start on ", "addr", app.config.addr)
	err := srv.ListenAndServe()
	if err != nil {
		return err
	}

	app.logger.Infow("server stop", "addr", app.config.addr)

	return nil
}
