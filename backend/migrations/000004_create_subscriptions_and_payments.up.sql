-- Tabelas subscriptions e payments (Feature 003 - Fase 3)
CREATE TABLE wisa_crm_db.subscriptions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES wisa_crm_db.tenants(id),
    product_id  UUID NOT NULL REFERENCES wisa_crm_db.products(id),
    type        wisa_crm_db.subscription_type NOT NULL DEFAULT 'payment',
    status      wisa_crm_db.subscription_status NOT NULL DEFAULT 'pending',
    start_date  DATE NOT NULL,
    end_date    DATE NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_subscription_dates CHECK (end_date >= start_date)
);

CREATE TABLE wisa_crm_db.payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id  UUID NOT NULL REFERENCES wisa_crm_db.subscriptions(id),
    amount          DECIMAL(12, 2) NOT NULL,
    payment_date    DATE NOT NULL,
    status          wisa_crm_db.payment_status NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_amount_positive CHECK (amount >= 0)
);
