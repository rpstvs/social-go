package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rpstvs/social/internal/store"
)

type application struct {
	config config
	store  store.Storage
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	addrDB      string
	maxOpenConn int
	maxIdleConn int
	maxIdleTime string
}

func NewConfig(addr, addrDB, maxIdleTime string, maxOpenConn, maxIdleConn int) config {
	return config{
		addr: addr,
		db:   NewDBConfig(addrDB, maxIdleTime, maxOpenConn, maxIdleConn),
	}
}

func NewDBConfig(addrDB, maxIdleTime string, maxOpenCon, maxIdleConn int) dbConfig {
	return dbConfig{
		addrDB:      addrDB,
		maxOpenConn: maxOpenCon,
		maxIdleConn: maxIdleConn,
		maxIdleTime: maxIdleTime,
	}
}

func NewApplication(config config, storage store.Storage) *application {
	return &application{
		config: config,
		store:  storage,
	}
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.HealthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Post("/", app.CreatePostHandler)

			r.Route("/{postsID}", func(r chi.Router) {
				r.Use(app.postsContextMiddleware)

				r.Get("/", app.GetPostHandler)
				r.Delete("/", app.DeletePostHandler)
				r.Patch("/", app.UpdatePostHandler)
			})

		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/{userid}", func(r chi.Router) {
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
		})
	})

	return r
}

func (app *application) run(mux http.Handler) error {

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	return srv.ListenAndServe()
}
