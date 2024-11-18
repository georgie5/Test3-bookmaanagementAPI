package main

import (
	"errors"
	"net/http"

	"github.com/georgie5/Test3-bookclubapi/internal/data"
	"github.com/georgie5/Test3-bookclubapi/internal/validator"
)

func (a *applicationDependencies) addReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Get the book ID from the URL
	bookID, err := a.readIDParam(r, "book_id")
	if err != nil || bookID < 1 {
		a.notFoundResponse(w, r)
		return
	}

	var incomingData struct {
		UserID int64  `json:"user_id"`
		Rating int    `json:"rating"`
		Review string `json:"review"`
	}

	// Decode the incoming JSON
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Create the review object
	review := &data.Review{
		BookID: bookID,
		UserID: incomingData.UserID,
		Rating: incomingData.Rating,
		Review: incomingData.Review,
	}

	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the review into the database
	err = a.reviewModel.Insert(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.bookModel.UpdateAverageRating(bookID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send the response with the created review
	data := envelope{
		"review": review,
	}

	err = a.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// Get the review ID from the URL
func (a *applicationDependencies) updateReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Get the review ID from the URL
	reviewID, err := a.readIDParam(r, "review_id")
	if err != nil || reviewID < 1 {
		a.notFoundResponse(w, r)
		return
	}

	var incomingData struct {
		Rating *int    `json:"rating"`
		Review *string `json:"review"`
	}

	// Decode the incoming JSON
	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Get the review from the database
	review, err := a.reviewModel.Get(reviewID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	//// We need to now check the fields to see which ones need updating
	if incomingData.Rating != nil {
		review.Rating = *incomingData.Rating
	}
	if incomingData.Review != nil {
		review.Review = *incomingData.Review
	}

	// Before we write the updates to the DB let's validate
	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the review in the database
	err = a.reviewModel.Update(review)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	err = a.bookModel.UpdateAverageRating(review.BookID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	//Send a JSON response with the updated product
	data := envelope{"review": review}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	// Get the review ID from the URL
	reviewID, err := a.readIDParam(r, "review_id")
	if err != nil || reviewID < 1 {
		a.notFoundResponse(w, r)
		return
	}

	// Get the review from the database
	review, err := a.reviewModel.Get(reviewID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Delete the review from the database
	err = a.reviewModel.Delete(reviewID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	err = a.bookModel.UpdateAverageRating(review.BookID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send a confirmation response
	data := envelope{
		"message": "review successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// Get all reviews for a book
func (a *applicationDependencies) getReviewsForBookHandler(w http.ResponseWriter, r *http.Request) {
	// Get the book ID from the URL
	bookID, err := a.readIDParam(r, "book_id")
	if err != nil || bookID < 1 {
		a.notFoundResponse(w, r)
		return
	}

	// set up query parameter struct
	var queryParametersData struct {
		Rating int
		Review string
		data.Filters
	}

	query := r.URL.Query()
	queryParametersData.Rating = a.getSingleIntegerParameter(query, "rating", 0, nil)
	queryParametersData.Review = a.getSingleQueryParameter(query, "review", "")

	v := validator.New()
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, v)
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "rating", "review", "-id", "-rating", "-review"}

	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Get the reviews from the database
	reviews, metadata, err := a.reviewModel.GetAll(bookID, queryParametersData.Rating, queryParametersData.Review, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	responseData := envelope{
		"reviews":   reviews,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, responseData, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
