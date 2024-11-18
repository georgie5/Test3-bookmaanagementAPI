package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/georgie5/Test3-bookclubapi/internal/validator"
)

// Author represents an author of a book.
type Author struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type AuthorModel struct {
	DB *sql.DB
}

// validate validates the author fields.
func ValidateAuthor(v *validator.Validator, a *Author) {
	v.Check(a.Name != "", "name", "author must be provided")
}

// Insert inserts a new author into the database.
func (m *AuthorModel) Insert(author *Author) error {
	query := `
		INSERT INTO authors (name)
		VALUES ($1)
		RETURNING id
	`
	args := []any{author.Name}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&author.ID)
}

// Get returns an author by name.
func (m *AuthorModel) Get(name string) (*Author, error) {
	query := `
		SELECT id, name
		FROM authors
		WHERE name = $1
	`
	var author Author

	err := m.DB.QueryRow(query, name).Scan(
		&author.ID,
		&author.Name,
	)

	// If the author doesn't exist, return ErrRecordNotFound
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &author, nil
}
