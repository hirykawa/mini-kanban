-- name: CreateProject :one
INSERT INTO projects (name) VALUES (?) RETURNING *;

-- name: GetProjectByName :one
SELECT * FROM projects WHERE name = ?;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = ?;

-- name: ListProjects :many
SELECT * FROM projects ORDER BY name;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;

-- name: CountTasksByProject :one
SELECT COUNT(*) FROM tasks WHERE project_id = ?;

-- name: CreateTask :one
INSERT INTO tasks (project_id, title, body, status, due_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetTask :one
SELECT t.*, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
WHERE t.id = ? AND t.project_id = ?;

-- name: UpdateTask :one
UPDATE tasks SET
    title = COALESCE(sqlc.narg('title'), title),
    body = COALESCE(sqlc.narg('body'), body),
    due_at = CASE WHEN sqlc.narg('clear_due') = 1 THEN NULL ELSE COALESCE(sqlc.narg('due_at'), due_at) END,
    updated_at = unixepoch()
WHERE id = ? AND project_id = ?
RETURNING *;

-- name: MarkTaskDone :one
UPDATE tasks SET
    status = 'done',
    done_at = unixepoch(),
    updated_at = unixepoch()
WHERE id = ? AND project_id = ?
RETURNING *;

-- name: MarkTaskOpen :one
UPDATE tasks SET
    status = 'open',
    done_at = NULL,
    updated_at = unixepoch()
WHERE id = ? AND project_id = ?
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = ? AND project_id = ?;

-- name: ListTasks :many
SELECT t.*, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
WHERE t.project_id = ?
ORDER BY
    CASE WHEN t.status = 'open' THEN 0 ELSE 1 END,
    t.created_at DESC;

-- name: ListTasksOpen :many
SELECT t.*, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
WHERE t.project_id = ? AND t.status = 'open'
ORDER BY t.created_at DESC;

-- name: ListTasksDone :many
SELECT t.*, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
WHERE t.project_id = ? AND t.status = 'done'
ORDER BY t.done_at DESC;

-- name: SearchTasksLike :many
SELECT t.*, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
WHERE t.project_id = ?
  AND (t.title LIKE '%' || ? || '%' OR t.body LIKE '%' || ? || '%')
ORDER BY t.created_at DESC;

-- name: GetOrCreateTag :one
INSERT INTO tags (name) VALUES (?)
ON CONFLICT (name) DO UPDATE SET name = excluded.name
RETURNING *;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = ?;

-- name: ListTagsForTask :many
SELECT t.* FROM tags t
JOIN task_tags tt ON t.id = tt.tag_id
WHERE tt.task_id = ?
ORDER BY t.name;

-- name: AddTagToTask :exec
INSERT OR IGNORE INTO task_tags (task_id, tag_id) VALUES (?, ?);

-- name: RemoveTagFromTask :exec
DELETE FROM task_tags WHERE task_id = ? AND tag_id = ?;

-- name: RemoveAllTagsFromTask :exec
DELETE FROM task_tags WHERE task_id = ?;

-- name: ListTasksByTag :many
SELECT t.*, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
JOIN task_tags tt ON t.id = tt.task_id
JOIN tags tg ON tt.tag_id = tg.id
WHERE t.project_id = ? AND tg.name = ?
ORDER BY t.created_at DESC;

-- name: TopTagsForProject :many
SELECT tg.name, COUNT(*) as usage_count
FROM tags tg
JOIN task_tags tt ON tg.id = tt.tag_id
JOIN tasks t ON tt.task_id = t.id
WHERE t.project_id = ?
GROUP BY tg.id
ORDER BY usage_count DESC
LIMIT ?;

-- name: ListTasksWithDue :many
SELECT t.*, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
WHERE t.project_id = ? AND t.due_at IS NOT NULL
ORDER BY t.due_at ASC;

-- name: ListOverdueTasks :many
SELECT t.*, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
WHERE t.project_id = ? AND t.status = 'open' AND t.due_at < unixepoch()
ORDER BY t.due_at ASC;
