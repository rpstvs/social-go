package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
	"github.com/rpstvs/social/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required, max=100"`
	Password string `json:"password" validate:"required, min=3,max=72"`
	Email    string `json:"email" validate:"required, email, max=255"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload

	err := ReadJson(w, r, payload)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	err = Validate.Struct(payload)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	err = user.Password.Set(payload.Password)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))

	hashToken := hex.EncodeToString(hash[:])

	err = app.store.Users.CreateAndInvite(r.Context(), user, hashToken, app.config.mail.exp)

	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}

		return
	}

	UserWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}

	if err := app.jsonResponse(w, http.StatusCreated, UserWithToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
