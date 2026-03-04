-- Rollback: Schema wisa_crm_db e tipos ENUM
DROP TYPE IF EXISTS wisa_crm_db.access_profile;
DROP TYPE IF EXISTS wisa_crm_db.user_status;
DROP TYPE IF EXISTS wisa_crm_db.payment_status;
DROP TYPE IF EXISTS wisa_crm_db.subscription_type;
DROP TYPE IF EXISTS wisa_crm_db.subscription_status;
DROP TYPE IF EXISTS wisa_crm_db.product_status;
DROP TYPE IF EXISTS wisa_crm_db.tenant_status;
DROP TYPE IF EXISTS wisa_crm_db.tenant_type;
DROP SCHEMA IF EXISTS wisa_crm_db;
