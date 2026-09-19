package internal

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type Handler struct {
	s Service
}

func NewHandler(s Service) Handler {
	return Handler{s: s}
}

func (h *Handler) HandleCreateNote(w http.ResponseWriter, r *http.Request) {
	var i NoteInput
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		log.Println(err.Error())
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	created, err := h.s.createNote(r.Context(), i)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeResponse(w, created)
}

func (h *Handler) HandleGetNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.s.getNotes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeResponse(w, notes)
}

func (h *Handler) HandleGetNoteById(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id, must be a positive integer", http.StatusBadRequest)
		return
	}
	note, err := h.s.getNoteById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeResponse(w, note)
}

func (h *Handler) HandleDeleteNote() {}

func writeResponse(w http.ResponseWriter, message any) {
	bytes, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}
