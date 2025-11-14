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

-- name: GetUserWithId :one
SELECT
    *
FROM
    users
WHERE
    id = $1;

-- name: UpdateUserEmailAndPassword :one
UPDATE
    users
SET
    email = $2,
    hashed_password = $3,
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    id,
    email,
    created_at,
    updated_at;
