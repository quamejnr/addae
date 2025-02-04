-- +goose Up
-- +goose StatementBegin
-- Create projects table
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT CHECK(length(name) <= 100) NOT NULL,
    summary TEXT CHECK(length(summary) <= 255),
    desc TEXT DEFAULT '',
    status TEXT CHECK(status IN ('todo', 'in progress', 'completed', 'archived')) NOT NULL DEFAULT 'todo',
    date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
    date_updated DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER,
    title TEXT CHECK(length(title) <= 100) NOT NULL,
    desc TEXT DEFAULT '',
    status TEXT CHECK(status IN ('todo', 'in progress', 'completed', 'archived')) NOT NULL DEFAULT 'todo',
    date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
    date_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create logs table
CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER,
    title TEXT CHECK(length(title) <= 100),
    desc TEXT DEFAULT '',
    date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
    date_updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create update triggers for each table
CREATE TRIGGER IF NOT EXISTS update_projects_date_updated 
    AFTER UPDATE ON projects
    FOR EACH ROW
    BEGIN
        UPDATE projects SET date_updated = CURRENT_TIMESTAMP
        WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_tasks_date_updated 
    AFTER UPDATE ON tasks
    FOR EACH ROW
    BEGIN
        UPDATE tasks SET date_updated = CURRENT_TIMESTAMP
        WHERE id = NEW.id;
    END;

CREATE TRIGGER IF NOT EXISTS update_logs_date_updated 
    AFTER UPDATE ON logs
    FOR EACH ROW
    BEGIN
        UPDATE logs SET date_updated = CURRENT_TIMESTAMP
        WHERE id = NEW.id;
    END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop triggers first
DROP TRIGGER IF EXISTS update_projects_date_updated;
DROP TRIGGER IF EXISTS update_tasks_date_updated;
DROP TRIGGER IF EXISTS update_logs_date_updated;

-- Drop tables in reverse order of creation (due to foreign key constraints)
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
-- +goose StatementEnd
