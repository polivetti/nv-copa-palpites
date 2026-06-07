package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"nv-copa/internal/app"
	"nv-copa/internal/db"
)

func main() {
	addr := env("ADDR", ":8080")
	dbPath := env("DATABASE_PATH", "data/copa.db")

	store, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	if err := store.Seed(); err != nil {
		log.Fatalf("seed database: %v", err)
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.New(store),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("nv-copa listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
