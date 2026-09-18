package internal

import (
	"context"
	"database/sql"
)

type service struct {
	db *sql.DB
}

func (s *service) createNote(ctx context.Context, i noteInput) (*note, error) {
	query := "insert into notes(title,content,author) values ($1,$2,$3) returning *"
	result := s.db.QueryRow(query, i.Title, i.Content, i.Author)
	var n *note
	if err := result.Scan(&n.ID, &n.Title, &n.Content, &n.Author, &n.CreatedAt); err != nil {
		return nil, err
	}

}

func (s *service) getNotes() {}

func (s *service) getNoteById() {}

func (s *service) deleteNote() {}
