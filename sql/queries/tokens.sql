-- name: CreateToken :one
INSERT INTO
    refresh_tokens(
        token,
        created_at,
        updated_at,
        user_id,
        expires_at
    )
VALUES
    ($1, NOW(), NOW(), $2, $3)
RETURNING
    *;

-- name: GetToken :one
SELECT
    token,
    expires_at,
    revoked_at
FROM
    refresh_tokens
WHERE
    token = $1;

-- name: GetUserForToken :one
SELECT
    users.id
FROM
    refresh_tokens
    LEFT JOIN users ON refresh_tokens.user_id = users.id
WHERE
    refresh_tokens.token = $1;

-- name: RevokeToken :exec
UPDATE
    refresh_tokens
SET
    revoked_at = NOW(),
    updated_at = NOW()
WHERE
    token = $1;
