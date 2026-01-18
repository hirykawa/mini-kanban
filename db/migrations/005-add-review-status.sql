-- Migration 005: Add review status
-- Changes status from ('todo', 'doing', 'done') to ('todo', 'doing', 'review', 'done')

-- Step 1: Create new table with updated constraint
CREATE TABLE IF NOT EXISTS tasks_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'doing', 'review', 'done')),
    due_at INTEGER,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    done_at INTEGER
);

-- Step 2: Copy data
INSERT INTO tasks_new (id, project_id, title, body, status, due_at, created_at, updated_at, done_at)
SELECT id, project_id, title, body, status, due_at, created_at, updated_at, done_at
FROM tasks;

-- Step 3: Drop old table and rename new one
DROP TABLE tasks;
ALTER TABLE tasks_new RENAME TO tasks;

-- Step 4: Recreate indexes
CREATE INDEX IF NOT EXISTS idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_due_at ON tasks(due_at) WHERE due_at IS NOT NULL;

-- Step 5: Recreate FTS triggers
DROP TRIGGER IF EXISTS tasks_ai;
DROP TRIGGER IF EXISTS tasks_ad;
DROP TRIGGER IF EXISTS tasks_au;

CREATE TRIGGER tasks_ai AFTER INSERT ON tasks BEGIN
    INSERT INTO tasks_fts(rowid, title, body) VALUES (new.id, new.title, new.body);
END;

CREATE TRIGGER tasks_ad AFTER DELETE ON tasks BEGIN
    INSERT INTO tasks_fts(tasks_fts, rowid, title, body) VALUES('delete', old.id, old.title, old.body);
END;

CREATE TRIGGER tasks_au AFTER UPDATE ON tasks BEGIN
    INSERT INTO tasks_fts(tasks_fts, rowid, title, body) VALUES('delete', old.id, old.title, old.body);
    INSERT INTO tasks_fts(rowid, title, body) VALUES (new.id, new.title, new.body);
END;

-- Record migration
INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (005, '005-add-review-status');
