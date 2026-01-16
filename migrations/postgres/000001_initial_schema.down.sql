-- LinkFlow Initial Schema Migration - Rollback
-- Version: 000001
-- Description: Drop all tables and extensions

-- Drop triggers first
DROP TRIGGER IF EXISTS update_templates_updated_at ON templates;
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TRIGGER IF EXISTS update_usage_records_updated_at ON usage_records;
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS update_plans_updated_at ON plans;
DROP TRIGGER IF EXISTS update_webhook_endpoints_updated_at ON webhook_endpoints;
DROP TRIGGER IF EXISTS update_schedules_updated_at ON schedules;
DROP TRIGGER IF EXISTS update_credentials_updated_at ON credentials;
DROP TRIGGER IF EXISTS update_node_executions_updated_at ON node_executions;
DROP TRIGGER IF EXISTS update_executions_updated_at ON executions;
DROP TRIGGER IF EXISTS update_workflows_updated_at ON workflows;
DROP TRIGGER IF EXISTS update_workspace_members_updated_at ON workspace_members;
DROP TRIGGER IF EXISTS update_workspaces_updated_at ON workspaces;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables in reverse order of creation (respecting foreign key dependencies)
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS template_categories;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS usage_records;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS webhook_logs;
DROP TABLE IF EXISTS webhook_endpoints;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS credential_shares;
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS execution_logs;
DROP TABLE IF EXISTS node_executions;
DROP TABLE IF EXISTS executions;
DROP TABLE IF EXISTS workflow_versions;
DROP TABLE IF EXISTS workflows;
DROP TABLE IF EXISTS workspace_members;
DROP TABLE IF EXISTS workspaces;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS users;

-- Note: We don't drop the uuid-ossp extension as it might be used by other databases
