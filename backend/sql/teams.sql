-- name: ListTeams :many
SELECT t.id, t.code, t.name, c.code AS category_code, c.name AS category_name
FROM teams t
JOIN categories c ON c.id = t.category_id
ORDER BY c.id, t.name;

-- name: TeamExists :one
SELECT EXISTS (SELECT 1 FROM teams WHERE id = $1);
