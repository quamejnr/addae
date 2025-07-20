-- +goose Up
ALTER TABLE tasks ADD COLUMN completed_at TIMESTAMP;

UPDATE tasks SET completed_at = CURRENT_TIMESTAMP WHERE status = 'complete';

ALTER TABLE tasks DROP COLUMN status;

-- +goose Down
ALTER TABLE tasks ADD COLUMN status TEXT DEFAULT 'todo';

UPDATE tasks SET status = 'complete' WHERE completed_at IS NOT NULL;

ALTER TABLE tasks DROP COLUMN completed_at;