package main

import "net/http"

func (app *application) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {

	if err := RespondWithJson(200, w, "OK"); err != nil {
		RespondWithError(http.StatusInternalServerError, w, err.Error())
	}

}
