-- name: InsertChunk :exec
INSERT INTO chunks (id, chunk_index, chunk_hash)
VALUES ($1, $2, $3);

-- name: GetChunkByID :one
SELECT * FROM chunks WHERE id = $1;

-- name: GetChunksByHash :many
SELECT * FROM chunks WHERE chunk_hash = $1;

-- name: GetChunksByObjectID :many
SELECT * FROM chunks WHERE object_id = $1 ORDER BY chunk_index;

-- name: UpsertChunk :exec
INSERT INTO chunks (id, chunk_index, chunk_hash)
VALUES ($1, $2, $3)
ON CONFLICT (id)
DO UPDATE SET chunk_index = EXCLUDED.chunk_index, chunk_hash = EXCLUDED.chunk_hash;

-- name: DeleteChunkByID :exec
DELETE FROM chunks WHERE id = $1;

-- name: InsertChunkNodeMapping :exec
INSERT INTO chunk_node_locations (chunk_id, node_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: GetNodesForChunk :many
SELECT n.*
FROM nodes n
JOIN chunk_node_locations cnl ON n.id = cnl.node_id
WHERE cnl.chunk_id = $1;

-- name: GetChunksOnNode :many
SELECT c.*
FROM chunks c
JOIN chunk_node_locations cnl ON c.id = cnl.chunk_id
WHERE cnl.node_id = $1;

-- name: DeleteChunkNodeMapping :exec
DELETE FROM chunk_node_locations WHERE chunk_id = $1 AND node_id = $2;
