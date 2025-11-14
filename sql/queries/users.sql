-- name: CreateUser :one
INSERT INTO
    users (
        id,
        email,
        created_at,
        updated_at,
        hashed_password
    )
VALUES
    (gen_random_uuid(), $1, NOW(), NOW(), $2)
RETURNING
    id,
    email,
    created_at,
    updated_at;

-- name: DeleteUsers :exec
DELETE FROM
    users;

-- name: GetUserWithEmail :one
SELECT
    *
FROM
    users
WHERE
    email = $1;
