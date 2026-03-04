# Fase 6 — Índices, Row-Level Security e Triggers

## Objetivo

Implementar índices de performance, habilitar Row-Level Security (RLS) com políticas de isolamento por tenant, e criar triggers para atualização automática de `updated_at` em todas as tabelas que possuem essa coluna, conforme ADR-003 e ADR-008.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-003 e ADR-008 listam índices críticos para as operações mais frequentes:

**Tenants e Products:**
- idx_tenants_slug_status ON tenants (slug, status)
- idx_products_slug_status ON products (slug, status)

**Subscriptions:**
- idx_subscriptions_tenant ON subscriptions (tenant_id, status)
- idx_subscriptions_product ON subscriptions (product_id, status)
- idx_subscriptions_dates ON subscriptions (tenant_id, start_date, end_date)

**Payments:**
- idx_payments_subscription ON payments (subscription_id)
- idx_payments_status_date ON payments (status, payment_date)

**Users:**
- idx_users_tenant_email ON users (tenant_id, email) — consulta mais frequente no login
- idx_users_tenant_status ON users (tenant_id, status)

**User Product Access:**
- idx_user_product_access_user ON user_product_access (user_id)
- idx_user_product_access_tenant ON user_product_access (tenant_id, product_id)

**Refresh Tokens:**
- idx_refresh_tokens_hash ON refresh_tokens (token_hash) WHERE revoked_at IS NULL — partial index

**Audit Logs:**
- idx_audit_tenant_date ON audit_logs (tenant_id, created_at DESC)

### Ação 6.1

Criar migration 000007 com todos os índices. Usar `CREATE INDEX CONCURRENTLY` se a migration for executada em produção com tabelas já populadas — evita lock exclusivo prolongado. O golang-migrate em modo padrão não suporta CONCURRENTLY dentro de transação (CREATE INDEX CONCURRENTLY não pode rodar em transação). Opções:
1. Usar CREATE INDEX (sem CONCURRENTLY) — pode bloquear writes brevemente; aceitável em deploy inicial com poucos dados
2. Configurar migrate para rodar sem transação (run migrations outside transaction) — complexo
3. Para ambiente de desenvolvimento e deploy inicial: usar CREATE INDEX simples

### Observação 6.1

Para a estrutura inicial (tabelas vazias ou com poucos dados), CREATE INDEX sem CONCURRENTLY é adequado. Em produção futura com milhões de linhas, uma migration separada com CONCURRENTLY pode ser usada. Documentar essa consideração.

### Ação 6.1 (decisão)

Usar `CREATE INDEX` (sem CONCURRENTLY) nesta migration. Tabelas recém-criadas estão vazias; o lock é mínimo.

---

### Pensamento 2

Row-Level Security: habilitar em todas as tabelas que possuem `tenant_id`:
- subscriptions
- payments (sem tenant_id direto — policy via join; ou não habilitar RLS em payments se o acesso for sempre via subscriptions)
- users
- user_product_access
- refresh_tokens
- audit_logs

Tenants e products **não** possuem tenant_id (são entidades globais ou raiz). Não aplicar RLS nelas.

Para payments: o ADR-008 não inclui tenant_id. A policy de RLS teria que usar subquery. Alternativa: não habilitar RLS em payments e confiar nas camadas superiores (use case sempre acessa payments via subscription, que já é filtrada por tenant). O ADR-008 lista "ALTER TABLE payments ENABLE ROW LEVEL SECURITY" — então habilitar. A policy será: usuário pode ver payments cuja subscription pertence ao tenant atual. Exemplo:
```sql
CREATE POLICY tenant_isolation ON wisa_crm_db.payments
USING (
  subscription_id IN (
    SELECT id FROM wisa_crm_db.subscriptions
    WHERE tenant_id = current_setting('app.current_tenant_id', true)::uuid
  )
);
```

### Ação 6.2

Criar policies para cada tabela com tenant_id. A aplicação precisa executar `SET LOCAL app.current_tenant_id = '<uuid>'` no início de cada request/transação. O middleware ou handler deve fazer isso após extrair o tenant do JWT ou do request.

### Observação 6.2

Políticas RLS com `current_setting('app.current_tenant_id', true)` — o segundo parâmetro `true` faz com que retorne NULL se a setting não existir, evitando erro. Se NULL, a condição `tenant_id = NULL` não retorna linhas (NULL = NULL é NULL em SQL). Assim, sem a setting configurada, nenhum dado multi-tenant é retornado — comportamento seguro (fail-closed).

### Pensamento 2 (continuação)

Para operações de sistema (criação de tenant, login antes de ter tenant no JWT), a aplicação pode usar role com BYPASSRLS ou conectar com usuário superuser em operações específicas. O usuário normal do backend (wisa_crm) não é superuser — RLS será aplicado. Para login: o use case de autenticação identifica o tenant pelo tenant_slug no request, valida credenciais, e depois de autenticado o JWT terá tenant_id. Durante o login, a query busca user por tenant_id e email — o tenant_id vem do tenant_slug (lookup em tenants). A aplicação deve SET app.current_tenant_id antes de buscar o user. O fluxo: 1) Request com tenant_slug; 2) Buscar tenant por slug para obter tenant_id; 3) SET app.current_tenant_id = tenant_id; 4) Buscar user por tenant_id e email. A tabela tenants não tem RLS; a busca por slug é livre. Após obter tenant_id, setamos a variável e as demais queries (users, etc.) são filtradas. Correto.

### Ação 6.2 (complemento)

A policy para users, subscriptions, user_product_access, refresh_tokens, audit_logs usa:
```sql
USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid)
```
Para payments, usar a subquery com subscriptions.

Também é necessário policy para INSERT/UPDATE/DELETE. Para isolamento completo, as policies devem cobrir SELECT, INSERT, UPDATE, DELETE. A policy USING aplica-se a SELECT, UPDATE, DELETE. Para INSERT, usa-se WITH CHECK: `WITH CHECK (tenant_id = current_setting(...)::uuid)`. Isso garante que não se possa inserir linha com tenant_id diferente do contexto.

### Observação 6.2 (complemento)

Políticas completas: SELECT com USING; INSERT/UPDATE com WITH CHECK. Para tabelas com tenant_id, a mesma condição em USING e WITH CHECK.

---

### Pensamento 3

Triggers para `updated_at`: criar função `set_updated_at()` no schema wisa_crm_db e triggers em cada tabela que possui updated_at:
- tenants
- products
- subscriptions
- payments
- users
- user_product_access
- **refresh_tokens** — passa a incluir `updated_at` (útil quando o token é revogado via UPDATE em `revoked_at`)
- **audit_logs** — passa a incluir `updated_at` (permite rastrear correções ou emendas em registros de auditoria, quando aplicável)

Conforme alteração no ADR-008, as tabelas `refresh_tokens` e `audit_logs` passam a possuir a coluna `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`. A migration da Fase 5 (ou uma migration de alteração) deve adicionar essa coluna caso as tabelas já existam; ou a Fase 5 deve ser atualizada para incluir `updated_at` na criação original dessas tabelas.

### Ação 6.3

Criar função set_updated_at():
```sql
CREATE OR REPLACE FUNCTION wisa_crm_db.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

Criar triggers nas tabelas: tenants, products, subscriptions, payments, users, user_product_access, **refresh_tokens** e **audit_logs**. O trigger é BEFORE UPDATE FOR EACH ROW EXECUTE FUNCTION set_updated_at();

Em PostgreSQL 11+, usa-se EXECUTE FUNCTION (antigo EXECUTE PROCEDURE está deprecated).

### Observação 6.3

A sintaxe correta é `EXECUTE FUNCTION` ou `EXECUTE PROCEDURE`. Em PostgreSQL 11+, funções que retornam TRIGGER são invocadas com EXECUTE FUNCTION.

---

### Pensamento 4

Ordem na migration 000007:
1. Criar índices (não dependem de RLS nem triggers)
2. Criar função set_updated_at
3. Criar triggers de updated_at
4. Habilitar RLS nas tabelas
5. Criar políticas RLS

A ordem de políticas pode afetar se há dependências entre tabelas nas policies (ex: payments usa subquery em subscriptions). Ao criar a policy de payments, a tabela subscriptions já existe e tem RLS. A subquery na policy de payments é executada no contexto do usuário da sessão — a tabela subscriptions também terá RLS. Porém, a policy de subscriptions usa tenant_id; quando a policy de payments faz SELECT em subscriptions, a RLS de subscriptions será aplicada. Se app.current_tenant_id estiver setado, a subquery retornará subscriptions do tenant atual. Correto.

### Ação 6.4

Ordem na migration up: índices → função → triggers → ALTER TABLE ENABLE ROW LEVEL SECURITY → CREATE POLICY para cada tabela.

### Observação 6.4

Para tenants e products, não habilitar RLS. Apenas nas tabelas com tenant_id.

---

### Pensamento 5

Rollback: remover na ordem inversa.
1. DROP POLICY para cada tabela (ou ALTER TABLE ... DISABLE ROW LEVEL SECURITY — mas as policies precisam ser dropadas antes)
2. DROP TRIGGER para cada tabela
3. DROP FUNCTION set_updated_at
4. DROP INDEX para cada índice

Ao dropar tabelas em migrations anteriores (down das fases 2-5), os índices e triggers seriam dropados em cascata? Não — ao fazer down da 000007, estamos apenas revertendo a 000006. As tabelas continuam existindo. O down da 000007 deve: DROP POLICY (todas), DROP TRIGGER (todos), DROP FUNCTION, DROP INDEX (todos). A ordem para policies: pode dropar em qualquer ordem. Para triggers: DROP TRIGGER ... ON table_name; Para function: DROP FUNCTION após remover todos os triggers que a usam. Para índices: DROP INDEX ...;

### Ação 6.5

Migration down 000007:
1. DROP POLICY (para cada tabela com RLS)
2. ALTER TABLE ... DISABLE ROW LEVEL SECURITY (opcional, pois sem policy o RLS não filtra nada; mas desabilitar é mais limpo)
3. DROP TRIGGER (cada um)
4. DROP FUNCTION wisa_crm_db.set_updated_at
5. DROP INDEX (cada um)

### Observação 6.5

Ao dropar a função, os triggers que a usam devem ser dropados antes. Ordem: DROP TRIGGER primeiro, depois DROP FUNCTION.

---

### Pensamento 6

Configuração da aplicação: após esta fase, o backend Go precisa:
1. Configurar `search_path` na conexão para incluir wisa_crm_db (ou qualificar nomes de tabelas)
2. Executar `SET app.current_tenant_id = $1` no início de cada request que opera em dados multi-tenant
3. Para operações sem tenant (ex: health check, métricas), a configuração sem tenant_id faz com que as queries em tabelas RLS retornem vazio — pode ser aceitável para health que apenas verifica conectividade

O documento de configuração (docs/backend ou .env.example) deve ser atualizado para incluir a opção de search_path na DATABASE_URL. Exemplo para lib/pq: `postgres://user:pass@host/db?search_path=wisa-labs-db` — mas o schema tem hífen, precisa de aspas. A opção seria `options=-c search_path=wisa_crm_db`. Ou configurar no código após abrir a conexão: `db.Exec("SET search_path TO \"wisa-labs-db\"")` no startup.

### Ação 6.6

Documentar na fase ou em guia separado que:
- DATABASE_URL deve incluir search_path ou o código deve SET search_path após conectar
- O middleware de tenant deve executar SET LOCAL app.current_tenant_id antes de operações de dados

### Observação 6.6

Essas configurações são de aplicação, não de migration. A migration apenas cria a estrutura. O planejamento documenta o que a aplicação precisará fazer.

---

### Decisão final Fase 6

**Implementar:**
1. Migration `000007_add_indexes_rls_triggers.up.sql`:
   - CREATE INDEX para todos os índices listados
   - CREATE FUNCTION set_updated_at
   - CREATE TRIGGER em tenants, products, subscriptions, payments, users, user_product_access, refresh_tokens, audit_logs
   - ALTER TABLE ENABLE ROW LEVEL SECURITY em subscriptions, payments, users, user_product_access, refresh_tokens, audit_logs
   - CREATE POLICY tenant_isolation para cada uma (payments com policy especial via subquery)
2. Migration `000007_add_indexes_rls_triggers.down.sql`:
   - DROP POLICY (todas)
   - DROP TRIGGER (todos)
   - DROP FUNCTION
   - DROP INDEX (todos)

---

### Checklist de Implementação

1. [ ] Criar todos os índices conforme ADR-008
2. [ ] Criar função set_updated_at no schema wisa_crm_db
3. [ ] Criar triggers BEFORE UPDATE em 8 tabelas (tenants, products, subscriptions, payments, users, user_product_access, refresh_tokens, audit_logs)
4. [ ] Habilitar RLS em 6 tabelas (subscriptions, payments, users, user_product_access, refresh_tokens, audit_logs)
5. [ ] Criar policy para subscriptions, users, user_product_access, refresh_tokens, audit_logs (USING e WITH CHECK)
6. [ ] Criar policy para payments com subquery em subscriptions
7. [ ] Criar migration down completo
8. [ ] Atualizar documentação com search_path e app.current_tenant_id
9. [ ] Testar que RLS bloqueia acesso cross-tenant (teste manual ou integração)

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-003 RLS | Políticas de isolamento por tenant_id |
| ADR-008 Índices | Todos os índices documentados |
| ADR-008 Triggers | set_updated_at em tabelas com updated_at |
| Defense in depth | RLS como camada 5 de isolamento (ADR-008) |

---

## Referências

- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../../adrs/ADR-008-arquitetura-multi-tenant.md)
- [PostgreSQL Row-Level Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [PostgreSQL CREATE INDEX](https://www.postgresql.org/docs/current/sql-createindex.html)
