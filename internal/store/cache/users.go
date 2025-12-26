package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rpstvs/social/internal/store"
)

type UserStore struct {
	rdb *redis.Client
}

func (u *UserStore) Get(ctx context.Context, id int64) (*store.User, error) {
	cacheKey := fmt.Sprintf("user-%v", id)

	data, err := u.rdb.Get(ctx, cacheKey).Result()

	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var user store.User

	if data != "" {
		err := json.Unmarshal([]byte(data), &user)

		if err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (u *UserStore) Set(ctx context.Context, user *store.User) error {

	cacheKey := fmt.Sprintf("user-%v", user.ID)

	data, err := json.Marshal(user)

	if err != nil {
		return err
	}

	return u.rdb.Set(ctx, cacheKey, data, 24*time.Hour).Err()
}
