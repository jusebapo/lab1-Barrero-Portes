package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"team-notes-api/internal"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/lib/pq"
)

func main() {
	port := requiredEnv("APP_PORT")
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		log.Fatal("APP_PORT must be a valid TCP port")
	}
	dsn := &url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			requiredEnv("DB_USER"), requiredEnv("DB_PASSWORD"),
		),
		Host: net.JoinHostPort(requiredEnv("DB_HOST"), requiredEnv("DB_PORT")),
		Path: "/" + requiredEnv("DB_NAME"),
	}
	query := dsn.Query()
	query.Set("sslmode", "disable")
	query.Set("connect_timeout", "5")
	dsn.RawQuery = query.Encode()

	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		log.Fatal("database configuration failed")
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = db.PingContext(ctx)
	cancel()
	if err != nil {
		log.Fatal("database connection failed; check database service and environment")
	}
	mux := http.NewServeMux()

	v := validator.New()
	service := internal.NewService(db, v)
	handler := internal.NewHandler(*service)

	mux.HandleFunc("GET		/api/v1/notes", handler.HandleGetNotes)
	mux.HandleFunc("GET 	/api/v1/notes/{id}", handler.HandleGetNoteById)
	mux.HandleFunc("POST	/api/v1/notes", handler.HandleCreateNote)
	mux.HandleFunc("PUT		/api/v1/notes", handler.HandleUpdateNote)
	mux.HandleFunc("DELETE 	/api/v1/notes/{id}", handler.HandleDeleteNote)
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("API listening on port %s", port)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("HTTP server stopped unexpectedly")
	}
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing environment variable: %s", key)
	}
	return value
}
