package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/georgie5/Test3-bookclubapi/internal/data"
	"github.com/georgie5/Test3-bookclubapi/internal/validator"
)

func (a *applicationDependencies) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get the passed in data from the request body and store in a temporary struct
	var incomingData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Validate the email and password provided by the client.
	v := validator.New()

	data.ValidateEmail(v, incomingData.Email)
	data.ValidatePasswordPlaintext(v, incomingData.Password)

	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Fetch the user record from the database using the email
	user, err := a.userModel.GetByEmail(incomingData.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.invalidCredentialsResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Check if the password provided matches the hashed password in the database
	match, err := user.Password.Matches(incomingData.Password)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}
	if !match {
		a.invalidCredentialsResponse(w, r)
		return
	}

	// Generate a new authentication token which expires in 24 hours
	token, err := a.tokenModel.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"authentication_token": token,
	}

	// Return the bearer token
	err = a.writeJSON(w, http.StatusCreated, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}

}
func (a *applicationDependencies) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
    var incomingData struct {
        Email string `json:"email"`
    }

    err := a.readJSON(w, r, &incomingData)
    if err != nil {
        a.badRequestResponse(w, r, err)
        return
    }

    v := validator.New()
    data.ValidateEmail(v, incomingData.Email)

    if !v.IsEmpty() {
        a.failedValidationResponse(w, r, v.Errors)
        return
    }

    user, err := a.userModel.GetByEmail(incomingData.Email)
    if err != nil {
        switch {
        case errors.Is(err, data.ErrRecordNotFound):
            a.invalidCredentialsResponse(w, r)
        default:
            a.serverErrorResponse(w, r, err)
        }
        return
    }

    if !user.Activated {
        a.inactiveAccountResponse(w, r)
        return
    }

    token, err := a.tokenModel.New(user.ID, 30*time.Minute, data.ScopePasswordReset)
    if err != nil {
        a.serverErrorResponse(w, r, err)
        return
    }

	a.background(func() {
		
		data := envelope{
			"passwordResetToken": token.Plaintext,
		}

		err = a.mailer.Send(user.Email, "password_reset.tmpl", data)
		if err != nil {
			a.logger.Error(err.Error())
		}
	})

    envelope := envelope{
		"message": "an email will be sent to you containing password reset instructions",
	}
    err = a.writeJSON(w, http.StatusAccepted, envelope, nil)
    if err != nil {
        a.serverErrorResponse(w, r, err)
    }
}