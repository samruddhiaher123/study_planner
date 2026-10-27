CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200),
    email VARCHAR(320),
    password VARCHAR(100),
    created_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uni_users_email ON users (email);
