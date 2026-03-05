# Fase 1 — Migration: Adicionar product_id à tabela refresh_tokens

## Objetivo

Alterar a tabela `refresh_tokens` para incluir `product_id`, permitindo que o endpoint de refresh valide o token com base em `refresh_token`, `product_slug` e `tenant_slug`. O escopo de um refresh token passa a ser (user, tenant, product), conforme requisito da feature e ADR-008.

---

## Análise: Necessidade da Alteração

### Pensamento 1

O requisito especifica que o software de autenticação deve "validar se a hash existe no banco com base no refresh_token, product_slug e tenant_slug enviado pela aplicação cliente". A tabela atual possui apenas `user_id`, `tenant_id`, `token_hash`, `expires_at`, `revoked_at`.

**Conclusão:** Sim, é **necessário** alterar a tabela. O `product_slug` (e consequentemente `product_id`) não está presente. Um usuário pode ter acesso a múltiplos produtos no mesmo tenant; o refresh token deve ser escopado por produto para garantir que um token emitido para o produto A não seja usado para renovar sessão do produto B.

### Ação 1.1

Verificar estrutura atual da tabela `refresh_tokens`:

```sql
-- Estrutura atual (migration 000006)
CREATE TABLE wisa_crm_db.refresh_tokens (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES wisa_crm_db.users(id),
    tenant_id   UUID NOT NULL REFERENCES wisa_crm_db.tenants(id),
    token_hash  CHAR(64) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    ...
);
```

### Observação 1.1

A tabela possui `tenant_id` mas não `product_id`. O índice `idx_refresh_tokens_hash` é `(token_hash) WHERE revoked_at IS NULL`. A busca no refresh será por `token_hash` (único) + validação de `tenant_id` e `product_id`. O `token_hash` já é UNIQUE globalmente, então a query retorna no máximo um registro; a aplicação deve verificar se `tenant_id` e `product_id` correspondem aos slugs informados.

---

### Pensamento 2

A migration deve adicionar `product_id` como `NOT NULL` com FK para `products(id)`. Se houver registros existentes (cenário improvável, pois refresh tokens ainda não são emitidos), a migration falhará. Documentar que a tabela deve estar vazia ou que será necessário truncar antes da migration em ambientes de desenvolvimento que tenham dados de teste.

### Ação 1.2

Criar migration `0 `:

**Arquivo `.up.sql`:**
```sql
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
```

**Arquivo `.down.sql`:**
```sql
DROP INDEX IF EXISTS wisa_crm_db.idx_refresh_tokens_lookup;
ALTER TABLE wisa_crm_db.refresh_tokens DROP COLUMN IF EXISTS product_id;
```

### Observação 1.2

O índice composto otimiza a query `WHERE token_hash = ? AND tenant_id = ? AND product_id = ? AND revoked_at IS NULL`. O RLS da tabela `refresh_tokens` usa `tenant_id`; ao adicionar `product_id`, a policy existente (`tenant_isolation`) continua válida, pois a policy usa `tenant_id = current_setting(...)`. A aplicação fornecerá o `tenant_id` resolvido a partir de `tenant_slug` antes da query. Não é necessário alterar a policy de RLS para `product_id` neste estágio, pois o isolamento primário é por tenant.

---

### Pensamento 3

Verificar se a migration 000007 (índices, RLS, triggers) precisa ser atualizada. O índice `idx_refresh_tokens_hash` existente é `(token_hash) WHERE revoked_at IS NULL`. O novo índice `idx_refresh_tokens_lookup` inclui `tenant_id` e `product_id` para cobrir a query de refresh. Ambos podem coexistir: o primeiro serve buscas apenas por hash; o segundo otimiza a query completa. Avaliar se manter ambos ou substituir. Para o refresh, a query sempre inclui tenant_id e product_id (resolvidos dos slugs), então o índice composto é preferível. O índice antigo pode permanecer para outras consultas (ex.: revogação em logout por token_hash). Manter ambos.

### Ação 1.3

Não remover o índice `idx_refresh_tokens_hash` existente. O novo índice `idx_refresh_tokens_lookup` complementa para a operação de refresh com filtros adicionais.

### Observação 1.3

Conformidade com ADR-003 (índices para operações frequentes) e ADR-004 (GORM/migrations).

---

### Pensamento 4 (Segurança)

O `product_id` não introduce vetores de ataque adicionais. A validação no endpoint de refresh garantirá que o `tenant_slug` e `product_slug` informados correspondam ao registro encontrado, impedindo que um refresh token de um produto seja usado em outro.

### Observação 1.4

Alinhado com code guidelines (SQL parametrizado, sem concatenação).

---

## Checklist de Implementação

- [ ] 1. Criar migration `000008_add_product_id_to_refresh_tokens.up.sql`
- [ ] 2. Criar migration `000008_add_product_id_to_refresh_tokens.down.sql`
- [ ] 3. Executar migration em ambiente de desenvolvimento
- [ ] 4. Verificar que a tabela `refresh_tokens` está vazia antes da migration (ou documentar truncate se necessário)
- [ ] 5. Atualizar documentação do schema em docs/backend ou ADR-008 (nota sobre product_id)

---

## Referências

- ADR-003 — PostgreSQL como banco de dados
- ADR-008 — Arquitetura multi-tenant (modelo refresh_tokens)
- docs/code_guidelines/backend.md (migrations versionadas)
- backend/migrations/000006_create_refresh_tokens_and_audit_logs.up.sql
