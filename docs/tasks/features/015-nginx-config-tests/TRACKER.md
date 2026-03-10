# Feature 015 — Configuração NGINX para Execução de Testes — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Estrutura da pasta nginx e arquivo de configuração NGINX com roteamento revisado | 1/1 | Concluída |
| Fase 2 | Docker Compose na pasta nginx para executar NGINX isoladamente | 1/1 | Concluída |
| Fase 3 | Integração ao docker-compose principal; alterar APP_URL e FRONTEND_URL na test-app | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-config-nginx-roteamento.md](./fase-1-config-nginx-roteamento.md)
- [fase-2-docker-compose-nginx.md](./fase-2-docker-compose-nginx.md)
- [fase-3-integracao-docker-compose-principal.md](./fase-3-integracao-docker-compose-principal.md)

## Resumo das Tasks

- [x] Fase 1 — Criar pasta `nginx/`, definir estrutura de arquivos de config e revisar/corrigir roteamento (auth.wisa.labs.com.br, lingerie-maria.wisa.labs.com.br)
- [x] Fase 2 — Criar docker-compose.yml na pasta nginx para rodar NGINX com a config, permitindo testes isolados
- [x] Fase 3 — Adicionar serviço nginx ao docker-compose principal; alterar APP_URL e FRONTEND_URL na test-app

## Escopo

- **Objetivo:** Configurar NGINX como reverse proxy para executar testes de integração end-to-end (auth central, test-app) em ambiente Docker.
- **Restrições:** NÃO alterar nenhum código em `backend/` ou `frontend/`.
- **Subdomínios:** `auth.wisa.labs.com.br` (portal de auth + API), `lingerie-maria.wisa.labs.com.br` (test-app)
- **Conformidade:** docs/code_guidelines, docs/adrs, ADR-007 (NGINX), ADR-009 (VPS)

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | Fase 2 |

## Ordem Sugerida de Implementação

1. Fase 1 — Config NGINX e roteamento
2. Fase 2 — Docker Compose na pasta nginx
3. Fase 3 — Integração ao docker-compose principal

## Referências

- [docs/context.md](../../context.md)
- [docs/adrs/ADR-007-nginx-como-reverse-proxy.md](../../adrs/ADR-007-nginx-como-reverse-proxy.md)
- [docs/adrs/ADR-009-infraestrutura-vps-linux.md](../../adrs/ADR-009-infraestrutura-vps-linux.md)
- [docs/vps/configurations.md](../../vps/configurations.md)
- [docs/code_guidelines/](../../code_guidelines/)
