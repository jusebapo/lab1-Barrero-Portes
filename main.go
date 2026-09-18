package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	_ "github.com/lib/pq"
)

type application struct {
	db  *sql.DB
	env string
}

type note struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

type noteInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

type endpoint func(*application, http.ResponseWriter, *http.Request)

var routes = map[string]endpoint{}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("missing environment variable: %s", key)
	}
	return value
}

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

	app := &application{db: db, env: requiredEnv("APP_ENV")}
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           app,
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

func (app *application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/notes/") {
		id := strings.TrimPrefix(path, "/notes/")
		if id != "" && !strings.Contains(id, "/") {
			path = "/notes/{id}"
			r.SetPathValue("id", id)
		}
	}
	if handler, ok := routes[r.Method+" "+path]; ok {
		handler(app, w, r)
		return
	}
	allowed := map[string]string{
		"/health":     "GET",
		"/notes":      "GET, POST",
		"/notes/{id}": "GET, PUT, DELETE",
	}
	if methods, ok := allowed[path]; ok {
		w.Header().Set("Allow", methods)
		fail(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}
	fail(w, http.StatusNotFound, "Ruta no encontrada")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Print("response could not be written")
	}
}

func fail(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func internalError(w http.ResponseWriter) {
	log.Print("database operation failed")
	fail(w, http.StatusInternalServerError, "No fue posible completar la operación")
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		fail(w, http.StatusBadRequest, "El id debe ser un entero positivo")
		return 0, false
	}
	return id, true
}

func readInput(w http.ResponseWriter, r *http.Request) (noteInput, bool) {
	var input noteInput
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		fail(w, http.StatusBadRequest, "JSON inválido o campos no permitidos")
		return input, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		fail(w, http.StatusBadRequest, "Debe enviar un único objeto JSON")
		return input, false
	}
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	input.Author = strings.TrimSpace(input.Author)
	if input.Title == "" || input.Content == "" || input.Author == "" {
		fail(w, http.StatusBadRequest, "title, content y author son obligatorios")
		return input, false
	}
	if utf8.RuneCountInString(input.Title) > 200 ||
		utf8.RuneCountInString(input.Author) > 100 {
		fail(w, http.StatusBadRequest, "title admite hasta 200 caracteres y author hasta 100")
		return input, false
	}
	return input, true
}

type scanner interface {
	Scan(...any) error
}

func scanNote(row scanner) (note, error) {
	var n note
	err := row.Scan(&n.ID, &n.Title, &n.Content, &n.Author, &n.CreatedAt)
	return n, err
}

func setLocation(w http.ResponseWriter, id int64) {
	w.Header().Set("Location", fmt.Sprintf("/notes/%d", id))
}
