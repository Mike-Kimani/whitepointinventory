-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, name, api_key, password, email)
VALUES($1, $2, $3, $4, 
encode(sha256(random()::text::bytea), 'hex'), $5, $6
)
RETURNING *;

-- name: GetUserByAPIKey :one
SELECT * FROM users WHERE api_key = $1;

-- name: GetUserByPasswordAndEmail :one
SELECT * FROM users WHERE password = $1 AND email = $2;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY name ASC;