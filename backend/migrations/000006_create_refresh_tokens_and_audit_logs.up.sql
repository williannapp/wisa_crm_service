-- Tabelas refresh_tokens e audit_logs (Feature 003 - Fase 5)
CREATE TABLE wisa_crm_db.refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES wisa_crm_db.users(id),
    tenant_id   UUID NOT NULL REFERENCES wisa_crm_db.tenants(id),
    token_hash  CHAR(64) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address  INET,
    user_agent  TEXT
);

CREATE TABLE wisa_crm_db.audit_logs (
    id          UUID NOT NULL DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES wisa_crm_db.tenants(id),
    user_id     UUID REFERENCES wisa_crm_db.users(id),
    event_type  VARCHAR(50) NOT NULL,
    ip_address  INET,
    user_agent  TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE wisa_crm_db.audit_logs_default PARTITION OF wisa_crm_db.audit_logs
FOR VALUES FROM ('1970-01-01') TO ('2030-01-01');
