-- Rollback Migration 000006
DROP POLICY IF EXISTS automation_logs_org_isolation ON automation_execution_logs;
DROP POLICY IF EXISTS automation_rules_org_isolation ON automation_rules;
DROP TABLE IF EXISTS automation_execution_logs;
DROP TABLE IF EXISTS automation_rules;
