package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	RespondWithError(http.StatusInternalServerError, w, "internal server error")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	RespondWithError(http.StatusBadRequest, w, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not found resposne error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())

	RespondWithError(http.StatusNotFound, w, "not found")
}
