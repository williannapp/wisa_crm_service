# Feature 004 — Package de Erro Padronizado — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Estrutura de diretórios e tipo estruturado AppError | 1/1 | Concluída |
| Fase 2 | Catálogo de códigos de erro e mapeamento domain→AppError | 1/1 | Concluída |
| Fase 3 | ErrorMapper na delivery e integração com handlers | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-estrutura-diretorios-tipo-erro.md](./fase-1-estrutura-diretorios-tipo-erro.md)
- [fase-2-catalogo-codigos-mapeador.md](./fase-2-catalogo-codigos-mapeador.md)
- [fase-3-integracao-delivery.md](./fase-3-integracao-delivery.md)

## Resumo das Tasks

- [x] Fase 1 — Criar estrutura `pkg/errors/` e tipo estruturado AppError (código, mensagem, detalhe, status HTTP)
- [x] Fase 2 — Definir catálogo de códigos padronizados e lógica de mapeamento domain→AppError
- [x] Fase 3 — Implementar ErrorMapper na camada delivery, garantir conformidade Clean Architecture e evitar vazamento de informações

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | Fase 2 |

## Ordem Sugerida de Implementação

1. Fase 1 — Estrutura e tipo base
2. Fase 2 — Códigos e mapeamento domain
3. Fase 3 — Integração na delivery

## Referências

- [docs/context.md](../../context.md)
- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../adrs/ADR-005-clean-architecture.md)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
