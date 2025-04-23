-- name: InsertNode :one
INSERT INTO nodes (id, address, grpc_address, hostname, last_seen, is_healthy, total_space, free_space, readonly, controller_master_id)
VALUES ($1, $2, $3, $4, NOW(), $5, $6, $7, $8, NULL) RETURNING *;


-- name: GetAllNodes :many
SELECT * FROM nodes;

-- name: GetNodeByID :one
SELECT * FROM nodes WHERE id = $1;

-- name: GetHealthyNodes :many
SELECT * FROM nodes WHERE is_healthy = true ORDER BY free_space DESC;

-- name: UpdateNodeFreeSpace :exec
UPDATE nodes SET free_space = $2 WHERE id = $1;

-- name: UpdateNodeHealth :exec
UPDATE nodes SET is_healthy = $2 WHERE id = $1;

-- name: UpdateNodeControllerMaster :exec
UPDATE nodes SET controller_master_id = $2 WHERE id = $1;

-- name: DeleteNode :exec
DELETE FROM nodes WHERE id = $1;