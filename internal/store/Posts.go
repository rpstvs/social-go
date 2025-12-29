package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   string    `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetaData struct {
	Post
	CommentCount int `json:"comment_count"`
}

type PostsStore struct {
	db *sql.DB
}

func (s *PostsStore) Create(ctx context.Context, post *Post) error {
	query := `
	INSERT INTO posts (content, title, user_id, tags)
	VALUES($1,$2,$3,$4)
	RETURNING id, created_at, updated_at
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, post.Content, post.Title, post.UserID, pq.Array(post.Tags)).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostsStore) GetById(ctx context.Context, id int64) (*Post, error) {
	var post Post

	query := `
	SELECT * from posts
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.Content, &post.Title, &post.UserID, pq.Array(&post.Tags), &post.Version, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			{
				return nil, ErrNotFound
			}
		default:
			return nil, err
		}
	}
	return &post, nil
}

func (s *PostsStore) DeletePost(ctx context.Context, id int64) error {
	query := `
		DELETE FROM posts
		WHERE id = $1;
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostsStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content =$2, version = version +1
		WHERE id = $3 AND version = $4
		RETURNING version;
	`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}

	}

	return nil
}

func (s *PostsStore) GetUserFeed(ctx context.Context, id int64, Pag PaginatedFeedQuery) ([]PostWithMetaData, error) {
	query := `
	SELECT p.id, p.user_id, p.content, p.created_at, p.version,p.tags,
	COUNT(c.id) AS comments_counts 
	FROM posts p
	LEFT JOIN comments c ON c.post_id = p.id
	LEFT JOIN users u ON p user_id = u.id
	JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
	WHERE 
		f.user_id = $1 OR p.user_id = $1 AND 
		(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
		(p.tags @> $5 OR $5 = '{}')
	GROUP BY p.id
	ORDER BY p.creadted_at ` + Pag.Sort + `
	LIMIT $2 OFFSET $3;`

	rows, err := s.db.QueryContext(ctx, query, id, Pag.Limit, Pag.Offset, Pag.Search, pq.Array(Pag.Tags))

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var feed []PostWithMetaData

	for rows.Next() {
		var p PostWithMetaData

		err = rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(p.Tags),
			&p.User.Username,
			&p.CommentCount,
		)

		if err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}
	return feed, nil
}
