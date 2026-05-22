-- objects.sql
-- 知识对象 CRUD 查询

-- name: ListObjects :many
SELECT * FROM objects
WHERE space_id = ? AND is_deleted = 0
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: ListObjectsByType :many
SELECT * FROM objects
WHERE space_id = ? AND type_id = ? AND is_deleted = 0
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: GetObject :one
SELECT * FROM objects
WHERE id = ? AND space_id = ? AND is_deleted = 0;

-- name: CreateObject :one
INSERT INTO objects (
    id, space_id, type_id, title, properties, tags,
    cover_image, source, source_meta, word_count,
    version, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?,
    1, ?, ?
) RETURNING *;

-- name: UpdateObject :one
UPDATE objects
SET title = ?, properties = ?, tags = ?, cover_image = ?,
    word_count = ?, version = version + 1, updated_at = ?
WHERE id = ? AND space_id = ? AND is_deleted = 0
RETURNING *;

-- name: SoftDeleteObject :exec
UPDATE objects
SET is_deleted = 1, updated_at = ?
WHERE id = ? AND space_id = ?;

-- name: ArchiveObject :exec
UPDATE objects
SET is_archived = CASE WHEN is_archived = 0 THEN 1 ELSE 0 END,
    updated_at = ?
WHERE id = ? AND space_id = ?;

-- name: DeleteObjectPermanently :exec
DELETE FROM objects WHERE id = ? AND space_id = ?;

-- name: SearchObjects :many
SELECT * FROM objects
WHERE space_id = ? AND is_deleted = 0
  AND title LIKE '%' || ? || '%'
ORDER BY updated_at DESC
LIMIT ? OFFSET ?;

-- name: CountObjectsBySpace :one
SELECT COUNT(*) FROM objects
WHERE space_id = ? AND is_deleted = 0;

-- name: CountObjectsByType :one
SELECT COUNT(*) FROM objects
WHERE space_id = ? AND type_id = ? AND is_deleted = 0;
