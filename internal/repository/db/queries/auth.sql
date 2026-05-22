-- auth.sql
-- 认证相关查询

-- name: CreateAccount :one
INSERT INTO accounts (
    id, email, name, password_hash, locale, timezone, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetAccountByEmail :one
SELECT * FROM accounts
WHERE email = ? AND is_active = 1;

-- name: GetAccountByID :one
SELECT * FROM accounts WHERE id = ?;

-- name: UpdateLastLogin :exec
UPDATE accounts SET last_login_at = ?, updated_at = ? WHERE id = ?;

-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (id, account_id, token_hash, device_id, expires_at, created_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE id = ? AND revoked = 0 AND expires_at > ?;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked = 1 WHERE id = ?;

-- name: RevokeAllAccountTokens :exec
UPDATE refresh_tokens SET revoked = 1
WHERE account_id = ? AND revoked = 0;

-- name: ListAccountsByIDs :many
SELECT * FROM accounts WHERE id IN (sqlc.slice('ids'));

-- name: CreateSpace :one
INSERT INTO spaces (id, name, type, account_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?) RETURNING *;

-- name: GetPersonalSpace :one
SELECT * FROM spaces
WHERE account_id = ? AND type = 'personal' AND is_deleted = 0
LIMIT 1;

-- name: CreateDefaultTypes :exec
INSERT INTO object_types (id, space_id, name, icon, description, is_builtin, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 1, ?, ?);
