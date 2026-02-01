-- +goose Up

CREATE TABLE links (
    id SERIAL PRIMARY KEY,
    original_url VARCHAR(2048) NOT NULL,
    short_name VARCHAR(255) UNIQUE NOT NULL,
    short_url VARCHAR(255) NOT NULL
);
CREATE TABLE link_visits (
    id SERIAL PRIMARY KEY,
    link_id INTEGER NOT NULL,
    ip VARCHAR(45),
    user_agent TEXT,
    status INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
DROP TABLE link_visits;
DROP TABLE links;