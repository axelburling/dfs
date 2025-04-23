-- name: InsertObject :one
INSERT INTO objects (name, total_size, total_chunks, content_type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetObjectByID :one
SELECT * FROM objects WHERE id = $1;

-- name: GetObjectByName :one
SELECT * FROM objects WHERE name = $1;

-- name: GetAllObjects :many
SELECT * FROM objects ORDER BY name ASC;

-- name: UpdateObjectSizeAndChunks :exec
UPDATE objects
SET total_size = $2, total_chunks = $3
WHERE id = $1;

-- name: DeleteObject :exec
DELETE FROM objects WHERE id = $1;
