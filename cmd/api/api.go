package main

import (
	"net/http"
	"time"

	"github.com/bishop254/bursary/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Create a struct for our server configurations
type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
}

type config struct {
	addr string
	db   dbConfig
	env  string
	mail mailConfig
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type mailConfig struct {
	tokenExp time.Duration
}

func (a *application) mount() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Use(middleware.Timeout(time.Minute))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", a.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", a.createPostHandler)
			r.Post("/comment", a.createPostCommentHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Use(a.postContextMiddleware)
				r.Get("/", a.getOnePostHandler)
				r.Delete("/", a.deletePostHandler)
				r.Patch("/", a.updatePostHandler)
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/activate/{token}", func(r chi.Router) {
				r.Get("/", a.activateUserHandler)
			})

			r.Route("/{userId}", func(r chi.Router) {
				r.Use(a.userContextMiddleware)
				r.Get("/", a.getOneUserHandler)
				r.Put("/follow", a.followUserHandler)
				r.Put("/unfollow", a.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Get("/feed", a.getUserFeedHandler)
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/user", a.createUserHandler)
			r.Post("/activate", a.createUserHandler)
		})
	})

	return r
}

func (a *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         a.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	a.logger.Infow("Server started on port ", "addr", a.config.addr)

	return srv.ListenAndServe()
}
