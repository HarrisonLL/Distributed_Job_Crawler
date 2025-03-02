-- Add new columns
ALTER TABLE tasks ADD COLUMN host_type TEXT;
ALTER TABLE tasks ADD COLUMN run_id TEXT;
ALTER TABLE tasks ADD COLUMN company TEXT;
ALTER TABLE tasks ADD COLUMN job_type TEXT;
ALTER TABLE tasks ADD COLUMN location TEXT;
ALTER TABLE tasks ADD COLUMN numbers_of_jobs INT;

-- Remove old columns
ALTER TABLE tasks DROP COLUMN container_id;
ALTER TABLE tasks DROP COLUMN args;
ALTER TABLE tasks DROP COLUMN failed_job_ids;
ALTER TABLE tasks DROP COLUMN is_retry_task;
ALTER TABLE tasks DROP COLUMN parent_task_id;
