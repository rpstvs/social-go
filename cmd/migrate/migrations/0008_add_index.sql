-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_comments_content ON comments USING gin(content gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_title ON posts USING gin(title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_users_usernames ON users USING gin(username);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts USING gin(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_posts_id ON posts USING gin(comments);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF NOT EXISTS idx_comments_content;
DROP INDEX IF NOT EXISTS idx_posts_title;
DROP INDEX IF NOT EXISTS idx_posts_tags;
DROP INDEX IF NOT EXISTS idx_users_usernames;
DROP INDEX IF NOT EXISTS idx_posts_user_id;
DROP INDEX IF NOT EXISTS idx_comments_posts_id;
-- +goose StatementEnd