-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_invitations(
   token bytea PRIMARY KEY,
   user_id bigint NOT NULL,
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_invitations;
-- +goose StatementEnd