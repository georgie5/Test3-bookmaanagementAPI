package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/georgie5/Test3-bookclubapi/internal/data"
	"github.com/georgie5/Test3-bookclubapi/internal/validator"
)

type bookResponse struct {
	Title           string    `json:"title"`
	Authors         []string  `json:"authors"`
	ISBN            string    `json:"isbn"`
	PublicationDate time.Time `json:"publication_date"`
	Genre           string    `json:"genre"`
	Description     string    `json:"description"`
	AverageRating   float64   `json:"average_rating"`
	Version         int32     `json:"version"`
}

// create a book handler
func (a *applicationDependencies) createBookHandler(w http.ResponseWriter, r *http.Request) {

	// create a struct to hold a product
	var incomingData struct {
		Title           string   `json:"title"`
		Authors         []string `json:"authors"`
		ISBN            string   `json:"isbn"`
		PublicationDate string   `json:"publication_date"`
		Genre           string   `json:"genre"`
		Description     string   `json:"description"`
	}

	// decode the incoming JSON
	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Parse the date manually using the custom format
	publicationDate, err := time.Parse("2006-01-02", incomingData.PublicationDate)
	if err != nil {
		a.badRequestResponse(w, r, fmt.Errorf("invalid publication_date format, expected YYYY-MM-DD"))
		return
	}

	// create a new Book struct
	book := &data.Book{
		Title: incomingData.Title,
		// Authors:         incomingData.Authors,
		ISBN:            incomingData.ISBN,
		PublicationDate: publicationDate,
		Genre:           incomingData.Genre,
		Description:     incomingData.Description,
	}

	//validate that the authors are not empty
	if len(incomingData.Authors) == 0 {
		a.badRequestResponse(w, r, fmt.Errorf("authors must be provided"))
		return
	}

	v := validator.New()

	data.ValidateBook(v, book)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// insert the book into the database
	err = a.bookModel.Insert(book)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//insert the authors into the book_authors table
	for _, authorName := range incomingData.Authors {
		// Check if the author exists in the database, otherwise create a new one
		author, err := a.AuthorModel.Get(authorName)
		if err != nil {
			switch err {
			case data.ErrRecordNotFound:
				author = &data.Author{Name: authorName}
				err = a.AuthorModel.Insert(author)
				if err != nil {
					a.serverErrorResponse(w, r, err)
					return
				}
			default:
				a.serverErrorResponse(w, r, err)
				return
			}
		}

		// Create a relationship between the book and the author
		err = a.BookAuthorModel.Insert(book.ID, author.ID)
		if err != nil {
			a.serverErrorResponse(w, r, err)
			return
		}
	}

	// Send a response with the created book
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/books/%d", book.ID)) // Location header for RESTful practice

	data := envelope{
		"book": bookResponse{
			Title:           book.Title,
			Authors:         incomingData.Authors,
			ISBN:            book.ISBN,
			PublicationDate: book.PublicationDate,
			Genre:           book.Genre,
			Description:     book.Description,
			AverageRating:   book.AverageRating,
			Version:         book.Version,
		},
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}

// display a book handler
func (a *applicationDependencies) getBookHandler(w http.ResponseWriter, r *http.Request) {

	// get the book ID from the URL
	id, err := a.readIDParam(r, "book_id")
	if err != nil || id < 1 {
		a.notFoundResponse(w, r)
		return
	}

	// fetch the book from the database
	book, authors, err := a.bookModel.Get(id)
	if err != nil {
		switch err {
		case data.ErrRecordNotFound:
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Convert authors slice from []data.Author to []string
	authorNames := make([]string, len(authors))
	for i, author := range authors {
		authorNames[i] = author.Name
	}

	data := envelope{
		"book": bookResponse{
			Title:           book.Title,
			Authors:         authorNames,
			ISBN:            book.ISBN,
			PublicationDate: book.PublicationDate,
			Genre:           book.Genre,
			Description:     book.Description,
			AverageRating:   book.AverageRating,
			Version:         book.Version,
		},
	}

	// send the response
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// update a book handler
func (a *applicationDependencies) updateBookHandler(w http.ResponseWriter, r *http.Request) {

	id, err := a.readIDParam(r, "book_id")
	if err != nil || id < 1 {
		a.notFoundResponse(w, r)
		return
	}

	book, _, err := a.bookModel.Get(id)
	if err != nil {
		switch err {
		case data.ErrRecordNotFound:
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	var incomingData struct {
		Title           *string   `json:"title"`
		Authors         *[]string `json:"authors"`
		ISBN            *string   `json:"isbn"`
		PublicationDate *string   `json:"publication_date"`
		Genre           *string   `json:"genre"`
		Description     *string   `json:"description"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// we need to check if the fields to see which ones need updating
	// if the incomcingData. Content is nil, we don't update it
	if incomingData.Title != nil {
		book.Title = *incomingData.Title
	}
	if incomingData.ISBN != nil {
		book.ISBN = *incomingData.ISBN
	}
	if incomingData.PublicationDate != nil {
		parsedPublicationDate, err := time.Parse("2006-01-02", *incomingData.PublicationDate)
		if err != nil {
			a.badRequestResponse(w, r, fmt.Errorf("invalid publication_date format, expected YYYY-MM-DD"))
			return
		}
		book.PublicationDate = parsedPublicationDate
	}
	if incomingData.Genre != nil {
		book.Genre = *incomingData.Genre
	}
	if incomingData.Description != nil {
		book.Description = *incomingData.Description
	}
	// if authors are provided, we need to update the authors
	if incomingData.Authors != nil {
		// delete all the current authors
		err = a.BookAuthorModel.Delete(id)
		if err != nil {
			a.serverErrorResponse(w, r, err)
			return
		}

		// insert the new authors
		for _, authorName := range *incomingData.Authors {
			// Check if the author exists in the database, otherwise create a new one
			author, err := a.AuthorModel.Get(authorName)
			if err != nil {
				switch err {
				case data.ErrRecordNotFound:
					author = &data.Author{Name: authorName}
					err = a.AuthorModel.Insert(author)
					if err != nil {
						a.serverErrorResponse(w, r, err)
						return
					}
				default:
					a.serverErrorResponse(w, r, err)
					return
				}
			}

			// Create a relationship between the book and the author
			err = a.BookAuthorModel.Insert(book.ID, author.ID)
			if err != nil {
				a.serverErrorResponse(w, r, err)
				return
			}
		}
	}

	// validate the updated book
	v := validator.New()
	data.ValidateBook(v, book)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// update the book in the database
	err = a.bookModel.Update(book)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	//Get updated authors
	var currentAuthors []data.Author
	_, currentAuthors, err = a.bookModel.Get(book.ID)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	authorNames := make([]string, len(currentAuthors))
	for i, author := range currentAuthors {
		authorNames[i] = author.Name
	}

	data := envelope{
		"book": bookResponse{
			Title:           book.Title,
			Authors:         authorNames,
			ISBN:            book.ISBN,
			PublicationDate: book.PublicationDate,
			Genre:           book.Genre,
			Description:     book.Description,
			AverageRating:   book.AverageRating,
			Version:         book.Version,
		},
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

}

// delete a book handler
func (a *applicationDependencies) deleteBookHandler(w http.ResponseWriter, r *http.Request) {

	// get the book ID from the URL
	id, err := a.readIDParam(r, "book_id")
	if err != nil || id < 1 {
		a.notFoundResponse(w, r)
		return
	}

	// First, delete the relationships in the book_authors table
	err = a.BookAuthorModel.Delete(id)
	if err != nil {
		a.serverErrorResponse(w, r, fmt.Errorf("unable to delete book-author relationships: %w", err))
		return
	}

	// delete the book from the database
	err = a.bookModel.Delete(id)
	if err != nil {
		switch err {
		case data.ErrRecordNotFound:
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"message": "book successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// list all books handler
func (a *applicationDependencies) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	var queryParametersData struct {
		Title string
		Genre string
		data.Filters
	}

	// Get query parameters from the URL
	query := r.URL.Query()
	queryParametersData.Title = a.getSingleQueryParameter(query, "title", "")
	queryParametersData.Genre = a.getSingleQueryParameter(query, "genre", "")

	v := validator.New()

	// Set pagination and sorting
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, v)

	// Sort by title, default is "title" for ascending order, "-title" for descending order
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(query, "sort", "title")
	queryParametersData.Filters.SortSafeList = []string{"id", "title", "genre", "-id", "-title", "-genre"}

	// Validate the filters
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Fetch books from the database
	books, metadata, err := a.bookModel.GetAll(queryParametersData.Title, queryParametersData.Genre, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send the JSON response
	data := envelope{
		"books":     books,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// // search books handler
func (a *applicationDependencies) searchBooksHandler(w http.ResponseWriter, r *http.Request) {
	var queryParametersData struct {
		Title  string
		Author string
		Genre  string
		data.Filters
	}

	// Get query parameters from the URL
	query := r.URL.Query()
	queryParametersData.Title = a.getSingleQueryParameter(query, "title", "")
	queryParametersData.Author = a.getSingleQueryParameter(query, "author", "")
	queryParametersData.Genre = a.getSingleQueryParameter(query, "genre", "")

	v := validator.New()

	// Set pagination and sorting
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, v)

	// Sort by title, default is "title" for ascending order, "-title" for descending order
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(query, "sort", "title")
	queryParametersData.Filters.SortSafeList = []string{"id", "title", "author", "genre", "-id", "-title", "-author", "-genre"}

	// Validate the filters
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Fetch books matching search criteria
	books, metadata, err := a.bookModel.Search(queryParametersData.Title, queryParametersData.Author, queryParametersData.Genre, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Send the JSON response
	data := envelope{
		"books":     books,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
