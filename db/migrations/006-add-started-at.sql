-- Add started_at column to tasks table
-- Records when a task was first started (status changed to 'doing')

ALTER TABLE tasks ADD COLUMN started_at INTEGER;

-- Record migration
INSERT OR IGNORE INTO migrations (migration_number, migration_name)
VALUES (006, '006-add-started-at');
