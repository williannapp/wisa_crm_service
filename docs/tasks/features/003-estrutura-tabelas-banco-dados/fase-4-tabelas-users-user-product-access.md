# Fase 4 — Tabelas users e user_product_access

## Objetivo

Implementar as tabelas `users` e `user_product_access` no schema `wisa-labs-db`, vinculando usuários a tenants e definindo o perfil de acesso por produto, conforme ADR-008.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A tabela `users` armazena credenciais e dados dos usuários por tenant:
- `id` UUID PRIMARY KEY
- `tenant_id` UUID NOT NULL REFERENCES wisa_crm_db.tenants(id)
- `name` VARCHAR(255) NOT NULL
- `email` VARCHAR(320) NOT NULL — limite RFC 5321
- `password_hash` VARCHAR(72) NOT NULL — bcrypt produz até 60 chars, 72 permite margem
- `status` user_status NOT NULL DEFAULT 'active'
- `last_login_at` TIMESTAMPTZ — opcional
- `created_at`, `updated_at` TIMESTAMPTZ
- UNIQUE (tenant_id, email) — email único por tenant

O ADR-010 menciona account lockout com `failed_attempts` e `locked_until` — essas colunas **não** constam no ADR-008. Para esta feature (estrutura inicial), seguir estritamente o ADR-008. Uma migration futura pode adicionar essas colunas quando implementar o lockout.

### Ação 4.1

Criar tabela users conforme ADR-008, sem failed_attempts e locked_until. Documentar no planejamento que ADR-010 requer lockout — colunas adicionais em feature futura.

### Observação 4.1

Manter consistência com ADR-008. A estrutura base fica correta; extensões são tratadas em migrations incrementais.

---

### Pensamento 2

A tabela `user_product_access` associa usuário a produto com perfil de acesso:
- `id` UUID PRIMARY KEY
- `user_id` UUID NOT NULL REFERENCES wisa_crm_db.users(id)
- `product_id` UUID NOT NULL REFERENCES wisa_crm_db.products(id)
- `tenant_id` UUID NOT NULL REFERENCES wisa_crm_db.tenants(id) — desnormalizado para RLS
- `access_profile` access_profile NOT NULL DEFAULT 'view'
- `created_at`, `updated_at` TIMESTAMPTZ
- UNIQUE (user_id, product_id)

O ADR-008 observa: "tenant_id deve ser igual ao tenant do user; validar via trigger ou na aplicação". Uma CHECK constraint ou trigger pode garantir que user.tenant_id = user_product_access.tenant_id. Em SQL: pode-se criar trigger BEFORE INSERT OR UPDATE que verifica se (SELECT tenant_id FROM users WHERE id = NEW.user_id) = NEW.tenant_id.

### Ação 4.2

Incluir um trigger de validação ou CHECK. CHECK não pode fazer subquery diretamente em PostgreSQL de forma simples. A opção é um trigger BEFORE INSERT OR UPDATE que faz a verificação. Alternativa: confiar na aplicação (camada de validação no Use Case). O ADR-008 diz "validar via trigger ou na aplicação" — ambas são aceitáveis. Para defense in depth, incluir **trigger** na migration que valida a consistência tenant_id.

### Observação 4.2

Trigger adiciona complexidade à migration. A Fase 6 já terá triggers (set_updated_at). Podemos incluir o trigger de validação de tenant_id em user_product_access nesta fase ou na Fase 6. Para manter Fase 6 focada em RLS e updated_at, incluir o trigger de validação **nesta fase** junto com a tabela user_product_access.

### Ação 4.2 (revisão)

O trigger de validação requer uma função. Criar função `validate_user_product_access_tenant()` que retorna NEW ou raises exception se tenant_id do user não bater. Chamar no trigger. Isso adiciona complexidade. Alternativa mais simples: **não** criar trigger nesta fase; documentar que a aplicação deve validar. O RLS e o Global Scope do GORM já filtram por tenant; um bug que insira user_product_access com tenant_id errado poderia criar registro órfão, mas o RLS impediria que o tenant errado acessasse. Por pragmatismo, **não** implementar trigger de validação na estrutura inicial — deixar para a camada de aplicação. O ADR permite "na aplicação".

### Observação 4.2 (revisão)

Decisão: seguir ADR-008, validar na aplicação. Não criar trigger de consistência tenant_id nesta fase.

---

### Pensamento 3

Segurança da coluna password_hash: nunca logar, nunca retornar em APIs. A estrutura do banco apenas define o tipo. O ADR-003 menciona possível column-level encryption via pgcrypto para campos ultrasensíveis — password_hash é sensível. Por ora, armazenar como VARCHAR (bcrypt output já é hash, não plaintext). Criptografia adicional em repouso pode ser feita via filesystem encryption ou TDE — fora do escopo desta feature.

### Ação 4.3

Manter password_hash como VARCHAR(72). Nenhuma alteração adicional de segurança no schema.

### Observação 4.3

Bcrypt hashes são one-way. O formato é adequado. Obrigatoriedade de não logar/retornar fica na aplicação (guidelines).

---

### Pensamento 4

Ordem de criação: users depende de tenants; user_product_access depende de users, products, tenants. Criar users primeiro, depois user_product_access.

### Ação 4.4

Migration 000005: CREATE TABLE users; CREATE TABLE user_product_access;

### Observação 4.4

Ambas as tabelas possuem tenant_id — serão protegidas por RLS na Fase 6.

---

### Pensamento 5

Rollback: DROP user_product_access (tem FK para users), depois DROP users.

### Ação 4.5

Migration down: DROP TABLE IF EXISTS wisa_crm_db.user_product_access; DROP TABLE IF EXISTS wisa_crm_db.users;

### Observação 4.5

Ordem correta. ON DELETE nas FKs: se user for deletado, o que acontece com user_product_access? RESTRICT por padrão — impede deletar user com registros em user_product_access. A aplicação deve remover acessos antes de remover usuário, ou usar CASCADE em cenários específicos (não recomendado para dados de acesso).

---

### Decisão final Fase 4

**Implementar:**
1. Migration `000005_create_users_and_user_product_access.up.sql`:
   - CREATE TABLE users com UNIQUE(tenant_id, email)
   - CREATE TABLE user_product_access com UNIQUE(user_id, product_id)
2. Migration `000005_create_users_and_user_product_access.down.sql`:
   - DROP TABLE user_product_access, users

---

### Checklist de Implementação

1. [ ] Criar 000005_create_users_and_user_product_access.up.sql
2. [ ] Garantir UNIQUE (tenant_id, email) em users
3. [ ] Garantir UNIQUE (user_id, product_id) em user_product_access
4. [ ] Incluir tenant_id em user_product_access para RLS
5. [ ] Criar migration down
6. [ ] Documentar que lockout (failed_attempts, locked_until) será migration futura se ADR-010 for implementado

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-008 users | tenant_id, name, email, password_hash, status, last_login_at, UNIQUE |
| ADR-008 user_product_access | user_id, product_id, tenant_id, access_profile, UNIQUE |
| Segurança | password_hash não plaintext; validação de tenant na aplicação |
| Multi-tenant | Ambas tabelas com tenant_id |

---

## Referências

- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../../adrs/ADR-008-arquitetura-multi-tenant.md)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
