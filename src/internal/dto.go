package internal

type NoteInput struct {
	Title   string `json:"title" validate:"required,min=2"`
	Content string `json:"content" validate:"required,min=2"`
	Author  string `json:"author" validate:"required,min=2"`
}
