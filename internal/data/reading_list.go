package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/georgie5/Test3-bookclubapi/internal/validator"
)

type ReadingList struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedBy   int64  `json:"created_by"`
	Status      string `json:"status"`
	Version     int32  `json:"version"`
}

type ReadingListModel struct {
	DB *sql.DB
}

// validateReadingList validates the reading list
func ValidateReadingList(v *validator.Validator, r *ReadingList) {
	v.Check(r.Name != "", "name", "must be provided")
	v.Check(r.Description != "", "description", "must be provided")
	v.Check(r.CreatedBy > 0, "created_by", "must be valid")
	v.Check(r.Status == "currently reading" || r.Status == "completed", "status", "must be either 'draft' or 'published'")
}

func (m *ReadingListModel) Insert(r *ReadingList) error {
	query := `
		INSERT INTO reading_lists (name, description, created_by, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, version
	`
	args := []any{r.Name, r.Description, r.CreatedBy, r.Status}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&r.ID, &r.Version)
}

func (m *ReadingListModel) Get(id int64) (*ReadingList, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, description, created_by, status, version
		FROM reading_lists
		WHERE id = $1
	`

	var r ReadingList
	err := m.DB.QueryRow(query, id).Scan(&r.ID, &r.Name, &r.Description, &r.CreatedBy, &r.Status, &r.Version)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &r, nil
}

func (m *ReadingListModel) Update(r *ReadingList) error {
	query := `
		UPDATE reading_lists
		SET name = $1, description = $2, status = $3, created_by = $4,version = version + 1
		WHERE id = $5
		RETURNING version
	`
	args := []any{r.Name, r.Description, r.Status, r.CreatedBy, r.ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&r.Version)
}

func (m *ReadingListModel) Delete(id int64) error {
	query := `DELETE FROM reading_lists WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, id)
	return err
}
func (m *ReadingListModel) GetAll(name, description, status string, filters Filters) ([]*ReadingList, Metadata, error) {

	// Safely format the query string, using the dynamic sort column and direction
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, name, description, created_by, status, version
		FROM reading_lists
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple', description) @@ plainto_tsquery('simple', $2) OR $2 = '')
		AND (status = $3 OR $3 = '')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Pass the result of the function calls (filters.limit() and filters.offset())
	rows, err := m.DB.QueryContext(ctx, query, name, description, status, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	var readingLists []*ReadingList

	// Loop through the rows and scan the results into the readingLists slice
	for rows.Next() {
		var r ReadingList
		err := rows.Scan(
			&totalRecords,
			&r.ID,
			&r.Name,
			&r.Description,
			&r.CreatedBy,
			&r.Status,
			&r.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		readingLists = append(readingLists, &r)
	}

	// Check for any errors during iteration over rows
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate the metadata for pagination (e.g., current page, total pages, etc.)
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return readingLists, metadata, nil
}
