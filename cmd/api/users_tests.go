package main

import (
	"net/http"
	"testing"

	"github.com/rpstvs/social/internal/store/cache"
	"github.com/stretchr/testify/mock"
)

func TestGetUser(t *testing.T) {
	withRedis := config{
		redisCfg: RedisConfig{
			enabled: true,
		},
	}
	app := NewTestApplication(t, withRedis)
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

	t.Run("should hit the cache first and if not exists it sets the user on the cache", func(t *testing.T) {
		mockCacheStore := app.cacheStorage.Users.(*cache.MockCacheStorage)

		mockCacheStore.On("Get", int64(42)).Return(nil, nil)
		mockCacheStore.On("Get", int64(1)).Return(nil, nil)
		mockCacheStore.On("Set", mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)

		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)
		rr := execute(mux, req)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.AssertCalled(t, "Get", 2)

		mockCacheStore.Calls = nil

	})

	t.Run("should hit the cache first and if not exists it sets the user on the cache", func(t *testing.T) {
		app.config.redisCfg.enabled = false
		mockCacheStore := app.cacheStorage.Users.(*cache.MockCacheStorage)

		mockCacheStore.On("Get", int64(42)).Return(nil, nil)
		mockCacheStore.On("Get", int64(1)).Return(nil, nil)
		mockCacheStore.On("Set", mock.Anything, mock.Anything).Return(nil)

		req, err := http.NewRequest(http.MethodGet, "/v1/users/1", nil)

		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+testToken)
		rr := execute(mux, req)

		checkResponseCode(t, http.StatusOK, rr.Code)

		mockCacheStore.AssertCalled(t, "Get", 2)

		mockCacheStore.Calls = nil

	})
}
