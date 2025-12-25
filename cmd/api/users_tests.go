package main

import (
	"net/http"
	"testing"
)

func TestGetUser(t *testing.T) {
	app := NewTestApplication(t)
	mux := app.mount()
	testToken, err := app.authenticator.GenerateToken(nil)

	if err != nil {
		t.Fatal(err)
	}
	t.Run("should not allow unauthenticated request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)

		if err != nil {
			t.Fatal(err)
		}

		rr := execute(mux, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected the response to be %d and got %d \n", http.StatusUnauthorized, rr.Code)
		}

	})

	t.Run("should allow authenticated request", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)

		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)
		rr := execute(mux, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected the response to be %d and got %d \n", http.StatusUnauthorized, rr.Code)
		}

	})
}
