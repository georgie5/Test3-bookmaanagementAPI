package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ReadingListBookModel struct {
	DB *sql.DB
}

// AddBook adds a book to a reading list
func (m *ReadingListBookModel) AddBook(listID int64, bookID int64) error {
	query := `
		INSERT INTO reading_lists_books (reading_list_id, book_id)
		SELECT $1, $2
		WHERE NOT EXISTS (
            SELECT 1 FROM book_authors WHERE book_id = $1 AND author_id = $2
        )
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, listID, bookID)
	if err != nil {
		return fmt.Errorf("unable to add book to reading list: %w", err)
	}
	return nil
}

// RemoveBook removes a book from a reading list
func (m *ReadingListBookModel) RemoveBook(listID int64, bookID int64) error {
	query := `
		DELETE FROM reading_lists_books
		WHERE reading_list_id = $1 AND book_id = $2
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, listID, bookID)
	return err
}
