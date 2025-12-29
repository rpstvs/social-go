package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rpstvs/social/internal/auth"
	"github.com/rpstvs/social/internal/ratelimiter"
	"github.com/rpstvs/social/internal/store"
	"github.com/rpstvs/social/internal/store/cache"
	HttpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	cacheStorage  cache.Storage
	logger        *zap.SugaredLogger
	authenticator auth.Authenticator
	rateLimiter   ratelimiter.Limiter
}

type config struct {
	addr        string
	db          dbConfig
	env         string
	apiURL      string
	mail        mailConfig
	authConfig  AuthConfig
	redisCfg    RedisConfig
	rateLimiter ratelimiter.Config
}

type AuthConfig struct {
	basic BasicConfig
	token TokenConfig
}

type TokenConfig struct {
	secret string
	exp    time.Duration
}

type BasicConfig struct {
	username string
	password string
}

type RedisConfig struct {
	addr     string
	password string
	database int
	enabled  bool
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

func NewConfig(addr, addrRedis, addrDB, maxIdleTime, username, password, pwRedis, secret string, maxOpenConn, maxIdleConn, dbRedis int, mailExp, expToken time.Duration, redisEnabled bool) config {
	return config{
		addr:       addr,
		db:         NewDBConfig(addrDB, maxIdleTime, maxOpenConn, maxIdleConn),
		mail:       NewMailConfig(mailExp),
		authConfig: NewAuthConfig(username, password, secret, expToken),
		redisCfg: RedisConfig{
			addr:     addrRedis,
			password: pwRedis,
			database: dbRedis,
			enabled:  redisEnabled,
		},
	}
}

func NewAuthConfig(username, password, secret string, exp time.Duration) AuthConfig {
	return AuthConfig{
		basic: BasicConfig{
			username: username,
			password: password,
		},
		token: TokenConfig{
			secret: secret,
			exp:    exp,
		},
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

func NewApplication(config config, storage store.Storage, cacheStorage cache.Storage, logger *zap.SugaredLogger) *application {
	return &application{
		config:        config,
		store:         storage,
		cacheStorage:  cacheStorage,
		logger:        logger,
		authenticator: auth.NewJwtAuthenticator(config.authConfig.token.secret, "gopherSocial", "gopherSocial"),
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
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Use(app.RateLimiterMiddleware)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware()).Get("/health", app.HealthCheckHandler)
		r.With(app.BasicAuthMiddleware()).Get("/debug/vars", expvar.Handler().ServeHTTP)

		docsUrl := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger", HttpSwagger.Handler(HttpSwagger.URL(docsUrl)))

		r.Route("/posts", func(r chi.Router) {
			r.Use(app.AuthTokenMiddleware())
			r.Post("/", app.CreatePostHandler)

			r.Route("/{postsID}", func(r chi.Router) {

				r.Use(app.postsContextMiddleware)

				r.Get("/", app.GetPostHandler)
				r.Patch("/", app.checkPostOwnership("moderator", app.UpdatePostHandler))
				r.Delete("/", app.checkPostOwnership("admin", app.DeletePostHandler))

			})

		})

		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{activate}", app.activateUserHandler)
			r.Route("/{userid}", func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware())
				r.Get("/", app.getUserHandler)
				r.Put("/follow", app.followUserHandler)
				r.Put("/unfollow", app.unfollowUserHandler)
			})
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware())
				r.Get("/feed", app.getUserFeedHandler)
			})
		})

		r.Route("/authentication", func(r chi.Router) {
			r.Post("/user", app.registerUserHandler)
			r.Post("/token", app.createTokenHandler)
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

	shutdown := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		defer cancel()

		app.logger.Infow("signal caught", "signal", s.String())

		shutdown <- srv.Shutdown(ctx)
	}()

	err := srv.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdown
	if err != nil {
		return err
	}

	app.logger.Infow("server has stopped", "addr", app.config.addr)
	return nil
}
