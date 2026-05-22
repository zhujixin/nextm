-- blocks.sql
-- 内容块查询

-- name: ListBlocksByObject :many
SELECT * FROM blocks
WHERE object_id = ?
ORDER BY position ASC;

-- name: GetBlock :one
SELECT * FROM blocks WHERE id = ?;

-- name: CreateBlock :one
INSERT INTO blocks (
    id, object_id, parent_id, type, content, properties,
    position, depth, color, version, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, 1, ?, ?
) RETURNING *;

-- name: UpdateBlock :one
UPDATE blocks
SET content = ?, properties = ?, type = ?,
    position = ?, depth = ?, color = ?,
    version = version + 1, updated_at = ?
WHERE id = ? AND object_id = ?
RETURNING *;

-- name: DeleteBlock :exec
DELETE FROM blocks WHERE id = ? AND object_id = ?;

-- name: ReorderBlocks :batch
UPDATE blocks SET position = ?, updated_at = ? WHERE id = ? AND object_id = ?;

-- name: CountBlocksByObject :one
SELECT COUNT(*) FROM blocks WHERE object_id = ?;
