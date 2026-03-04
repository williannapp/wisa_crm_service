-- Schema wisa_crm_db e tipos ENUM (Feature 003 - Fase 1)
CREATE SCHEMA IF NOT EXISTS wisa_crm_db;

CREATE TYPE wisa_crm_db.tenant_type AS ENUM ('person', 'company');
CREATE TYPE wisa_crm_db.tenant_status AS ENUM ('active', 'inactive', 'blocked');
CREATE TYPE wisa_crm_db.product_status AS ENUM ('active', 'inactive', 'blocked');
CREATE TYPE wisa_crm_db.subscription_status AS ENUM ('pending', 'active', 'suspended', 'canceled');
CREATE TYPE wisa_crm_db.subscription_type AS ENUM ('free', 'payment');
CREATE TYPE wisa_crm_db.payment_status AS ENUM ('pending', 'paid', 'failed', 'refunded');
CREATE TYPE wisa_crm_db.user_status AS ENUM ('active', 'blocked');
CREATE TYPE wisa_crm_db.access_profile AS ENUM ('admin', 'operator', 'view');
