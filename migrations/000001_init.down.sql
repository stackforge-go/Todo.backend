-- ------------------------------------------------------------------
-- Tasks
-- ------------------------------------------------------------------
DROP INDEX IF EXISTS todo.ix_tasks_user_created;
DROP INDEX IF EXISTS todo.ix_tasks_user_active;
DROP INDEX IF EXISTS todo.ix_tasks_user_id;
DROP TABLE IF EXISTS todo.tasks;
-- ------------------------------------------------------------------
-- Users
-- ------------------------------------------------------------------
DROP INDEX IF EXISTS todo.ux_users_email_lower;
DROP TABLE IF EXISTS todo.users;
-- ------------------------------------------------------------------
-- Schema
-- ------------------------------------------------------------------
DROP SCHEMA IF EXISTS todo;
