# Fase 3 — Tabelas subscriptions e payments

## Objetivo

Implementar as tabelas `subscriptions` e `payments` no schema `wisa-labs-db`, vinculando assinaturas a tenants e produtos, e pagamentos a assinaturas, conforme ADR-008.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A tabela `subscriptions` vincula um tenant a um produto com status, datas e tipo:
- `id` UUID PRIMARY KEY
- `tenant_id` UUID NOT NULL REFERENCES wisa_crm_db.tenants(id)
- `product_id` UUID NOT NULL REFERENCES wisa_crm_db.products(id)
- `type` subscription_type NOT NULL DEFAULT 'payment' — free ou payment
- `status` subscription_status NOT NULL DEFAULT 'pending'
- `start_date` DATE NOT NULL
- `end_date` DATE NOT NULL
- `created_at`, `updated_at` TIMESTAMPTZ
- CONSTRAINT chk_subscription_dates: end_date >= start_date

A tabela `payments` registra pagamentos vinculados a assinaturas:
- `id` UUID PRIMARY KEY
- `subscription_id` UUID NOT NULL REFERENCES wisa_crm_db.subscriptions(id)
- `amount` DECIMAL(12,2) NOT NULL
- `payment_date` DATE NOT NULL
- `status` payment_status NOT NULL DEFAULT 'pending'
- `created_at`, `updated_at` TIMESTAMPTZ
- CONSTRAINT chk_amount_positive: amount >= 0

### Ação 3.1

Migration 000004. Criar subscriptions antes de payments (payments depende de subscriptions).

### Observação 3.1

A ordem de criação segue o grafo de dependências: tenants, products (Fase 2) → subscriptions → payments.

---

### Pensamento 2

Política de ON DELETE para FKs:
- subscriptions.tenant_id: se um tenant for deletado, o que acontece com suas assinaturas? Em sistema SaaS, deleção de tenant é operação rara e geralmente soft-delete. Se hard-delete: RESTRICT ou CASCADE? ADR-008 não especifica. Por segurança e integridade: **RESTRICT** ou **ON DELETE RESTRICT** (padrão) — impede deletar tenant com assinaturas ativas. O mesmo para product_id.
- payments.subscription_id: RESTRICT — não permitir deletar assinatura com pagamentos registrados sem processo explícito.
- subscriptions: ON DELETE RESTRICT em tenant_id e product_id (padrão do PostgreSQL)

### Ação 3.2

Usar REFERENCES sem ON DELETE (RESTRICT é o padrão). Documentar que operações de deleção em cascata não são automáticas — a aplicação deve gerenciar a ordem (ex: soft-delete ou processo de encerramento de tenant).

### Observação 3.2

RESTRICT preserva integridade referencial. Adequado para domínio de billing.

---

### Pensamento 3

A tabela subscriptions possui `tenant_id` — portanto é multi-tenant e será protegida por RLS na Fase 6. A tabela payments não possui tenant_id diretamente; o tenant é inferido via subscription → tenant. Para RLS em payments, precisaríamos de uma policy que usa join com subscriptions, ou adicionar tenant_id desnormalizado em payments. O ADR-008 inclui tenant_id em user_product_access e refresh_tokens para RLS, mas **não** em payments. A política de RLS para payments na Fase 6 pode usar subquery: `subscription_id IN (SELECT id FROM wisa_crm_db.subscriptions WHERE tenant_id = current_setting(...))`. Ou podemos adicionar `tenant_id` em payments para simplificar RLS (desnormalização controlada). O ADR-008 não lista tenant_id em payments — seguir o ADR. A policy de RLS em payments será via join/subquery.

### Ação 3.3

Não adicionar tenant_id em payments. A policy de RLS na Fase 6 usará subquery/join com subscriptions para filtrar por tenant.

### Observação 3.3

Conformidade com ADR-008. Policies de RLS mais complexas são aceitáveis para tabelas com pouca frequência de acesso direto.

---

### Pensamento 4

Índices: o ADR-008 lista índices para subscriptions e payments:
- idx_subscriptions_tenant ON (tenant_id, status)
- idx_subscriptions_product ON (product_id, status)
- idx_subscriptions_dates ON (tenant_id, start_date, end_date)
- idx_payments_subscription ON (subscription_id)
- idx_payments_status_date ON (status, payment_date)

Esses índices serão criados na Fase 6. Esta fase apenas cria as tabelas.

### Ação 3.4

Não criar índices na Fase 3. Manter na Fase 6 para centralização.

### Observação 3.4

Tabelas funcionais sem índices podem impactar performance em ambiente de desenvolvimento; a Fase 6 deve ser implementada em sequência.

---

### Pensamento 5

Rollback: DROP TABLE payments primeiro (depende de subscriptions), depois subscriptions.

### Ação 3.5

Migration down: DROP TABLE IF EXISTS wisa_crm_db.payments; DROP TABLE IF EXISTS wisa_crm_db.subscriptions;

### Observação 3.5

Ordem correta para respeitar FKs no drop (PostgreSQL permite DROP CASCADE, mas preferir ordem explícita).

---

### Decisão final Fase 3

**Implementar:**
1. Migration `000004_create_subscriptions_and_payments.up.sql`:
   - CREATE TABLE subscriptions com FKs para tenants e products
   - CREATE TABLE payments com FK para subscriptions
2. Migration `000004_create_subscriptions_and_payments.down.sql`:
   - DROP TABLE payments, subscriptions

---

### Checklist de Implementação

1. [ ] Criar 000004_create_subscriptions_and_payments.up.sql
2. [ ] Incluir CONSTRAINT chk_subscription_dates e chk_amount_positive
3. [ ] Garantir REFERENCES com tabelas qualificadas wisa_crm_db.tenants, etc.
4. [ ] Criar 000004 down
5. [ ] Testar migrate up/down

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-008 subscriptions | tenant_id, product_id, type, status, start_date, end_date, CHECK dates |
| ADR-008 payments | subscription_id, amount, payment_date, status, CHECK amount |
| Integridade referencial | FKs com RESTRICT implícito |

---

## Referências

- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../../adrs/ADR-008-arquitetura-multi-tenant.md)
