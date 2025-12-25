package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rpstvs/social/internal/auth"
	"github.com/rpstvs/social/internal/store"
	"github.com/rpstvs/social/internal/store/cache"
	"go.uber.org/zap"
)

func NewTestApplication(t *testing.T) *application {
	t.Helper()

	logger := zap.NewNop().Sugar()
	mockStore := store.NewMockStore()
	mockCacheStorage := cache.NewMockCache()
	mockAuthenticator := auth.NewMockAuthenticator()

	return &application{
		logger:        logger,
		store:         mockStore,
		cacheStorage:  mockCacheStorage,
		authenticator: mockAuthenticator,
	}
}

func execute(mux http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("expected the response to be %d and got %d \n", expected, actual)
	}
}
