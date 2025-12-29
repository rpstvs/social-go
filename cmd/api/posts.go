package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rpstvs/social/internal/store"
)

type postKey string

const postCtxValue postKey = "post"

type CreatePostPayload struct {
	Title   string   `json:"title" validator:"required, max=100"`
	Content string   `json:"content" validator:"required, max=1000"`
	Tags    []string `json:"tags"`
}

func (app *application) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload

	if err := ReadJson(w, r, &payload); err != nil {
		RespondWithError(http.StatusBadRequest, w, err.Error())
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := getUserFromContext(r)

	post := store.Post{
		Title:   payload.Title,
		Content: payload.Title,
		UserID:  user.ID,
		Tags:    payload.Tags,
	}

	if err := app.store.Posts.Create(r.Context(), &post); err != nil {
		RespondWithError(http.StatusBadRequest, w, err.Error())
		return
	}

	if err := RespondWithJson(http.StatusCreated, w, post); err != nil {
		RespondWithError(http.StatusBadRequest, w, err.Error())
		return
	}
}

func (app *application) GetPostHandler(w http.ResponseWriter, r *http.Request) {

	post := getPostfromCtx(r)

	comments, err := app.store.Comments.GetById(r.Context(), post.ID)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	post.Comments = *comments
	err = RespondWithJson(http.StatusOK, w, post)

	if err != nil {
		RespondWithError(http.StatusInternalServerError, w, err.Error())
		return
	}

}

func (app *application) DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "postID")
	id, err := strconv.ParseInt(idParam, 10, 64)

	if err != nil {
		RespondWithError(http.StatusInternalServerError, w, err.Error())
		return
	}
	err = app.store.Posts.DeletePost(r.Context(), id)

	if err != nil {
		app.internalServerError(w, r, err)
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdatePostPayload struct {
	Title   *string `json:"title" validate:"omitempty, max=100"`
	Content *string `json:"content" validate:"omitempty, max=1000"`
}

func (app *application) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostfromCtx(r)

	var payload UpdatePostPayload

	if err := ReadJson(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if payload.Content != nil {
		post.Content = *payload.Content
	}

	if payload.Title != nil {
		post.Title = *payload.Title
	}

	err := app.store.Posts.Update(r.Context(), post)

	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := RespondWithJson(http.StatusCreated, w, post); err != nil {
		RespondWithError(http.StatusBadRequest, w, err.Error())
		return
	}
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "postID")

		id, err := strconv.ParseInt(idParam, 10, 64)

		if err != nil {
			RespondWithError(http.StatusInternalServerError, w, err.Error())
			return
		}
		ctx := r.Context()
		post, err := app.store.Posts.GetById(r.Context(), id)

		if err != nil {

			switch {
			case errors.Is(err, store.ErrNotFound):
				RespondWithError(http.StatusNotFound, w, err.Error())
				return
			default:
				RespondWithError(http.StatusInternalServerError, w, err.Error())
				return
			}

		}

		ctx = context.WithValue(ctx, postCtxValue, post)
		next.ServeHTTP(w, r)
	})

}

func getPostfromCtx(r *http.Request) *store.Post {
	post := r.Context().Value(postCtxValue).(*store.Post)
	return post
}
