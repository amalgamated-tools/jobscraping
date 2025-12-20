-- name: GetJob :one
SELECT * FROM jobs
WHERE id = ? LIMIT 1;

-- name: ListJobs :many
SELECT * FROM jobs
ORDER BY id;

-- name: CreateJob :one
INSERT INTO jobs (absolute_url, data)
VALUES (?, ?)
RETURNING *;