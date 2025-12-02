package store

import (
	"context"
	"database/sql"
	"errors"
)

type Comment struct {
	ID         int64  `json:"id"`
	PostID     int64  `json:"post_id"`
	UserID     int64  `json:"user_id"`
	Content    string `json:"content"`
	Created_at string `json:"created_at"`
	User       User   `json:"user"`
}

type CommentsStore struct {
	db *sql.DB
}

func (s *CommentsStore) GetById(ctx context.Context, postid int64) (*[]Comment, error) {

	query := `
	SELECT c.id, c.post_id, c.content, c.created_at, users.username, users.id from comments c
	JOIN users on users.id = comments.user_id
	WHERE comments.post_id = $1
	ORDER BY comments.created_at DESC`

	sqlRows, err := s.db.QueryContext(ctx, query, postid)

	if err != nil {
		return nil, err
	}

	Comments := []Comment{}
	defer sqlRows.Close()

	for sqlRows.Next() {
		var Comment Comment
		Comment.User = User{}
		err := sqlRows.Scan(&Comment.ID, &Comment.PostID, &Comment.Content, &Comment.Created_at, &Comment.User.Username, &Comment.User.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		Comments = append(Comments, Comment)
	}
	return &Comments, nil
}
