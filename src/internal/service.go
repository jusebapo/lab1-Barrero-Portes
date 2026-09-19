package internal

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-playground/validator/v10"
)

var NoteNotFoundErr = errors.New("Note not found")

type Service struct {
	db *sql.DB
	v  *validator.Validate
}

func NewService(db *sql.DB, v *validator.Validate) *Service {
	return &Service{
		db: db,
		v:  v,
	}
}

func (s *Service) createNote(ctx context.Context, i NoteInput) (note, error) {
	var n note
	if err := s.v.Struct(i); err != nil {
		return n, err
	}
	query := "insert into notes(title,content,author) values ($1,$2,$3) returning id,title,content,author,created_at"
	result := s.db.QueryRow(query, i.Title, i.Content, i.Author)
	if err := result.Scan(&n.ID, &n.Title, &n.Content, &n.Author, &n.CreatedAt); err != nil {
		return n, err
	}
	return n, nil
}

func (s *Service) getNotes() ([]note, error) {
	var notes []note
	query := "select id,title,content,author,created_at from notes"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Author, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return notes, nil
}

func (s *Service) getNoteById(id int) (note, error) {
	var n note
	query := "select id,title,content,author,created_at from notes where id = $1"
	row := s.db.QueryRow(query, id)
	err := row.Scan(&n.ID, &n.Title, &n.Content, &n.Author, &n.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return n, NoteNotFoundErr
		}
		return n, err
	}
	return n, nil
}

func (s *Service) deleteNote() {}
