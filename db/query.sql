-- name: GetJob :one
SELECT * FROM jobs
WHERE id = ? LIMIT 1;
