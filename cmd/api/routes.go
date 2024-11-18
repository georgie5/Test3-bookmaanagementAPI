package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	// setup a new router
	router := httprouter.New()
	// handle 404
	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	// handle 405
	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)
	// setup product routes
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthcheckHandler)

	//Books routes
	//Books routes
	router.HandlerFunc(http.MethodGet, "/api/v1/books/search", a.searchBooksHandler)
	router.HandlerFunc(http.MethodGet, "/v1/books", a.listBooksHandler)
	router.HandlerFunc(http.MethodPost, "/v1/books", a.createBookHandler)
	router.HandlerFunc(http.MethodGet, "/v1/books/:book_id", a.getBookHandler)
	router.HandlerFunc(http.MethodPut, "/v1/books/:book_id", a.updateBookHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/books/:book_id", a.deleteBookHandler)

	// Reading lists routes
	router.HandlerFunc(http.MethodGet, "/api/v1/lists", a.listReadingListsHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/lists/:list_id", a.getReadingListHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/lists", a.createReadingListHandler)
	router.HandlerFunc(http.MethodPut, "/api/v1/lists/:list_id", a.updateReadingListHandler)
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:list_id", a.deleteReadingListHandler)
	router.HandlerFunc(http.MethodPost, "/api/v1/lists/:list_id/books", a.addBookToReadingListHandler)
	router.HandlerFunc(http.MethodDelete, "/api/v1/lists/:list_id/books", a.removeBookFromReadingListHandler)

	// Reviews routes
	router.HandlerFunc(http.MethodGet, "/v1/books/:book_id/reviews", a.getReviewsForBookHandler) // Get reviews for a specific book
	router.HandlerFunc(http.MethodPost, "/v1/books/:book_id/reviews", a.addReviewHandler)        // Add a new review to a specific book
	router.HandlerFunc(http.MethodPut, "/v1/reviews/:review_id", a.updateReviewHandler)          // Update a review
	router.HandlerFunc(http.MethodDelete, "/v1/reviews/:review_id", a.deleteReviewHandler)       // Delete a review

	// Users routes
	router.HandlerFunc(http.MethodPost, "/v1/users", a.registerUserHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users/:user_id", a.getUserProfileHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users/:user_id/lists", a.getUserReadingListsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users/:user_id/reviews", a.getUserReviewsHandler)

	router.HandlerFunc(http.MethodPut, "/v1/users/activated", a.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", a.createAuthenticationTokenHandler)

	// Request sent first to recoverPanic() then sent to rateLimit()
	// finally it is sent to the router.
	return a.recoverPanic(a.rateLimit(a.authenticate(router)))

}
