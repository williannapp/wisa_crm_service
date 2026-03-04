-- Tabelas users e user_product_access (Feature 003 - Fase 4)
CREATE TABLE wisa_crm_db.users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES wisa_crm_db.tenants(id),
    name            VARCHAR(255) NOT NULL,
    email           VARCHAR(320) NOT NULL,
    password_hash   VARCHAR(72) NOT NULL,
    status          wisa_crm_db.user_status NOT NULL DEFAULT 'active',
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, email)
);

CREATE TABLE wisa_crm_db.user_product_access (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES wisa_crm_db.users(id),
    product_id      UUID NOT NULL REFERENCES wisa_crm_db.products(id),
    tenant_id       UUID NOT NULL REFERENCES wisa_crm_db.tenants(id),
    access_profile  wisa_crm_db.access_profile NOT NULL DEFAULT 'view',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, product_id)
);
