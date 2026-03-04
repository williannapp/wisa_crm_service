# Fase 5 — Tabelas refresh_tokens e audit_logs

## Objetivo

Implementar as tabelas `refresh_tokens` e `audit_logs` no schema `wisa_crm_db`, para controle de sessões de autenticação e rastreabilidade de eventos de segurança, conforme ADR-008 e ADR-003.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A tabela `refresh_tokens` armazena tokens de renovação de sessão:
- `id` UUID PRIMARY KEY
- `user_id` UUID NOT NULL REFERENCES wisa_crm_db.users(id)
- `tenant_id` UUID NOT NULL REFERENCES wisa_crm_db.tenants(id) — para RLS
- `token_hash` CHAR(64) NOT NULL UNIQUE — hash SHA-256 do token (64 chars hex)
- `expires_at` TIMESTAMPTZ NOT NULL
- `revoked_at` TIMESTAMPTZ — null se ativo
- `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
- `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW() — data da última atualização (ex.: quando revogado)
- `ip_address` INET — opcional
- `user_agent` TEXT — opcional

O token em si nunca é armazenado — apenas o hash. Isso previne vazamento em caso de dump do banco.

### Ação 5.1

Criar tabela refresh_tokens. CHAR(64) para SHA-256 em hexadecimal. UNIQUE em token_hash para evitar colisões e para lookup eficiente na renovação.

### Observação 5.1

Conforme ADR-010, Refresh Token Rotation: ao renovar, o token antigo é revogado (revoked_at = NOW()) e um novo é criado. O índice parcial `WHERE revoked_at IS NULL` será criado na Fase 6 para otimizar consultas de tokens ativos.

---

### Pensamento 2

A tabela `audit_logs` registra eventos de segurança (login_success, login_failed, etc.):
- `id` UUID PRIMARY KEY
- `tenant_id` UUID NOT NULL REFERENCES tenants(id)
- `user_id` UUID REFERENCES users(id) — NULL quando login com email inválido (usuário não existe)
- `event_type` VARCHAR(50) NOT NULL
- `ip_address` INET
- `user_agent` TEXT
- `metadata` JSONB — dados adicionais flexíveis
- `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
- `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW() — data da última atualização (emendas/correções, quando aplicável)

O ADR-008 define `PARTITION BY RANGE (created_at)` para audit_logs. Tabelas particionadas no PostgreSQL requerem:
1. Criar a tabela pai com PARTITION BY
2. Criar partições (tabelas filhas) para cada intervalo de dados

### Ação 5.2

Para a migration inicial, criar a tabela particionada e **pelo menos uma partição**. Exemplo: partição para dados a partir de 2026-01-01. Sem partições, INSERT na tabela pai falha. O ADR-003 sugere rotação: partições mensais, manter 12 meses. Para a estrutura inicial, criar uma partição com range "início dos tempos" até "fim de 2026" (ou um range amplo). Exemplo: `PARTITION audit_logs_2026 VALUES FROM ('2026-01-01') TO ('2027-01-01')`. Uma partição default ou range amplo permite que a migration funcione imediatamente.

### Observação 5.2

Em PostgreSQL 10+, a sintaxe é:
```sql
CREATE TABLE audit_logs (...)
PARTITION BY RANGE (created_at);

CREATE TABLE audit_logs_2026 PARTITION OF wisa_crm_db.audit_logs
FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
```

Precisamos também de uma partição para dados antes de 2026 (ex: audit_logs_default) — mas PARTITION BY RANGE não permite partição "default" antes do 11. Em PostgreSQL 11+, existe PARTITION BY RANGE com DEFAULT partition. Ou usamos PARTITION BY RANGE sem default e criamos partição que cubra desde '1970-01-01' até '2027-01-01' para garantir que qualquer timestamp de 2026 caiba. Depois, em migrations futuras, adicionar partições mensais. Para simplificar a estrutura inicial: uma partição `audit_logs_default` com range de 1970-01-01 a 2030-01-01.

### Ação 5.2 (revisão)

Criar partição com range amplo: `FOR VALUES FROM ('1970-01-01') TO ('2030-01-01')`. Isso cobre todos os inserts no horizonte próximo. Uma migration futura pode adicionar particionamento mensal mais granular e rotacionar partições antigas.

### Observação 5.2 (revisão)

Prático para o estágio atual. Documentar no down que ao dropar audit_logs, as partições são dropadas em cascata pela tabela pai.

---

### Pensamento 3

Conforme alteração no ADR-008, as tabelas `refresh_tokens` e `audit_logs` passam a incluir a coluna `updated_at`. Para refresh_tokens: útil quando o token é revogado via UPDATE em `revoked_at`. Para audit_logs: permite rastrear correções ou emendas em registros de auditoria, quando aplicável. O trigger `set_updated_at` será criado na Fase 6 para ambas as tabelas.

### Ação 5.3

Incluir `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()` em refresh_tokens e audit_logs.

### Observação 5.3

Conformidade com ADR-008 atualizado. O modelo segue o padrão das demais tabelas do sistema.

---

### Pensamento 4

Ordem de criação: refresh_tokens depende de users e tenants; audit_logs depende de tenants e users (opcional). Criar refresh_tokens primeiro, depois audit_logs (ou em paralelo, pois não há dependência entre elas).

### Ação 5.4

Migration 000006: CREATE TABLE refresh_tokens; CREATE TABLE audit_logs (... PARTITION BY RANGE); CREATE TABLE audit_logs_YYYY PARTITION OF audit_logs FOR VALUES FROM (...) TO (...);

### Observação 5.4

A partição deve ser criada no mesmo schema. A tabela pai audit_logs pertence ao schema wisa_crm_db; as partições herdam o schema da tabela pai.

---

### Pensamento 5

Rollback: Dropar audit_logs (as partições são removidas em cascata ao dropar a tabela pai). Dropar refresh_tokens.

### Ação 5.5

Migration down:
- DROP TABLE IF EXISTS wisa_crm_db.audit_logs; — CASCADE remove partições
- DROP TABLE IF EXISTS wisa_crm_db.refresh_tokens;

### Observação 5.5

Em PostgreSQL, ao dropar uma tabela particionada, as partições são automaticamente dropadas. Não é necessário dropar partições explicitamente.

---

### Pensamento 6

Segurança: token_hash jamais deve ser logado em texto. O hash é seguro para armazenar. Audit_logs pode conter metadata sensível — a aplicação deve sanitizar o que vai em metadata (nunca senhas, tokens completos). O event_type e dados estruturados são suficientes para rastreabilidade.

### Ação 5.6

Documentar nas guidelines que metadata em audit_logs não deve conter credenciais. A estrutura do banco está adequada.

### Observação 5.6

Segurança em profundidade — aplicação e code review.

---

### Decisão final Fase 5

**Implementar:**
1. Migration `000006_create_refresh_tokens_and_audit_logs.up.sql`:
   - CREATE TABLE refresh_tokens
   - CREATE TABLE audit_logs PARTITION BY RANGE (created_at)
   - CREATE TABLE audit_logs_2026 (ou nome similar) PARTITION OF audit_logs FOR VALUES FROM ('1970-01-01') TO ('2030-01-01')
2. Migration `000006_create_refresh_tokens_and_audit_logs.down.sql`:
   - DROP TABLE audit_logs (partições em cascata)
   - DROP TABLE refresh_tokens

---

### Checklist de Implementação

1. [ ] Criar 000006_create_refresh_tokens_and_audit_logs.up.sql
2. [ ] Incluir token_hash CHAR(64) UNIQUE, revoked_at, updated_at em refresh_tokens
3. [ ] Criar audit_logs com PARTITION BY RANGE (created_at)
4. [ ] Criar pelo menos uma partição com range que cubra 2026-2030
5. [ ] user_id em audit_logs como REFERENCES users(id) — nullable; updated_at incluído
6. [ ] Criar migration down
7. [ ] Testar INSERT em audit_logs e refresh_tokens

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-008 refresh_tokens | user_id, tenant_id, token_hash, expires_at, revoked_at, updated_at, ip, user_agent |
| ADR-008 audit_logs | tenant_id, user_id nullable, event_type, metadata JSONB, updated_at, particionamento |
| ADR-003 | Particionamento para crescimento controlado do audit_log |
| Segurança | Apenas hash do token armazenado |

---

## Referências

- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../../adrs/ADR-008-arquitetura-multi-tenant.md)
- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [PostgreSQL Declarative Partitioning](https://www.postgresql.org/docs/current/ddl-partitioning.html)
