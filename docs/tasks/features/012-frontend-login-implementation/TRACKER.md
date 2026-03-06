# Feature 012 — Frontend Login Implementation — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Leitura e validação dos parâmetros de query (tenant_slug, product_slug, state) | 0/1 | Pendente |
| Fase 2 | Serviço de autenticação e configuração HTTP | 0/1 | Pendente |
| Fase 3 | Integração no formulário de login e redirect | 0/1 | Pendente |

## Arquivos de Tasks

- [fase-1-parametros-query-validacao.md](./fase-1-parametros-query-validacao.md)
- [fase-2-servico-auth-config-http.md](./fase-2-servico-auth-config-http.md)
- [fase-3-integracao-formulario-redirect.md](./fase-3-integracao-formulario-redirect.md)

## Resumo das Tasks

- [ ] Fase 1 — Receber tenant_slug, product_slug e state via query params; validar no carregamento; exibir mensagem de erro sem chamar backend
- [ ] Fase 2 — Criar AuthService, configurar HttpClient, definir interface de resposta do login
- [ ] Fase 3 — Integrar formulário com AuthService; POST ao backend; redirect via window.location.href; tratamento de erros

## Escopo

- **Objetivo:** Implementar a lógica de autenticação no frontend do wisa-crm-service, integrando com o endpoint POST /api/v1/auth/login e realizando o redirect conforme fluxo Authorization Code (ADR-010).
- **Dependência:** Feature 011 (Tela de Login) — design já implementado.
- **Backend:** Endpoint POST /api/v1/auth/login existente (Feature 005, 007). O frontend espera resposta JSON com `redirect_url`; ver nota em fase-2 sobre compatibilidade.

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | — |
| Fase 3 | Fase 1, Fase 2 |

## Ordem Sugerida de Implementação

1. Fase 1 — Parâmetros de query e validação
2. Fase 2 — Serviço de autenticação
3. Fase 3 — Integração e redirect

## Referências

- [docs/context.md](../../context.md)
- [docs/code_guidelines/frontend.md](../../code_guidelines/frontend.md)
- [docs/adrs/ADR-002-angular-como-framework-frontend.md](../../adrs/ADR-002-angular-como-framework-frontend.md)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [docs/integration/auth-code-flow-integration.md](../../integration/auth-code-flow-integration.md)
