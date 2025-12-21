package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("record not found")
	ErrConflict          = errors.New("resource already exists")
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(ctx context.Context, post *Post) error
		GetById(ctx context.Context, id int64) (*Post, error)
		DeletePost(ctx context.Context, id int64) error
		Update(ctx context.Context, post *Post) error
		GetUserFeed(ctx context.Context, id int64, Pag PaginatedFeedQuery) ([]PostWithMetaData, error)
	}
	Users interface {
		Create(ctx context.Context, user *User) error
		GetById(ctx context.Context, id int64) (*User, error)
		GetByEmail(ctx context.Context, email string) (*User, error)
		CreateAndInvite(ctx context.Context, user *User, token string) error
		Activate(ctx context.Context, token string) error
	}
	Comments interface {
		GetById(ctx context.Context, id int64) (*[]Comment, error)
		Create(ctx context.Context, comment *Comment) error
	}
	Followers interface {
		Follow(ctx context.Context, followingId, userId int64) error
		Unfollow(ctx context.Context, followingId, userId int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostsStore{db: db},
		Users:     &UsersStore{db: db},
		Comments:  &CommentsStore{db: db},
		Followers: &FollowersStore{db: db},
	}
}

func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
