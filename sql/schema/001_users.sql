-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(email)
);

-- +goose Down
DROP TABLE users;
