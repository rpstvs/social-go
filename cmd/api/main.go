package main

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/rpstvs/social/internal/db"
	"github.com/rpstvs/social/internal/env"
	"github.com/rpstvs/social/internal/store"
	"go.uber.org/zap"
)

const DEFAULT_PORT = ":8080"
const DEFAULT_DB_ADDR = "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"
const DEFAULT_DB_MAXIDLETIME = "15m"
const DEFAULT_DB_MAX_OPENCONNS = 30
const DEFAULT_DB_MAX_IDLE_CONN = 30
const DEFAULT_EXP_MAIL_INVITATION = 3 * time.Hour

func main() {

	godotenv.Load(".env")

	config := NewConfig(
		env.GetString("ADDR", DEFAULT_PORT),
		env.GetString("DB_ADDR", DEFAULT_DB_ADDR),
		env.GetString("DB_MAX_IDLE_TIME", DEFAULT_DB_MAXIDLETIME),
		env.GetInt("DB_MAX_OPEN_CONNS", DEFAULT_DB_MAX_OPENCONNS),
		env.GetInt("DB_MAX_IDLE_CONNS", DEFAULT_DB_MAX_IDLE_CONN),
		DEFAULT_EXP_MAIL_INVITATION)

	db, err := db.New(config.db.addrDB, config.db.maxOpenConn, config.db.maxIdleConn, config.db.maxIdleTime)

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()

	logger.Info("Connection to DB established.")

	store := store.NewStorage(db)

	app := NewApplication(config, store, logger)

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
