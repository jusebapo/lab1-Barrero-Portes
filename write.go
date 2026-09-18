package main

import (
	"context"
	"net/http"
	"time"
)

func init() {
	routes["POST /notes"] = createNote
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
