package main

import (
	"context"
	"net/http"
	"time"
)

func init() {
	routes["GET /health"] = health
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
