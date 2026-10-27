CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    title TEXT,
    subject TEXT,
    description TEXT,
    date TEXT,
    start_time TEXT,
    end_time TEXT,
    status TEXT NOT NULL DEFAULT 'Not Started',
    priority TEXT NOT NULL DEFAULT 'Medium',
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ,
    CONSTRAINT fk_tasks_user
        FOREIGN KEY (user_id) REFERENCES users (id)
);
