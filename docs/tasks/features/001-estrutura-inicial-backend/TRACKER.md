# Feature 001 — Estrutura Inicial do Backend — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Estrutura de diretórios do backend | 1/1 | Concluída |
| Fase 2 | Importar bibliotecas necessárias | 1/1 | Concluída |
| Fase 3 | Adicionar .gitignore | 1/1 | Concluída |
| Fase 4 | Configuração de variáveis de ambiente | 1/1 | Concluída |
| Fase 5 | Criar Dockerfile | 1/1 | Concluída |
| Fase 6 | Endpoint health | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-estrutura-diretorios.md](./fase-1-estrutura-diretorios.md)
- [fase-2-importar-bibliotecas.md](./fase-2-importar-bibliotecas.md)
- [fase-3-gitignore.md](./fase-3-gitignore.md)
- [fase-4-variaveis-ambiente.md](./fase-4-variaveis-ambiente.md)
- [fase-5-dockerfile.md](./fase-5-dockerfile.md)
- [fase-6-endpoint-health.md](./fase-6-endpoint-health.md)

## Resumo das Tasks

- [x] Fase 1 — Criar estrutura de diretórios do backend em `backend/` conforme Clean Architecture e guidelines
- [x] Fase 2 — Inicializar go.mod e importar bibliotecas (Gin como framework HTTP padrão, godotenv)
- [x] Fase 3 — Adicionar .gitignore na raiz e em backend/ para evitar commits desnecessários
- [x] Fase 4 — Configurar suporte a variáveis de ambiente (.env e .env.example)
- [x] Fase 5 — Criar Dockerfile multi-stage para build e execução do backend
- [x] Fase 6 — Implementar endpoint GET /health para validação de disponibilidade da aplicação
