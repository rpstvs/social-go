package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

			username := app.config.authConfig.basic.username
			password := app.config.authConfig.basic.password

			creds := strings.SplitN(string(decoded), ":", 2)

			if len(creds) != 0 || creds[0] != username || creds[1] != password {
				app.UnauthorizedErrorResponse(w, r, fmt.Errorf("bad credentials"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) AuthTokenMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				app.UnauthorizedBasicErrorResponse(w, r, fmt.Errorf("no auth header provided"))
				return
			}

			parts := strings.Split(authHeader, " ")

			if len(parts) != 2 || parts[0] != "Bearer" {
				app.UnauthorizedErrorResponse(w, r, fmt.Errorf("auth header malformed"))
				return
			}

			jwtToken, err := app.authenticator.ValidateToken(parts[1])

			if err != nil {
				app.UnauthorizedErrorResponse(w, r, err)
				return
			}

			claims := jwtToken.Claims.(jwt.MapClaims)

			userId, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)

			if err != nil {
				app.UnauthorizedErrorResponse(w, r, err)
				return
			}

			ctx := r.Context()

			user, err := app.store.Posts.GetById(ctx, userId)

			if err != nil {
				app.UnauthorizedErrorResponse(w, r, err)
				return
			}

			ctx = context.WithValue(ctx, userCtx, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
