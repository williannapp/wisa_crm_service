# Feature 013 — Revisão e Aprimoramento da Documentação Auth Code Flow — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Revisar e documentar processo de Autenticação (redirect, login, callback, token) | 1/1 | Concluída |
| Fase 2 | Documentar processo de Refresh com parâmetros e retornos | 1/1 | Concluída |
| Fase 3 | Processos adicionais, exemplos de código e tabela de erros | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-processo-autenticacao.md](./fase-1-processo-autenticacao.md)
- [fase-2-processo-refresh.md](./fase-2-processo-refresh.md)
- [fase-3-processos-adicionais-exemplos.md](./fase-3-processos-adicionais-exemplos.md)

## Resumo das Tasks

- [x] Fase 1 — Revisar doc; descrever passo a passo: redirect para login, POST /auth/login, callback, POST /auth/token; parâmetros de entrada e retornos (códigos HTTP, respostas)
- [x] Fase 2 — Documentar fluxo de refresh: POST /auth/refresh com parâmetros, retornos e códigos de erro
- [x] Fase 3 — Processos adicionais (JWKS, validação JWT); exemplos de código; tabela completa de erros; correção formato JSON de erro (error_code)

## Escopo

- **Objetivo:** Revisar e aprimorar `docs/integration/auth-code-flow-integration.md` para que seja o **guia definitivo** para desenvolvedores implementarem o sistema de login em suas aplicações.
- **Abordagem:** Somente planejamento e documentação — **nenhuma implementação de código** na estrutura do projeto.
- **Conformidade:** Atender code_guidelines e ADRs; garantir consistência com o backend real (endpoints, DTOs, erros).

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | Fase 1, Fase 2 |

## Ordem Sugerida de Implementação

1. Fase 1 — Processo de Autenticação completo
2. Fase 2 — Processo de Refresh
3. Fase 3 — Processos adicionais, exemplos e erros

## Referências

- [docs/context.md](../../context.md)
- [docs/integration/auth-code-flow-integration.md](../../integration/auth-code-flow-integration.md)
- [docs/code_guidelines/](../../code_guidelines/)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [docs/adrs/ADR-006-jwt-com-assinatura-assimetrica.md](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
