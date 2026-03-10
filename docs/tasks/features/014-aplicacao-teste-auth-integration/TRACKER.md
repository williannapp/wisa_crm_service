# Feature 014 — Aplicação de Teste com Integração Auth Code Flow — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Estrutura base: pasta test-app, backend Go, frontend Angular, tela Hello World | 1/1 | Concluída |
| Fase 2 | Redirect + state, callback, token exchange e armazenamento em cookie | 1/1 | Concluída |
| Fase 3 | Validação JWT no backend, JWKS, rota protegida /api/hello | 1/1 | Concluída |
| Fase 4 | Frontend integrado e interceptor com refresh token | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-estrutura-hello-world.md](./fase-1-estrutura-hello-world.md)
- [fase-2-redirect-callback-token.md](./fase-2-redirect-callback-token.md)
- [fase-3-validacao-jwt-rota-protegida.md](./fase-3-validacao-jwt-rota-protegida.md)
- [fase-4-frontend-integrado-refresh.md](./fase-4-frontend-integrado-refresh.md)

## Resumo das Tasks

- [x] Fase 1 — Criar pasta test-app/, backend Go (estrutura mínima Clean Architecture), frontend Angular (tela Hello World), sem alterar backend/ ou frontend/
- [x] Fase 2 — Implementar redirect com state, callback GET /callback, troca code por token, armazenar em cookie HttpOnly
- [x] Fase 3 — Obter JWKS, validar JWT (RS256, iss, aud, exp, nbf, kid), rota protegida /api/hello
- [x] Fase 4 — Frontend consumindo /api/hello, interceptor HTTP para refresh em 401

## Escopo

- **Objetivo:** Criar uma aplicação de teste (frontend Angular + backend Go) em pasta separada para validar a integração completa com o fluxo de Authorization Code do wisa-crm-service.
- **Restrições:** NÃO alterar nenhum código em `backend/` ou `frontend/`.
- **AUTH_SERVER_URL:** `https://auth.wisa.labs.com.br`
- **Conformidade:** docs/code_guidelines, docs/adrs, docs/integration/auth-code-flow-integration.md

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | Fase 2 |
| Fase 4 | Fase 3 |

## Ordem Sugerida de Implementação

1. Fase 1 — Estrutura base e Hello World
2. Fase 2 — Fluxo de autenticação (redirect, callback, token exchange)
3. Fase 3 — Validação JWT e rota protegida
4. Fase 4 — Integração frontend e refresh

## Referências

- [docs/context.md](../../context.md)
- [docs/integration/auth-code-flow-integration.md](../../integration/auth-code-flow-integration.md)
- [docs/code_guidelines/](../../code_guidelines/)
- [docs/adrs/](../../adrs/)
