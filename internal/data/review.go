// internal/data/review.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/georgie5/Test3-bookclubapi/internal/validator"
)

// Review represents a review for a book.
type Review struct {
	ID         int64     `json:"id"`
	BookID     int64     `json:"book_id"`
	UserID     int64     `json:"user_id"`
	Rating     int       `json:"rating"`
	Review     string    `json:"review"`
	ReviewDate time.Time `json:"review_date"`
	Version    int32     `json:"version"`
}

type ReviewModel struct {
	DB *sql.DB
}

// validate validates the review data.
func ValidateReview(v *validator.Validator, review *Review) {
	v.Check(review.Rating > 0, "rating", "must be a positive integer")
	v.Check(review.Rating <= 5, "rating", "must not be greater than 5")
	v.Check(len(review.Review) > 0, "review", "must not be empty")
}

// Insert inserts a new review for a book.
func (m *ReviewModel) Insert(review *Review) error {
	query := `
		INSERT INTO reviews (book_id, user_id, rating, review, review_date)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, review_date, version
	`
	args := []interface{}{
		review.BookID, review.UserID, review.Rating, review.Review,
	}

	return m.DB.QueryRow(query, args...).Scan(
		&review.ID,
		&review.ReviewDate,
		&review.Version)
}

// Get fetches a review by ID.
func (m *ReviewModel) Get(id int64) (*Review, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, book_id, user_id, rating, review, review_date, version
		FROM reviews
		WHERE id = $1
	`
	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&review.ID,
		&review.BookID,
		&review.UserID,
		&review.Rating,
		&review.Review,
		&review.ReviewDate,
		&review.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &review, nil

}

// update review
func (m *ReviewModel) Update(review *Review) error {
	query := `
		UPDATE reviews
		SET rating = $1, review = $2, version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING  version
	`
	args := []interface{}{
		review.Rating, review.Review, review.ID, review.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&review.Version)

}

// delete review
func (m *ReviewModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM reviews
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// get all reviews for specific book
func (m *ReviewModel) GetAll(bookID int64, rating int, review string, filters Filters) ([]*Review, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, book_id, user_id, rating, review, review_date, version
        FROM reviews
        WHERE book_id = $1
        AND (rating = $2 OR $2 = 0)
        AND (review ILIKE '%%' || $3 || '%%' OR $3 = '')  -- filtering based on review content
        ORDER BY %s %s, id ASC
        LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, bookID, rating, review, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var reviews []*Review
	totalRecords := 0

	for rows.Next() {
		var review Review
		err := rows.Scan(
			&totalRecords,
			&review.ID,
			&review.BookID,
			&review.UserID,
			&review.Rating,
			&review.Review,
			&review.ReviewDate,
			&review.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return reviews, metadata, nil
}
