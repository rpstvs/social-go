package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/rpstvs/social/internal/db"
	"github.com/rpstvs/social/internal/env"
	"github.com/rpstvs/social/internal/store"
)

const DEFAULT_PORT = ":8080"
const DEFAULT_DB_ADDR = "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"
const DEFAULT_DB_MAXIDLETIME = "15m"
const DEFAULT_DB_MAX_OPENCONNS = 30
const DEFAULT_DB_MAX_IDLE_CONN = 30

func main() {

	godotenv.Load(".env")

	config := NewConfig(
		env.GetString("ADDR", DEFAULT_PORT),
		env.GetString("DB_ADDR", DEFAULT_DB_ADDR),
		env.GetString("DB_MAX_IDLE_TIME", DEFAULT_DB_MAXIDLETIME),
		env.GetInt("DB_MAX_OPEN_CONNS", DEFAULT_DB_MAX_OPENCONNS),
		env.GetInt("DB_MAX_IDLE_CONNS", DEFAULT_DB_MAX_IDLE_CONN))

	db, err := db.New(config.db.addrDB, config.db.maxOpenConn, config.db.maxIdleConn, config.db.maxIdleTime)

	if err != nil {
		log.Panic(err)
	}

	defer db.Close()

	log.Println("Connection to DB established.")

	store := store.NewStorage(db)

	app := NewApplication(config, store)

	mux := app.mount()

	log.Fatal(app.run(mux))
}
