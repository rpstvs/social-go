package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err)

	RespondWithError(http.StatusInternalServerError, w, "internal server error")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err)

	RespondWithError(http.StatusBadRequest, w, err.Error())
}

func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err)

	RespondWithError(http.StatusForbidden, w, "forbidden")
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnf("response not found", "method", r.Method, "path", r.URL.Path, "error", err)
	RespondWithError(http.StatusNotFound, w, "not found")
}

func (app *application) UnauthorizedErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err)
	RespondWithError(http.StatusUnauthorized, w, "unauthorized")
}

func (app *application) UnauthorizedBasicErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err)

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	RespondWithError(http.StatusUnauthorized, w, "unauthorized")
}

func (app *application) rateLimitExceedResponse(w http.ResponseWriter, r *http.Request, retryAfter string) {

	app.logger.Warnf("rate limit exceeded", "method", r.Method, "path", r.URL.Path, "error", retryAfter)

	w.Header().Set("Retry-After", retryAfter)

	RespondWithError(http.StatusTooManyRequests, w, retryAfter)
}
