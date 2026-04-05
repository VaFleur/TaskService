CREATE TABLE tasks (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       title TEXT NOT NULL,
                       description TEXT,
                       status VARCHAR(20) DEFAULT 'pending',
                       created_at TIMESTAMPTZ DEFAULT NOW(),
                       updated_at TIMESTAMPTZ DEFAULT NOW(),
                       deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_tasks_status_deleted ON tasks(status) WHERE deleted_at IS NULL;