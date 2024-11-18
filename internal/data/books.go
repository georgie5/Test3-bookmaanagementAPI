package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/georgie5/Test3-bookclubapi/internal/validator"
)

// bookModel wraps the database connection pool
type BookModel struct {
	DB *sql.DB
}

// Book represents a book in the catalog
type Book struct {
	ID              int64     `json:"id"`
	Title           string    `json:"title"`
	ISBN            string    `json:"isbn"`
	PublicationDate time.Time `json:"publication_date"`
	Genre           string    `json:"genre"`
	Description     string    `json:"description"`
	AverageRating   float64   `json:"average_rating"`
	Version         int32     `json:"version"` // incremented on each update
}

func ValidateBook(v *validator.Validator, b *Book) {
	v.Check(b.Title != "", "title", "must be provided")
	v.Check(b.ISBN != "", "isbn", "must be provided")
	v.Check(!b.PublicationDate.IsZero(), "publication_date", "must be provided")
	v.Check(b.Genre != "", "genre", "must be provided")
	v.Check(b.Description != "", "description", "must be provided")

}

// Insert inserts a new book into the database
func (m BookModel) Insert(book *Book) error {
	query := `
		INSERT INTO books (title, isbn, publication_date, genre, description, average_rating) 
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, version
		`
	args := []any{book.Title, book.ISBN, book.PublicationDate, book.Genre, book.Description, book.AverageRating}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Insert and retrieve the new book ID and version
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&book.ID, &book.Version)
}

// Get fetches a book by ID
func (m *BookModel) Get(bookID int64) (*Book, []Author, error) {
	if bookID < 1 {
		return nil, nil, ErrRecordNotFound
	}

	query := `
		SELECT b.*,a.name
		FROM books b
		JOIN book_authors ba ON b.id = ba.book_id
		JOIN authors a ON a.id = ba.author_id
		WHERE b.id = $1
		`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var book Book
	authorMap := make(map[string]Author) // To ensure unique authors

	rows, err := m.DB.QueryContext(ctx, query, bookID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	// Loop through each row, appending authors to a map (avoiding duplicates)
	for rows.Next() {
		var author Author
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.ISBN,
			&book.PublicationDate,
			&book.Genre,
			&book.Description,
			&book.AverageRating,
			&book.Version,
			&author.Name,
		)
		if err != nil {
			return nil, nil, err
		}

		// Ensure that we only add unique authors to the map
		if _, exists := authorMap[author.Name]; !exists {
			authorMap[author.Name] = author
		}
	}

	// Convert map to slice
	authors := make([]Author, 0, len(authorMap))
	for _, author := range authorMap {
		authors = append(authors, author)
	}

	return &book, authors, nil
}

// Update updates a book in the database
func (m BookModel) Update(book *Book) error {
	query := `
		UPDATE books
		SET title = $1, isbn = $2, publication_date = $3, genre = $4, description = $5, average_rating = $6, version = version + 1
		WHERE id = $7 
		RETURNING version
		`
	args := []any{book.Title, book.ISBN, book.PublicationDate, book.Genre, book.Description, book.AverageRating, book.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&book.Version)
}

// Delete deletes a book from the database
func (m BookModel) Delete(id int64) error {
	//check if the id is valid
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM books
		WHERE id = $1
		`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// were any rows deleted?
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

// list all the books with pagination
func (m *BookModel) GetAll(title, genre string, filters Filters) ([]*Book, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, title, isbn, publication_date, genre, description, average_rating, version
		FROM books
		WHERE (title ILIKE '%%' || $1 || '%%' OR $1 = '')
		AND (genre ILIKE '%%' || $2 || '%%' OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title, genre, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var books []*Book
	totalRecords := 0

	for rows.Next() {
		var book Book
		err := rows.Scan(&totalRecords, &book.ID, &book.Title, &book.ISBN, &book.PublicationDate, &book.Genre, &book.Description, &book.AverageRating, &book.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return books, metadata, nil
}

func (m *BookModel) Search(title, author, genre string, filters Filters) ([]*Book, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), b.id, b.title, b.isbn, b.publication_date, b.genre, b.description, b.average_rating, b.version
		FROM books b
		JOIN book_authors ba ON b.id = ba.book_id
		JOIN authors a ON a.id = ba.author_id
		WHERE (b.title ILIKE '%%' || $1 || '%%' OR $1 = '')
		AND (a.name ILIKE '%%' || $2 || '%%' OR $2 = '')
		AND (b.genre ILIKE '%%' || $3 || '%%' OR $3 = '')
		ORDER BY %s %s, b.id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, title, author, genre, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var books []*Book
	totalRecords := 0

	for rows.Next() {
		var book Book
		err := rows.Scan(&totalRecords, &book.ID, &book.Title, &book.ISBN, &book.PublicationDate, &book.Genre, &book.Description, &book.AverageRating, &book.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return books, metadata, nil
}

// update the book average rating
func (m *BookModel) UpdateAverageRating(id int64) error {
	query := `
		UPDATE books
		SET average_rating = (
			SELECT COALESCE(AVG(rating), 0)
			FROM reviews
			WHERE book_id = $1
		)
		WHERE id = $1
		`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, id)
	return err
}
