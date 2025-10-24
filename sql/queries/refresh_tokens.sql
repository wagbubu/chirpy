-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at)
VALUES (
  $1,
  NOW(),
  NOW(),
  $2,
  $3
);

-- name: GetUserRefreshToken :one
SELECT user_id FROM refresh_tokens 
WHERE token = $1 AND expires_at > NOW() AND revoked_at IS NULL;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE token = $1;