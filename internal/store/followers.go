package store

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type FollowersStore struct {
	db *sql.DB
}

type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	Created_at string `json:"created_at"`
}

func (f *FollowersStore) Follow(ctx context.Context, followingId, userId int64) error {
	query := `
	INSERT INTO followers (user_id, follower_id)
	VALUES($1, $2)
	`

	res, err := f.db.ExecContext(ctx, query, userId, followingId)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return err
	}

	return nil

}

func (f *FollowersStore) Unfollow(ctx context.Context, followingId, userId int64) error {
	query := `
	DELETE from followers
	WHERE user_id = $1, follower_id = $2
	`

	res, err := f.db.ExecContext(ctx, query, userId, followingId)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
		return err
	}

	if rows == 0 {
		return err
	}

	return nil

}
