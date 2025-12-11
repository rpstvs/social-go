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

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logger.Warnf("response not found", "method", r.Method, "path", r.URL.Path, "error", err)
	RespondWithError(http.StatusNotFound, w, "not found")
}
