# Fase 1 — Schema wisa_crm_db e Tipos ENUM

## Objetivo

Criar o schema `wisa_crm_db` no banco de dados PostgreSQL e definir todos os tipos ENUM necessários para as tabelas do sistema, conforme modelagem do ADR-008 e requisitos do projeto.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O schema adotado é `wisa_crm_db`, alinhado ao nome do database do projeto. Em PostgreSQL:
- Um **schema** é um namespace dentro de um **database**
- O projeto já usa o database `wisa_crm_db` (docker-compose, vps-configurations)
- O schema `wisa_crm_db` usa apenas underscores — não requer aspas no SQL
- Após criar o schema, todas as tabelas serão criadas dentro dele para isolamento lógico

### Ação 1.1

Definir a migration que criará o schema:
- Arquivo: `backend/migrations/000002_create_wisa_crm_schema.up.sql`
- Comando: `CREATE SCHEMA IF NOT EXISTS wisa_crm_db;`
- O `IF NOT EXISTS` garante idempotência em re-execução acidental

### Observação 1.1

O schema `wisa_crm_db` mantém consistência com o nome do database e não requer aspas por usar apenas letras e underscores.

---

### Pensamento 2

Após criar o schema, os tipos ENUM devem ser criados **dentro** desse schema para que as tabelas do mesmo schema possam referenciá-los sem qualificação completa. A sintaxe em PostgreSQL é:
- `CREATE TYPE wisa_crm_db.tenant_type AS ENUM ('person', 'company');`

Ou podemos criar no schema `wisa_crm_db` definindo o search_path antes:
- `SET search_path TO wisa_crm_db;`
- `CREATE TYPE tenant_type AS ENUM (...);`

A abordagem com qualificação explícita `wisa_crm_db.tipo` torna as migrations mais verbosas porém explícitas e menos sujeitas a search_path. Para consistência com golang-migrate (cada statement é independente), usar qualificação explícita.

### Ação 1.2

Definir os ENUMs conforme ADR-008. Os tipos documentados são:
- `tenant_type`: 'person', 'company'
- `tenant_status`: 'active', 'inactive', 'blocked'
- `product_status`: 'active', 'inactive', 'blocked'
- `subscription_status`: 'pending', 'active', 'suspended', 'canceled'
- `subscription_type`: 'free', 'payment'
- `payment_status`: 'pending', 'paid', 'failed', 'refunded'
- `user_status`: 'active', 'blocked'
- `access_profile`: 'admin', 'operator', 'view'

### Observação 1.2

Todos os valores devem ser strings em minúsculas para consistência e para evitar problemas com case-sensitivity no PostgreSQL (ENUMs são case-sensitive na comparação).

---

### Pensamento 3

O golang-migrate executa migrations na ordem numérica. A migration 000001 já existe (init_schema com extensão uuid-ossp). A nova migration deve ser 000002. O down da migration deve reverter na ordem inversa: DROP TYPE para cada ENUM (na ordem que respeite dependências — ENUMs não têm dependências entre si), e por fim DROP SCHEMA.

Em PostgreSQL, para dropar um schema que contém objetos, usa-se `DROP SCHEMA ... CASCADE`. Isso remove schema e todos os objetos dentro dele. Se criarmos ENUMs no schema, o CASCADE os remove automaticamente. Porém, a boa prática é fazer o down explícito: DROP TYPE de cada um, depois DROP SCHEMA. Isso evita CASCADE agressivo e documenta a ordem correta de rollback.

Porém: ENUMs não podem ser dropados se estiverem em uso por tabelas. Na migration 000002, apenas criamos schema e ENUMs — ainda não há tabelas. O down da 000002 pode fazer DROP TYPE e DROP SCHEMA sem problema. As migrations futuras (000003, 000004, etc.) que criam tabelas terão seus próprios downs que dropam tabelas primeiro, e em última instância a migração 000002 nunca seria "downada" sem antes fazer down das que dependem dela.

Estrutura de migrations:
- 000002: schema + ENUMs (up e down)
- 000003+: tabelas (cada fase uma ou mais migrations)

OU uma única migration 000002 com schema + ENUMs + tabelas? O usuário pediu para dividir em fases. Cada fase pode corresponder a uma migration ou a uma seção dentro de uma migration maior. Para rollback granular, **uma migration por fase** é preferível:
- 000002: schema + ENUMs
- 000003: tenants + products
- 000004: subscriptions + payments
- 000005: users + user_product_access
- 000006: refresh_tokens + audit_logs
- 000007: índices + RLS + triggers

### Ação 1.3

Adotar **uma migration por fase**. Fase 1 = migration 000002.
- `000002_create_wisa_crm_schema.up.sql`: CREATE SCHEMA + CREATE TYPE para cada ENUM no schema wisa_crm_db
- `000002_create_wisa_crm_schema.down.sql`: DROP TYPE para cada ENUM (ordem arbitrária, pois não há dependência entre ENUMs), depois DROP SCHEMA wisa_crm_db

### Observação 1.3

Ao dropar o schema, o PostgreSQL exige que o schema esteja vazio. Ao dropar os tipos, o schema fica vazio. Ordem: DROP TYPE ... (todos os 8), DROP SCHEMA wisa_crm_db.

---

### Pensamento 4

A aplicação (GORM) e as migrations precisarão usar o schema nas queries. Opções:
1. **search_path na conexão**: Ao conectar, executar `SET search_path TO wisa_crm_db, public;` — assim as tabelas são encontradas automaticamente
2. **Tabelas qualificadas**: Usar `wisa_crm_db.tenants` em todas as queries

O ADR-008 e ADR-003 recomendam row-level tenancy, não schema-per-tenant. Aqui temos um **único** schema de aplicação (wisa_crm_db) para organizar as tabelas. O padrão da indústria é configurar `search_path` na connection string ou no código de conexão para que o backend use o schema por padrão.

### Ação 1.4

Documentar na fase que, após a migration, a `DATABASE_URL` ou o código de conexão deve incluir `search_path=wisa_crm_db` na connection string. Exemplo de URL: `postgres://user:pass@host:5432/wisa_crm_db?search_path=wisa_crm_db`

### Observação 1.4

O golang-migrate usa a mesma DATABASE_URL. As migrations 000002+ devem qualificar explicitamente o schema nos comandos CREATE (ex: `CREATE TABLE wisa_crm_db.tenants ...`) para garantir que os objetos sejam criados no local correto, independentemente do search_path da conexão. Assim a migration é autossuficiente.

---

### Pensamento 5

Verificação de conformidade com ADRs e guidelines:
- ADR-003: PostgreSQL, schema compartilhado com tenant_id — o schema "wisa_crm_db" é o namespace da aplicação
- ADR-008: Modelagem completa com ENUMs — os tipos listados cobrem todos os status necessários
- Code guidelines: migrations versionadas, nomenclatura 00000X_descricao.up/down.sql
- Segurança: schema não expõe credenciais; ENUMs não são sensíveis

### Ação 1.5

Checklist de conformidade:
- [ ] ADR-003: schema dentro do database wisa_crm_db existente
- [ ] ADR-008: todos os 8 ENUMs conforme documentação
- [ ] ADR-004: migrations via golang-migrate, não AutoMigrate
- [ ] Nomenclatura: 000002_create_wisa_crm_schema

### Observação 1.5

Nenhuma falha de segurança identificada. O schema apenas organiza os objetos; a segurança multi-tenant vem do tenant_id nas linhas e do RLS (Fase 6).

---

### Decisão final Fase 1

**Implementar:**
1. Criar migration `000002_create_wisa_crm_schema.up.sql`:
   - `CREATE SCHEMA IF NOT EXISTS wisa_crm_db;`
   - `CREATE TYPE wisa_crm_db.tenant_type AS ENUM ('person', 'company');`
   - (repetir para os 8 tipos ENUM)
2. Criar migration `000002_create_wisa_crm_schema.down.sql`:
   - `DROP TYPE IF EXISTS wisa_crm_db.access_profile;` (e demais tipos)
   - `DROP SCHEMA IF EXISTS wisa_crm_db;`
3. Documentar a necessidade de configurar search_path na conexão do backend para desenvolvimento e produção

---

### Checklist de Implementação

1. [ ] Criar `backend/migrations/000002_create_wisa_crm_schema.up.sql` com CREATE SCHEMA e 8 CREATE TYPE
2. [ ] Criar `backend/migrations/000002_create_wisa_crm_schema.down.sql` com 8 DROP TYPE e DROP SCHEMA
3. [ ] Ordem no down: dropar tipos na ordem inversa da criação (ou qualquer ordem, pois não há dependência entre ENUMs)
4. [ ] Testar: `go run ./cmd/migrate` (up) e `go run ./cmd/migrate -down` (verificar rollback)
5. [ ] Atualizar `backend/.env.example` ou documentação com opção de search_path na DATABASE_URL (opcional para esta fase; pode ser Fase 6)

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Schema wisa_crm_db | CREATE SCHEMA wisa_crm_db |
| ADR-008 ENUMs | 8 tipos: tenant_type, tenant_status, product_status, subscription_status, subscription_type, payment_status, user_status, access_profile |
| ADR-004 | Migrations via golang-migrate |
| Code Guidelines | Nomenclatura 000002_descricao |
| Reversibilidade | down com DROP TYPE e DROP SCHEMA |

---

## Referências

- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../../adrs/ADR-008-arquitetura-multi-tenant.md)
- [PostgreSQL CREATE SCHEMA](https://www.postgresql.org/docs/current/sql-createschema.html)
- [PostgreSQL CREATE TYPE](https://www.postgresql.org/docs/current/sql-createtype.html)
