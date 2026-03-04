-- Índices, Row-Level Security e Triggers (Feature 003 - Fase 6)

-- Índices
CREATE INDEX idx_tenants_slug_status ON wisa_crm_db.tenants (slug, status);
CREATE INDEX idx_products_slug_status ON wisa_crm_db.products (slug, status);
CREATE INDEX idx_subscriptions_tenant ON wisa_crm_db.subscriptions (tenant_id, status);
CREATE INDEX idx_subscriptions_product ON wisa_crm_db.subscriptions (product_id, status);
CREATE INDEX idx_subscriptions_dates ON wisa_crm_db.subscriptions (tenant_id, start_date, end_date);
CREATE INDEX idx_payments_subscription ON wisa_crm_db.payments (subscription_id);
CREATE INDEX idx_payments_status_date ON wisa_crm_db.payments (status, payment_date);
CREATE INDEX idx_users_tenant_email ON wisa_crm_db.users (tenant_id, email);
CREATE INDEX idx_users_tenant_status ON wisa_crm_db.users (tenant_id, status);
CREATE INDEX idx_user_product_access_user ON wisa_crm_db.user_product_access (user_id);
CREATE INDEX idx_user_product_access_tenant ON wisa_crm_db.user_product_access (tenant_id, product_id);
CREATE INDEX idx_refresh_tokens_hash ON wisa_crm_db.refresh_tokens (token_hash) WHERE revoked_at IS NULL;
CREATE INDEX idx_audit_tenant_date ON wisa_crm_db.audit_logs (tenant_id, created_at DESC);

-- Função para updated_at
CREATE OR REPLACE FUNCTION wisa_crm_db.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers updated_at
CREATE TRIGGER tr_tenants_updated_at
    BEFORE UPDATE ON wisa_crm_db.tenants
    FOR EACH ROW EXECUTE FUNCTION wisa_crm_db.set_updated_at();
CREATE TRIGGER tr_products_updated_at
    BEFORE UPDATE ON wisa_crm_db.products
    FOR EACH ROW EXECUTE FUNCTION wisa_crm_db.set_updated_at();
CREATE TRIGGER tr_subscriptions_updated_at
    BEFORE UPDATE ON wisa_crm_db.subscriptions
    FOR EACH ROW EXECUTE FUNCTION wisa_crm_db.set_updated_at();
CREATE TRIGGER tr_payments_updated_at
    BEFORE UPDATE ON wisa_crm_db.payments
    FOR EACH ROW EXECUTE FUNCTION wisa_crm_db.set_updated_at();
CREATE TRIGGER tr_users_updated_at
    BEFORE UPDATE ON wisa_crm_db.users
    FOR EACH ROW EXECUTE FUNCTION wisa_crm_db.set_updated_at();
CREATE TRIGGER tr_user_product_access_updated_at
    BEFORE UPDATE ON wisa_crm_db.user_product_access
    FOR EACH ROW EXECUTE FUNCTION wisa_crm_db.set_updated_at();
CREATE TRIGGER tr_refresh_tokens_updated_at
    BEFORE UPDATE ON wisa_crm_db.refresh_tokens
    FOR EACH ROW EXECUTE FUNCTION wisa_crm_db.set_updated_at();
CREATE TRIGGER tr_audit_logs_updated_at
    BEFORE UPDATE ON wisa_crm_db.audit_logs
    FOR EACH ROW EXECUTE FUNCTION wisa_crm_db.set_updated_at();

-- Row-Level Security
ALTER TABLE wisa_crm_db.subscriptions ENABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.user_product_access ENABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE wisa_crm_db.audit_logs ENABLE ROW LEVEL SECURITY;

-- Policies RLS (tabelas com tenant_id direto)
CREATE POLICY tenant_isolation ON wisa_crm_db.subscriptions
    FOR ALL USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON wisa_crm_db.users
    FOR ALL USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON wisa_crm_db.user_product_access
    FOR ALL USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON wisa_crm_db.refresh_tokens
    FOR ALL USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

CREATE POLICY tenant_isolation ON wisa_crm_db.audit_logs
    FOR ALL USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid)
    WITH CHECK (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Policy RLS para payments (sem tenant_id direto - via subquery)
CREATE POLICY tenant_isolation ON wisa_crm_db.payments
    FOR ALL
    USING (
        subscription_id IN (
            SELECT id FROM wisa_crm_db.subscriptions
            WHERE tenant_id = current_setting('app.current_tenant_id', true)::uuid
        )
    )
    WITH CHECK (
        subscription_id IN (
            SELECT id FROM wisa_crm_db.subscriptions
            WHERE tenant_id = current_setting('app.current_tenant_id', true)::uuid
        )
    );
