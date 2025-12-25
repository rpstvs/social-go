package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rpstvs/social/internal/store"
)

type userKey string

var CTX_USER_KEY userKey = "user"

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "userId")

	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil || id < 1 {
		app.badRequestResponse(w, r, err)
		return
	}

	user, err := app.getUser(r.Context(), id)

	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			{
				app.badRequestResponse(w, r, err)
				return
			}
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	if err := app.jsonResponse(w, http.StatusAccepted, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err = app.store.Followers.Follow(r.Context(), user.ID, int64(followedID))

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	err = app.store.Followers.Unfollow(r.Context(), user.ID, int64(unfollowedID))

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)

	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
}

func (app *application) UserHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "userId")

		id, err := strconv.ParseInt(idParam, 10, 64)

		if err != nil {
			RespondWithError(http.StatusInternalServerError, w, err.Error())
			return
		}
		ctx := r.Context()
		userHandler, err := app.store.Users.GetById(ctx, id)

		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				{
					app.badRequestResponse(w, r, err)
					return
				}
			default:
				app.internalServerError(w, r, err)
				return
			}
		}
		ctx = context.WithValue(r.Context(), CTX_USER_KEY, userHandler)
		next.ServeHTTP(w, r)
	})
}

func getUserFromContext(r *http.Request) *store.User {
	user := r.Context().Value(CTX_USER_KEY).(*store.User)

	return user
}
