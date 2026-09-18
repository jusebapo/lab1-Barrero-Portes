package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"
)

func init() {
	routes["GET /health"] = health
	routes["GET /notes"] = listNotes
	routes["GET /notes/{id}"] = getNote
}

func health(app *application, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := app.db.PingContext(ctx); err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":      "ok",
		"environment": app.env,
	})
}

func listNotes(app *application, w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx,
		"SELECT id, title, content, author, created_at FROM notes ORDER BY id")
	if err != nil {
		internalError(w)
		return
	}
	defer rows.Close()
	notes := make([]note, 0)
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			internalError(w)
			return
		}
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func getNote(app *application, w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	n, err := scanNote(app.db.QueryRowContext(ctx,
		"SELECT id, title, content, author, created_at FROM notes WHERE id = $1", id))
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, http.StatusNotFound, "Nota no encontrada")
		return
	}
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, n)
}
