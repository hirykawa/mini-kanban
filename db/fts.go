package db

import (
	"context"
	"database/sql"
)

// TaskRow represents a task with project name joined.
type TaskRow struct {
	ID          int64
	ProjectID   int64
	Title       string
	Body        string
	Status      string
	DueAt       *int64
	CreatedAt   int64
	UpdatedAt   int64
	DoneAt      *int64
	ProjectName string
}

// SearchTasksFTS performs FTS5 full-text search on tasks.
func SearchTasksFTS(ctx context.Context, db *sql.DB, projectID int64, query string) ([]TaskRow, error) {
	const q = `
SELECT t.id, t.project_id, t.title, t.body, t.status, t.due_at, t.created_at, t.updated_at, t.done_at, p.name as project_name
FROM tasks t
JOIN projects p ON t.project_id = p.id
WHERE t.project_id = ?
  AND t.id IN (SELECT rowid FROM tasks_fts WHERE tasks_fts MATCH ?)
ORDER BY t.created_at DESC
`
	rows, err := db.QueryContext(ctx, q, projectID, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TaskRow
	for rows.Next() {
		var r TaskRow
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.Title, &r.Body, &r.Status, &r.DueAt, &r.CreatedAt, &r.UpdatedAt, &r.DoneAt, &r.ProjectName); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
