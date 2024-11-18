package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/georgie5/Test3-bookclubapi/internal/data"
	"github.com/georgie5/Test3-bookclubapi/internal/validator"
)

func (a *applicationDependencies) createReadingListHandler(w http.ResponseWriter, r *http.Request) {

	var incomingData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
		CreatedBy   int64  `json:"created_by"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	list := &data.ReadingList{
		Name:        incomingData.Name,
		Description: incomingData.Description,
		Status:      incomingData.Status,
		CreatedBy:   incomingData.CreatedBy,
	}

	// Validate the review data
	v := validator.New()
	data.ValidateReadingList(v, list)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.readingListModel.Insert(list)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/lists/%d", list.ID))

	data := envelope{
		"list": list,
	}

	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) getReadingListHandler(w http.ResponseWriter, r *http.Request) {
	// Get reading list ID from URL
	id, err := a.readIDParam(r, "list_id")
	if err != nil || id < 1 {
		a.notFoundResponse(w, r)
		return
	}

	// Fetch reading list from the database
	list, err := a.readingListModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"list": list,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
}

func (a *applicationDependencies) updateReadingListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r, "list_id")
	if err != nil || id < 1 {
		a.notFoundResponse(w, r)
		return
	}
	list, err := a.readingListModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r) // 404 if not found
		default:
			a.serverErrorResponse(w, r, err) // 500 if any other error
		}
		return
	}

	var incomingData struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
		CreatedBy   *int64  `json:"created_by"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.Name != nil {
		list.Name = *incomingData.Name
	}
	if incomingData.Description != nil {
		list.Description = *incomingData.Description
	}
	if incomingData.Status != nil {
		list.Status = *incomingData.Status
	}

	if incomingData.CreatedBy != nil {
		list.CreatedBy = *incomingData.CreatedBy
	}

	v := validator.New()
	data.ValidateReadingList(v, list)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.readingListModel.Update(list)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"list": list,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) deleteReadingListHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r, "list_id")
	if err != nil || id < 1 {
		a.notFoundResponse(w, r)
		return
	}

	err = a.readingListModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// display the comment
	data := envelope{
		"message": "readinglist successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) addBookToReadingListHandler(w http.ResponseWriter, r *http.Request) {

	listID, err := a.readIDParam(r, "list_id")
	if err != nil || listID < 1 {
		a.notFoundResponse(w, r)
		return
	}

	var incomingData struct {
		BookID int64 `json:"book_id"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.readingListBookModel.AddBook(listID, incomingData.BookID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"message": "book successfully added to reading list",
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) removeBookFromReadingListHandler(w http.ResponseWriter, r *http.Request) {
	listID, err := a.readIDParam(r, "list_id")
	if err != nil || listID < 1 {
		a.notFoundResponse(w, r)
		return
	}
	var incomingData struct {
		BookID int64 `json:"book_id"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	err = a.readingListBookModel.RemoveBook(listID, incomingData.BookID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"message": "book successfully removed from reading list",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

func (a *applicationDependencies) listReadingListsHandler(w http.ResponseWriter, r *http.Request) {
	var queryParametersData struct {
		Name        string
		Description string
		Status      string
		data.Filters
	}

	// Get the query parameters from the URL
	query := r.URL.Query()
	queryParametersData.Name = a.getSingleQueryParameter(query, "name", "")
	queryParametersData.Description = a.getSingleQueryParameter(query, "description", "")
	queryParametersData.Status = a.getSingleQueryParameter(query, "status", "")

	v := validator.New()

	// set pagination and sorting
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(query, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(query, "page_size", 10, v)

	queryParametersData.Filters.Sort = a.getSingleQueryParameter(query, "sort", "id")
	queryParametersData.Filters.SortSafeList = []string{"id", "name", "description", "created_by", "status", "-id", "-name", "-description", "-created_by", "-status"}

	//validate the filters
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	lists, metadata, err := a.readingListModel.GetAll(queryParametersData.Name, queryParametersData.Description, queryParametersData.Status, queryParametersData.Filters)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"lists":     lists,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
