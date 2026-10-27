CREATE TABLE study_completions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    task_id BIGINT,
    completed_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_study_completions_user
        FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE INDEX idx_study_completions_user_id
    ON study_completions (user_id);
CREATE INDEX idx_study_completions_completed_at
    ON study_completions (completed_at);
