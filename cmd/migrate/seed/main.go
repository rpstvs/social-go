package main

import (
	"log"

	"github.com/rpstvs/social/internal/db"
	"github.com/rpstvs/social/internal/env"
	"github.com/rpstvs/social/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "Fallback")
	conn, err := db.New(addr, 30, 30, "15m")
	if err != nil {
		log.Fatal(err)
	}
	store := store.NewStorage(conn)
	db.Seed(store, conn)
}
