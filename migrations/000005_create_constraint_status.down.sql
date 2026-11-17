ALTER TABLE tasks
DROP CONSTRAINT IF EXISTS check_task_status;

ALTER TABLE tasks
DROP CONSTRAINT IF EXISTS check_task_priority;