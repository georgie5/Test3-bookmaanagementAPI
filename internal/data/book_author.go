package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// BookAuthorModel handles the many-to-many relationship between books and authors.
type BookAuthorModel struct {
	DB *sql.DB
}

// Insert links a book with an author in the book_authors table.
func (m *BookAuthorModel) Insert(bookID int64, authorID int64) error {
	query := `
		INSERT INTO book_authors (book_id, author_id)
		SELECT $1, $2
		WHERE NOT EXISTS (
            SELECT 1 FROM book_authors WHERE book_id = $1 AND author_id = $2
        )
	`
	args := []any{bookID, authorID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("unable to insert book-author relationship: %w", err)
	}
	return nil

}

// Delete removes the relationship between a book and an author
func (m *BookAuthorModel) Delete(bookID int64) error {
	query := `
        DELETE FROM book_authors
        WHERE book_id = $1 
    `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, bookID)
	if err != nil {
		return fmt.Errorf("unable to delete book-author relationships: %w", err)
	}

	return nil
}
