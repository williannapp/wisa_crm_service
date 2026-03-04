-- Rollback: Índices, Row-Level Security e Triggers

-- Drop policies
DROP POLICY IF EXISTS tenant_isolation ON wisa_crm_db.payments;
DROP POLICY IF EXISTS tenant_isolation ON wisa_crm_db.audit_logs;
DROP POLICY IF EXISTS tenant_isolation ON wisa_crm_db.refresh_tokens;
DROP POLICY IF EXISTS tenant_isolation ON wisa_crm_db.user_product_access;
DROP POLICY IF EXISTS tenant_isolation ON wisa_crm_db.users;
DROP POLICY IF EXISTS tenant_isolation ON wisa_crm_db.subscriptions;

-- Disable RLS
ALTER TABLE wisa_crm_db.audit_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.refresh_tokens DISABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.user_product_access DISABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.users DISABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.payments DISABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.subscriptions DISABLE ROW LEVEL SECURITY;

-- Drop triggers
DROP TRIGGER IF EXISTS tr_audit_logs_updated_at ON wisa_crm_db.audit_logs;
DROP TRIGGER IF EXISTS tr_refresh_tokens_updated_at ON wisa_crm_db.refresh_tokens;
DROP TRIGGER IF EXISTS tr_user_product_access_updated_at ON wisa_crm_db.user_product_access;
DROP TRIGGER IF EXISTS tr_users_updated_at ON wisa_crm_db.users;
DROP TRIGGER IF EXISTS tr_payments_updated_at ON wisa_crm_db.payments;
DROP TRIGGER IF EXISTS tr_subscriptions_updated_at ON wisa_crm_db.subscriptions;
DROP TRIGGER IF EXISTS tr_products_updated_at ON wisa_crm_db.products;
DROP TRIGGER IF EXISTS tr_tenants_updated_at ON wisa_crm_db.tenants;

-- Drop function
DROP FUNCTION IF EXISTS wisa_crm_db.set_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS wisa_crm_db.idx_audit_tenant_date;
DROP INDEX IF EXISTS wisa_crm_db.idx_refresh_tokens_hash;
DROP INDEX IF EXISTS wisa_crm_db.idx_user_product_access_tenant;
DROP INDEX IF EXISTS wisa_crm_db.idx_user_product_access_user;
DROP INDEX IF EXISTS wisa_crm_db.idx_users_tenant_status;
DROP INDEX IF EXISTS wisa_crm_db.idx_users_tenant_email;
DROP INDEX IF EXISTS wisa_crm_db.idx_payments_status_date;
DROP INDEX IF EXISTS wisa_crm_db.idx_payments_subscription;
DROP INDEX IF EXISTS wisa_crm_db.idx_subscriptions_dates;
DROP INDEX IF EXISTS wisa_crm_db.idx_subscriptions_product;
DROP INDEX IF EXISTS wisa_crm_db.idx_subscriptions_tenant;
DROP INDEX IF EXISTS wisa_crm_db.idx_products_slug_status;
DROP INDEX IF EXISTS wisa_crm_db.idx_tenants_slug_status;
