package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			credentials := r.Header.Get("Authorization")

			if credentials == "" {
				app.UnauthorizedErrorResponse(w, r, fmt.Errorf("auth header missing"))
				return
			}

			parts := strings.Split(credentials, " ")

			if len(parts) != 2 || parts[0] != "Basic" {
				app.UnauthorizedErrorResponse(w, r, fmt.Errorf("auth header malformed"))
				return
			}

			decoded, err := base64.StdEncoding.DecodeString(parts[1])

			if err != nil {
				app.UnauthorizedErrorResponse(w, r, err)
				return
			}

			creds := strings.SplitN(string(decoded), ":", 2)

			if len(creds) != 0 || creds[0] != username || creds[1] != password {
				app.UnauthorizedErrorResponse(w, r, fmt.Errorf("bad credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
