# Feature 003 — Estrutura de Tabelas do Banco de Dados — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Schema `wisa_crm_db` e tipos ENUM | 1/1 | Concluída |
| Fase 2 | Tabelas `tenants` e `products` | 1/1 | Concluída |
| Fase 3 | Tabelas `subscriptions` e `payments` | 1/1 | Concluída |
| Fase 4 | Tabelas `users` e `user_product_access` | 1/1 | Concluída |
| Fase 5 | Tabelas `refresh_tokens` e `audit_logs` | 1/1 | Concluída |
| Fase 6 | Índices, Row-Level Security e triggers auxiliares | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-schema-e-enums.md](./fase-1-schema-e-enums.md)
- [fase-2-tabelas-tenants-products.md](./fase-2-tabelas-tenants-products.md)
- [fase-3-tabelas-subscriptions-payments.md](./fase-3-tabelas-subscriptions-payments.md)
- [fase-4-tabelas-users-user-product-access.md](./fase-4-tabelas-users-user-product-access.md)
- [fase-5-tabelas-refresh-tokens-audit-logs.md](./fase-5-tabelas-refresh-tokens-audit-logs.md)
- [fase-6-indices-rls-triggers.md](./fase-6-indices-rls-triggers.md)

## Resumo das Tasks

- [x] Fase 1 — Criar schema `wisa_crm_db` no banco e definir todos os tipos ENUM (tenant_type, tenant_status, product_status, subscription_status, subscription_type, payment_status, user_status, access_profile)
- [x] Fase 2 — Implementar tabelas `tenants` e `products` com constraints e comentários
- [x] Fase 3 — Implementar tabelas `subscriptions` e `payments` com relacionamentos e constraints
- [x] Fase 4 — Implementar tabelas `users` e `user_product_access` com relacionamentos e unicidade
- [x] Fase 5 — Implementar tabelas `refresh_tokens` e `audit_logs` (com particionamento em audit_logs)
- [x] Fase 6 — Criar índices de performance, habilitar Row-Level Security com políticas, triggers de `updated_at`

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | Fase 2 |
| Fase 4 | Fase 2 |
| Fase 5 | Fase 4 |
| Fase 6 | Fase 2, 3, 4, 5 |

## Ordem Sugerida de Implementação

1. Fase 1 — Schema e ENUMs (base)
2. Fase 2 — Tenants e Products
3. Fase 3 — Subscriptions e Payments
4. Fase 4 — Users e User Product Access
5. Fase 5 — Refresh Tokens e Audit Logs
6. Fase 6 — Índices, RLS e Triggers

## Referências

- [docs/context.md](../../context.md)
- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../adrs/ADR-008-arquitetura-multi-tenant.md)
- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
