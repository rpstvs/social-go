-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS users
ADD COLUMN role_id INT REFERENCES roles(id) DEFAULT 1;
UPDATE users
SET role_id = (
        SELECT id
        from roles
        WHERE name = 'user'
    );
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN role_id;
-- +goose StatementEnd