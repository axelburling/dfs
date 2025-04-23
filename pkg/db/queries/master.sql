-- name: InsertMaster :exec
INSERT INTO masters (id, address, grpc_address, hostname, last_seen, is_healthy)
VALUES ($1, $2, $3, $4, NOW(), $5);

-- name: GetAllMasters :many
SELECT * FROM masters;

-- name: UpdateMasterHealth :exec
UPDATE masters SET is_healthy = $2, last_seen = NOW() WHERE id = $1;

-- name: DeleteMaster :exec
DELETE FROM masters WHERE id = $1;

-- name: GetUnhealthyNodesWithControllers :many
SELECT n.*, m.address AS controller_master_address
FROM nodes n
LEFT JOIN masters m ON n.controller_master_id = m.id
WHERE n.is_healthy = false;