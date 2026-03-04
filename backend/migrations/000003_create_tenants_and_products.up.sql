-- Tabelas tenants e products (Feature 003 - Fase 2)
CREATE TABLE wisa_crm_db.tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(63) UNIQUE NOT NULL,
    name        VARCHAR(255) NOT NULL,
    tax_id      VARCHAR(18) NOT NULL,
    type        wisa_crm_db.tenant_type NOT NULL,
    status      wisa_crm_db.tenant_status NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_tax_id_length CHECK (
        (type = 'company' AND LENGTH(tax_id) = 14) OR
        (type = 'person' AND LENGTH(tax_id) = 11)
    )
);

CREATE TABLE wisa_crm_db.products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(63) UNIQUE NOT NULL,
    name        VARCHAR(255) NOT NULL,
    status      wisa_crm_db.product_status NOT NULL DEFAULT 'active',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
