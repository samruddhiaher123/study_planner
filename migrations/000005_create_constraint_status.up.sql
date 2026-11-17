ALTER TABLE tasks
ADD CONSTRAINT check_task_status
CHECK (
    status IN (
        'Not Started',
        'In Progress',
        'Completed',
        'On Hold'
    )
);

ALTER TABLE tasks
ADD CONSTRAINT check_task_priority
CHECK (
    priority IN ('High', 'Medium', 'Low')
);