package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/georgie5/Test3-bookclubapi/internal/validator"
	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

func (a *applicationDependencies) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	jsResponse, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	jsResponse = append(jsResponse, '\n')
	//additional headers to be set
	for key, value := range headers {
		w.Header()[key] = value
	}

	//set content type header
	w.Header().Set("Content-Type", "application/json")
	//explicitly set the respnonse status code
	w.WriteHeader(status)
	_, err = w.Write(jsResponse)
	if err != nil {
		return err
	}

	return nil

}

func (a *applicationDependencies) readJSON(w http.ResponseWriter, r *http.Request, destination any) error {

	// what is the max size of the request body (250KB seems reasonable)
	maxBytes := 256_000
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	// our decoder will check for unknown fields
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	// let start the decoding
	err := dec.Decode(destination)

	// err := json.NewDecoder(r.Body).Decode(destination) // delete this line
	if err != nil {
		// check for the different errors
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("the body contains badly-formed (at character %d)", syntaxError.Offset)
			// Decode can also send back an io error message
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("the body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("the body contains the incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("the body contains the incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("the body must not be empty")

		// check for unknown field error
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(),
				"json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// does the body exceed our limit of 250KB?
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("the body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.Is(err, io.EOF):
			return errors.New("the body must not be empty")

		// the programmer messed up
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
			// some other type of error
		default:
			return err
		}
	}
	// almost done. Let's lastly check if there is any data after
	// the valid JSON data. Maybe the person is trying to send
	// multiple request bodies during one request
	// We call decode once more to see if it gives us back anything
	// we use a throw away struct 'struct{}{}' to hold the result
	err = dec.Decode(&struct{}{})

	if !errors.Is(err, io.EOF) { // there is more data present
		return errors.New("the body must only contain a single JSON value")
	}

	return nil
}

func (a *applicationDependencies) readIDParam(r *http.Request, paramName string) (int64, error) {
	// Get the URL parameters
	params := httprouter.ParamsFromContext(r.Context())
	// Convert the specified parameter from string to int
	id, err := strconv.ParseInt(params.ByName(paramName), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid " + paramName + " parameter")
	}

	return id, nil
}
func (a *applicationDependencies) getSingleQueryParameter(queryParameters url.Values, key string, defaultValue string) string {

	// url.Values is a key:value hash map of the query parameters
	result := queryParameters.Get(key)
	if result == "" {
		return defaultValue
	}
	return result
}

// call when we have multiple comma-separated values
// func (a *applicationDependencies) getMultipleQueryParameters(queryParameters url.Values, key string, defaultValue []string) []string {

// 	result := queryParameters.Get(key)
// 	if result == "" {
// 		return defaultValue
// 	}
// 	return strings.Split(result, ",")

// }

// this method can cause a validation error when trying to convert the
// string to a valid integer value
func (a *applicationDependencies) getSingleIntegerParameter(queryParameters url.Values, key string, defaultValue int, v *validator.Validator) int {

	result := queryParameters.Get(key)
	if result == "" {
		return defaultValue
	}
	// try to convert to an integer
	intValue, err := strconv.Atoi(result)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return intValue

}

// Accept a function and run it in the background also recover from any panic
func (a *applicationDependencies) background(fn func()) {
	a.wg.Add(1) // Use a wait group to ensure all goroutines finish before we exit
	go func() {
		defer a.wg.Done() // signal goroutine is done
		defer func() {
			err := recover()
			if err != nil {
				a.logger.Error(fmt.Sprintf("%v", err))
			}
		}()
		fn() // Run the actual function
	}()
}
