package main

import (
	"expvar"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/rpstvs/social/internal/db"
	"github.com/rpstvs/social/internal/env"
	"github.com/rpstvs/social/internal/store"
	"github.com/rpstvs/social/internal/store/cache"
	"go.uber.org/zap"
)

const DEFAULT_PORT = ":8080"
const DEFAULT_DB_ADDR = "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"
const DEFAULT_DB_MAXIDLETIME = "15m"
const DEFAULT_DB_MAX_OPENCONNS = 30
const DEFAULT_DB_MAX_IDLE_CONN = 30
const DEFAULT_EXP_MAIL_INVITATION = 3 * time.Hour
const DEFAULT_USERNAME = "rui"
const DEFAULT_PASSWORD = "oliveira"
const DEFAULT_REDIS_ADDR = "localhost"
const DEFAULT_REDIS_PW = "admin"
const DEFAULT_EXP_TOKEN = 3 * time.Hour

func main() {

	godotenv.Load(".env")

	config := NewConfig(
		env.GetString("ADDR", DEFAULT_PORT),
		env.GetString("REDIS_ADDR", DEFAULT_REDIS_ADDR),
		env.GetString("DB_ADDR", DEFAULT_DB_ADDR),
		env.GetString("DB_MAX_IDLE_TIME", DEFAULT_DB_MAXIDLETIME),
		DEFAULT_USERNAME,
		DEFAULT_PASSWORD,
		DEFAULT_REDIS_PW,
		"cenas",
		env.GetInt("DB_MAX_OPEN_CONNS", DEFAULT_DB_MAX_OPENCONNS),
		env.GetInt("DB_MAX_IDLE_CONNS", DEFAULT_DB_MAX_IDLE_CONN),
		0,
		DEFAULT_EXP_TOKEN,
		DEFAULT_EXP_MAIL_INVITATION,
		true)

	db, err := db.New(config.db.addrDB, config.db.maxOpenConn, config.db.maxIdleConn, config.db.maxIdleTime)

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Info("Connection to DB established.")

	var rdb *redis.Client

	if config.redisCfg.enabled {
		rdb = cache.NewRedisClient(config.redisCfg.addr, config.redisCfg.password, config.redisCfg.database)
		logger.Info("redis connection established")
	}

	store := store.NewStorage(db)
	cacheStore := cache.NewRedisStorage(rdb)

	app := NewApplication(config, store, cacheStore, logger)

	expvar.NewString("version").Set("0.0.0.1")
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
