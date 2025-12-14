package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rpstvs/social/internal/store"
	HttpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config config
	store  store.Storage
	logger *zap.SugaredLogger
}

type config struct {
	addr   string
	db     dbConfig
	env    string
	apiURL string
	mail   mailConfig
}

type mailConfig struct {
	exp time.Duration
}

type dbConfig struct {
	addrDB      string
	maxOpenConn int
	maxIdleConn int
	maxIdleTime string
}

func NewConfig(addr, addrDB, maxIdleTime string, maxOpenConn, maxIdleConn int, mailExp time.Duration) config {
	return config{
		addr: addr,
		db:   NewDBConfig(addrDB, maxIdleTime, maxOpenConn, maxIdleConn),
		mail: NewMailConfig(mailExp),
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

func NewApplication(config config, storage store.Storage, logger *zap.SugaredLogger) *application {
	return &application{
		config: config,
		store:  storage,
		logger: logger,
	}
}

func NewMailConfig(mailExp time.Duration) mailConfig {
	return mailConfig{
		exp: mailExp,
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

		docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger", HttpSwagger.Handler(HttpSwagger.URL(docsUrl)))

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
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
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
