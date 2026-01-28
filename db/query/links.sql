-- name: CreateLink :one
INSERT INTO links (original_url, short_name, short_url)
VALUES ($1, $2, $3);

-- name: GetLinkByID :one
SELECT * FROM links
WHERE id = $1;

-- name: ListLinks :many
SELECT * FROM links;

-- name: UpdateLink :one
UPDATE links
SET 
    original_url = COALESCE($2, original_url),
    short_name = COALESCE($3, short_name),
    short_url = COALESCE($4, short_url)
WHERE id = $1;

-- name: DeleteLink :exec
DELETE FROM links
WHERE id = $1;