CREATE TABLE tasks (
    task_id TEXT PRIMARY KEY,
    container_id TEXT,
    date_time TEXT,
    args JSONB,
    status TEXT,
    success_job_ids TEXT[],
    failed_job_ids TEXT[],
    completion_rate FLOAT8,
    is_retry BOOLEAN,
    parent_task_id TEXT
);