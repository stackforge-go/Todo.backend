CREATE SCHEMA todo;
-- ------------------------------------------------------------------
-- Users
-- ------------------------------------------------------------------
CREATE TABLE todo.users (
    id UUID PRIMARY KEY,
    version BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    email TEXT NOT NULL,
    email_verified BOOLEAN NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(150)
);
-- email без учёта регистра
CREATE UNIQUE INDEX ux_users_email_lower ON todo.users (LOWER(email));
-- ------------------------------------------------------------------
-- Tasks
-- ------------------------------------------------------------------
CREATE TABLE todo.tasks (
    id UUID PRIMARY KEY,
    version BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    completed BOOLEAN NOT NULL,
    completed_at TIMESTAMPTZ,
    user_id UUID NOT NULL REFERENCES todo.users(id) ON DELETE CASCADE,
    CONSTRAINT chk_tasks_completed CHECK (
        (
            completed = FALSE
            AND completed_at IS NULL
        )
        OR (
            completed = TRUE
            AND completed_at IS NOT NULL
        )
    )
);
CREATE INDEX ix_tasks_user_id ON todo.tasks (user_id);
CREATE INDEX ix_tasks_user_active ON todo.tasks (user_id)
WHERE completed = FALSE;
CREATE INDEX ix_tasks_user_created ON todo.tasks (user_id, created_at DESC);
