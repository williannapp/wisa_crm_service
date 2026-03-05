# Feature 008 — Refresh Token Endpoint — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Migration: Adicionar product_id à tabela refresh_tokens | 1/1 | Concluída |
| Fase 2 | Refresh Token no fluxo de troca de code (POST /auth/token) | 1/1 | Concluída |
| Fase 3 | Endpoint POST /api/v1/auth/refresh | 1/1 | Concluída |
| Fase 4 | Documentação de integração para clientes | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-migration-product-id-refresh-tokens.md](./fase-1-migration-product-id-refresh-tokens.md)
- [fase-2-refresh-token-no-token-exchange.md](./fase-2-refresh-token-no-token-exchange.md)
- [fase-3-endpoint-post-auth-refresh.md](./fase-3-endpoint-post-auth-refresh.md)
- [fase-4-documentacao-integracao-refresh.md](./fase-4-documentacao-integracao-refresh.md)

## Resumo das Tasks

- [x] Fase 1 — Migration 000008: adicionar product_id e índice composto em refresh_tokens
- [x] Fase 2 — Estender AuthCodeData com ProductID; gerar e persistir refresh token no POST /auth/token; retornar refresh_token e refresh_expires_in
- [x] Fase 3 — Implementar endpoint POST /api/v1/auth/refresh; validação por hash + tenant_slug + product_slug; rotação atômica; verificação de assinatura
- [x] Fase 4 — Documentar fluxo de refresh para aplicações clientes em docs/integration

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | Fase 2 |
| Fase 4 | Fase 2, Fase 3 |

## Ordem Sugerida de Implementação

1. Fase 1 — Migration product_id
2. Fase 2 — Refresh token no token exchange
3. Fase 3 — Endpoint POST /auth/refresh
4. Fase 4 — Documentação de integração

## Notas Importantes

- **Tabela refresh_tokens:** Requer product_id para escopar o token por (user, tenant, product). Migration 000008.
- **Validação no refresh:** Hash SHA-256 do token + tenant_id + product_id (resolvidos de tenant_slug e product_slug). Não revelar motivo específico em 401.
- **Rotação:** Cada uso invalida o anterior (revoked_at = NOW()). Novo par (access + refresh) emitido.
- **Access Token:** Não armazenado em banco (15 min).
- **Refresh Token:** SHA-256 hash, 7 dias, rotativo.

## Referências

- [docs/context.md](../../context.md)
- [ADR-006 — JWT com Assinatura Assimétrica](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [ADR-010 — Fluxo Centralizado de Autenticação](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/integration/auth-code-flow-integration.md](../../integration/auth-code-flow-integration.md)
