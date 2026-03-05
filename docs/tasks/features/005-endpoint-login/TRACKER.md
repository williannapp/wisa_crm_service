# Feature 005 — Endpoint de Login — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Entidades, interfaces de repositório e models de persistência | 1/1 | Concluída |
| Fase 2 | Implementações dos repositórios GORM | 1/1 | Concluída |
| Fase 3 | Serviços de infraestrutura (Password, JWT RS256) | 1/1 | Concluída |
| Fase 4 | Use Case AuthenticateUser com validações completas | 1/1 | Concluída |
| Fase 5 | Handler HTTP e rota POST /api/v1/auth/login | 1/1 | Concluída |
| Fase 6 | Wiring no main.go e configurações (.env) | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-entidades-repositorios-models.md](./fase-1-entidades-repositorios-models.md)
- [fase-2-implementacoes-repositorios.md](./fase-2-implementacoes-repositorios.md)
- [fase-3-servicos-crypto.md](./fase-3-servicos-crypto.md)
- [fase-4-usecase-authenticate.md](./fase-4-usecase-authenticate.md)
- [fase-5-handler-e-rota.md](./fase-5-handler-e-rota.md)
- [fase-6-wiring-e-config.md](./fase-6-wiring-e-config.md)

## Resumo das Tasks

- [x] Fase 1 — Criar entidades de domínio, interfaces de repositório e models GORM (aud do JWT = slug + base domain, sem novo campo no banco)
- [x] Fase 2 — Implementar repositórios GORM (Tenant, Product, User, Subscription, UserProductAccess)
- [x] Fase 3 — Implementar PasswordService (bcrypt) e JWTService (RS256)
- [x] Fase 4 — Implementar Use Case AuthenticateUser com fluxo de validações (tenant, product, email, senha, status, assinatura) e emissão JWT
- [x] Fase 5 — Criar handler de login, DTOs, rota e integração com ErrorMapper
- [x] Fase 6 — Wiring de dependências no main.go e variáveis de ambiente

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | — |
| Fase 4 | Fase 1, Fase 2, Fase 3 |
| Fase 5 | Fase 4 |
| Fase 6 | Fase 4, Fase 5 |

## Ordem Sugerida de Implementação

1. Fase 1 — Entidades e interfaces
2. Fase 2 — Repositórios GORM
3. Fase 3 — Serviços crypto (paralelizável com Fase 2)
4. Fase 4 — Use Case
5. Fase 5 — Handler e rota
6. Fase 6 — Wiring final

## Notas Importantes

- **user_access_profile:** A especificação menciona `admin`, `editor`, `viewer`. O banco (enum `access_profile`) usa `admin`, `operator`, `view`. Usar os valores do banco como fonte da verdade. Se for necessário alinhar no futuro, criar migration separada.
- **iss (JWT):** Configurável via `JWT_ISSUER` (default `"wisa-crm-service"`) para facilitar adaptação futura ao domínio (ex: `https://auth.wisa-crm.com`).
- **aud (JWT):** Construído como `tenant.Slug + "." + JWT_AUD_BASE_DOMAIN` (ex: `cliente1.app.wisa-crm.com`). Sem campo extra no banco; o slug vem no redirect do sistema do cliente.
- **URL e product_slug:** A URL segue `https://slug.domain.com.br/<product_slug>`. O login recebe `slug` e `product_slug`; a assinatura é validada por tenant_id + product_id.

## Referências

- [docs/context.md](../../context.md)
- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../adrs/ADR-005-clean-architecture.md)
- [docs/adrs/ADR-006-jwt-com-assinatura-assimetrica.md](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../adrs/ADR-008-arquitetura-multi-tenant.md)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
