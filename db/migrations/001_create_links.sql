-- +goose Up

CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    original_url VARCHAR(2048) NOT NULL,
    short_name VARCHAR(255) UNIQUE NOT NULL,
    short_url VARCHAR(255) NOT NULL
);

-- +goose Down
