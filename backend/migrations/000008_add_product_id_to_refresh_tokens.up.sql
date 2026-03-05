-- Adiciona product_id à refresh_tokens (Feature 008 - Fase 1)
-- Refresh tokens são escopados por (user, tenant, product).
-- PRÉ-REQUISITO: Tabela refresh_tokens deve estar vazia (refresh ainda não implementado).
ALTER TABLE wisa_crm_db.refresh_tokens
    ADD COLUMN product_id UUID NOT NULL REFERENCES wisa_crm_db.products(id);

-- Índice composto para busca eficiente na validação do refresh.
-- A busca usa token_hash (único) + tenant_id + product_id para validação.
CREATE INDEX idx_refresh_tokens_lookup
    ON wisa_crm_db.refresh_tokens (token_hash, tenant_id, product_id)
    WHERE revoked_at IS NULL;
