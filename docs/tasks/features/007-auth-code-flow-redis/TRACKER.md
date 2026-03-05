# Feature 007 — Authorization Code Flow com Redis — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Infraestrutura Redis (Docker, variáveis de ambiente, cliente Go) | 1/1 | Concluída |
| Fase 2 | Armazenamento de Authorization Code no Redis (interface, implementação, TTL 40s) | 1/1 | Concluída |
| Fase 3 | Alteração do endpoint de login (gerar code, armazenar no Redis, responder HTTP 302) | 1/1 | Concluída |
| Fase 4 | Endpoint POST /api/v1/auth/token (validar code, remover do Redis, retornar JWT) | 1/1 | Concluída |
| Fase 5 | Documentação de integração para o cliente (GET /callback, troca code por token) | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-redis-infraestrutura.md](./fase-1-redis-infraestrutura.md)
- [fase-2-auth-code-storage-redis.md](./fase-2-auth-code-storage-redis.md)
- [fase-3-alteracao-login-redirect-302.md](./fase-3-alteracao-login-redirect-302.md)
- [fase-4-endpoint-auth-token.md](./fase-4-endpoint-auth-token.md)
- [fase-5-documentacao-integracao-cliente.md](./fase-5-documentacao-integracao-cliente.md)

## Resumo das Tasks

- [x] Fase 1 — Adicionar Redis ao Docker Compose, variáveis de ambiente (REDIS_URL), cliente Redis em Go
- [x] Fase 2 — Interface AuthCodeStore, implementação Redis com TTL 40s, estrutura de dados para claims do JWT
- [x] Fase 3 — Modificar Use Case e handler de login: gerar code aleatório, armazenar no Redis, responder HTTP 302 para redirect_url
- [x] Fase 4 — Novo Use Case ExchangeCodeForToken, handler POST /api/v1/auth/token, resposta { access_token, expires_in }
- [x] Fase 5 — Documentação para aplicações clientes: GET /callback, troca de code por token, validação de state

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | Fase 2 |
| Fase 4 | Fase 2 |
| Fase 5 | Fase 3, Fase 4 |

## Ordem Sugerida de Implementação

1. Fase 1 — Infraestrutura Redis
2. Fase 2 — Auth Code Store
3. Fase 3 — Alteração do login (paralelizável parcialmente com Fase 4)
4. Fase 4 — Endpoint /auth/token
5. Fase 5 — Documentação de integração

## Notas Importantes

- **TTL do code:** 40 segundos conforme especificação (vs. 120s na ATA) — janela curta reduz risco de interceptação.
- **Resposta do login:** HTTP 302 com Location para a redirect_url do cliente. Alternativa: JSON com redirect_url para SPAs que usam fetch (frontend faz window.location).
- **Code single-use:** Removido do Redis imediatamente após troca por token no endpoint /auth/token.
- **redirect_url:** Montada internamente pelo backend a partir de tenant_slug, product_slug e state — evita Open Redirect (ADR-010).
- **Aplicação cliente:** O wisa-crm-service é o IdP; as aplicações clientes (ex: gestao-pocket) devem implementar GET /callback e trocar code por token.

## Referências

- [docs/context.md](../../context.md)
- [docs/integration/auth-code-flow-integration.md](../../integration/auth-code-flow-integration.md) — Guia de integração para clientes
- [ATA-2025-03-05 — Redirect Seguro e Passagem de JWT](../../DON'T%20READ/ideas-new-features/ATA-2025-03-05-redirect-seguro-e-passagem-jwt.md)
- [ADR-005 — Clean Architecture](../../adrs/ADR-005-clean-architecture.md)
- [ADR-006 — JWT com Assinatura Assimétrica](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [ADR-010 — Fluxo Centralizado de Autenticação](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
