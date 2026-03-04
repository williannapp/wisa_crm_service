# Feature 002 — Configuração do Banco de Dados — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Estrutura base do banco de dados (conexão GORM, sem queries/tabelas) | 1/1 | Concluída |
| Fase 2 | Variáveis de ambiente para conexão com o banco | 1/1 | Concluída |
| Fase 3 | Containers (Dockerfile PostgreSQL + docker-compose.yml) | 1/1 | Concluída |
| Fase 4 | Documentação (como rodar backend + docs/backend/vps-configurations.md) | 1/1 | Concluída |
| Fase 5 | ORM (estrutura base, migrations, rollback) | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-estrutura-base-banco.md](./fase-1-estrutura-base-banco.md)
- [fase-2-variaveis-ambiente.md](./fase-2-variaveis-ambiente.md)
- [fase-3-containers.md](./fase-3-containers.md)
- [fase-4-documentacao.md](./fase-4-documentacao.md)
- [fase-5-orm-migrations-rollback.md](./fase-5-orm-migrations-rollback.md)

## Resumo das Tasks

- [x] Fase 1 — Adicionar estrutura base do banco: pastas, conexão GORM, pool de conexões; sem queries ou tabelas
- [x] Fase 2 — Criar variáveis de ambiente necessárias para conexão (DATABASE_URL em .env.example)
- [x] Fase 3 — Criar Dockerfile para PostgreSQL e docker-compose.yml na raiz do projeto
- [x] Fase 4 — Documentação em docs/backend/ explicando como rodar o backend; docs/backend/vps-configurations.md para VPS
- [x] Fase 5 — Adaptações ORM, estrutura base, migrations versionadas (golang-migrate) e rollback

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | — (pode ser paralela à Fase 1) |
| Fase 3 | — |
| Fase 4 | Fase 3 (para documentar docker-compose); Fase 5 (para documentar migrations) |
| Fase 5 | Fase 1 (conexão); Fase 2 (DATABASE_URL) |

## Ordem Sugerida de Implementação

1. Fase 1 — Estrutura base
2. Fase 2 — Variáveis de ambiente
3. Fase 3 — Containers
4. Fase 5 — ORM e migrations
5. Fase 4 — Documentação (com informações completas de todas as fases anteriores)
