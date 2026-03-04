-- Rollback: Tabelas refresh_tokens e audit_logs
DROP TABLE IF EXISTS wisa_crm_db.audit_logs;
DROP TABLE IF EXISTS wisa_crm_db.refresh_tokens;
