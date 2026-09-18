package internal

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"
)

func init() {
	routes["POST /notes"] = createNote
	routes["PUT /notes/{id}"] = updateNote
	routes["DELETE /notes/{id}"] = deleteNote
}

func createNote(app *application, w http.ResponseWriter, r *http.Request) {
	input, ok := readInput(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	n, err := scanNote(app.db.QueryRowContext(ctx, `
		INSERT INTO notes (title, content, author)
		VALUES ($1, $2, $3)
		RETURNING id, title, content, author, created_at`,
		input.Title, input.Content, input.Author))
	if err != nil {
		internalError(w)
		return
	}
	setLocation(w, n.ID)
	writeJSON(w, http.StatusCreated, n)
}

func updateNote(app *application, w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	input, ok := readInput(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	n, err := scanNote(app.db.QueryRowContext(ctx, `
		UPDATE notes SET title = $1, content = $2, author = $3
		WHERE id = $4
		RETURNING id, title, content, author, created_at`,
		input.Title, input.Content, input.Author, id))
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

func deleteNote(app *application, w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var deletedID int64
	err := app.db.QueryRowContext(ctx,
		"DELETE FROM notes WHERE id = $1 RETURNING id", id).Scan(&deletedID)
	if errors.Is(err, sql.ErrNoRows) {
		fail(w, http.StatusNotFound, "Nota no encontrada")
		return
	}
	if err != nil {
		internalError(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Nota eliminada",
		"id":      deletedID,
	})
}
