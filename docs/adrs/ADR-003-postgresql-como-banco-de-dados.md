# ADR-003 — PostgreSQL como Banco de Dados

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (persistência)

---

## Contexto

O `wisa-crm-service` precisa persistir e consultar dados críticos de segurança e negócio:

- **Tenants:** dados dos clientes contratantes, status de assinatura, plano
- **Usuários:** credenciais (hash de senha), perfil, vínculo com tenant
- **Assinaturas:** status, data de vencimento, período de tolerância, histórico
- **Sessões / Refresh Tokens:** controle de tokens ativos e revogação
- **Audit Log:** registro de autenticações, falhas, bloqueios (rastreabilidade)
- **Chaves públicas por tenant** (se cada tenant tiver sua própria keypair no futuro)

O banco de dados do `wisa-crm-service` é um **ponto crítico de falha e segurança**. Compromisso do banco significa:
- Exposição de hashes de senha de todos os tenants
- Possibilidade de manipulação de status de assinatura
- Perda de rastreabilidade de eventos de segurança

Requisitos específicos do banco:
- **ACID compliance** total — transações de autenticação + validação de assinatura devem ser atômicas
- Suporte a **multi-tenant** com isolamento lógico por `tenant_id`
- **Concorrência alta** para leitura (validações de token frequentes) e escrita moderada (logins, renovações)
- **Confiabilidade e durabilidade** — dados de billing e credenciais não podem ser perdidos
- Suporte a **row-level security** para reforço do isolamento entre tenants
- Capacidade de executar queries complexas de auditoria e relatório

---

## Decisão

**PostgreSQL 16+ é o banco de dados escolhido para o `wisa-crm-service`.**

Uma instância PostgreSQL rodará na própria VPS central, com configuração de hardening, backup automatizado e monitoramento contínuo.

---

## Justificativa

### 1. ACID compliance e integridade transacional

Autenticação e billing são domínios onde a consistência é inegociável. PostgreSQL oferece:

- **Transações ACID completas** com isolamento configurável (READ COMMITTED padrão, SERIALIZABLE quando necessário)
- **MVCC (Multi-Version Concurrency Control)** — leitores não bloqueiam escritores e vice-versa, fundamental para alta concorrência
- **Foreign keys com constraint enforcement** — impossibilita criar usuário sem tenant válido ou assinatura sem tenant existente
- **CHECK constraints** diretamente no banco — a regra de negócio "data de vencimento deve ser futura" pode ser enforçada no schema

Isso cria uma camada de proteção de integridade independente da aplicação.

### 2. Row-Level Security (RLS) para multi-tenant

PostgreSQL suporta **Row-Level Security nativamente**, permitindo que políticas de isolamento entre tenants sejam enforçadas no próprio banco:

```sql
-- Exemplo conceitual
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON users
  USING (tenant_id = current_setting('app.current_tenant_id')::uuid);
```

Isso significa que mesmo um bug na camada de aplicação que omita o filtro por `tenant_id` não resultará em vazamento de dados entre tenants. A segurança é aplicada em profundidade (defense in depth).

### 3. Controle de concorrência granular

Para o `wisa-crm-service`, os padrões de acesso são:

- **Alta frequência de leitura:** validação de sessões, consulta de status de assinatura a cada requisição
- **Escrita moderada:** login, logout, renovação de token, atualização de assinatura
- **Escrita crítica e exclusiva:** criação de tenant, bloqueio de conta

PostgreSQL oferece:
- **Advisory Locks** para operações críticas sem contention no nível de linha
- `SELECT FOR UPDATE SKIP LOCKED` para filas de processamento de eventos
- **Partial indexes** para otimizar queries frequentes (ex: `WHERE status = 'active' AND tenant_id = $1`)
- **Connection pooling** via PgBouncer para gerenciar o limite físico de conexões do PostgreSQL

### 4. Segurança nativa robusta

PostgreSQL oferece mecanismos de segurança maduros:

- **pg_hba.conf** para controle granular de autenticação por IP, usuário e banco
- **SSL/TLS obrigatório** para conexões do backend Go ao banco
- **Roles e privileges** com princípio do menor privilégio
- **Column-level encryption** via `pgcrypto` para campos ultrasenssíveis
- **Audit logging** via extensão `pgaudit` para rastreabilidade de queries
- **Prepared statements** nativos que eliminam SQL injection ao nível do protocolo de wire

### 5. Confiabilidade e maturidade

PostgreSQL tem 35+ anos de desenvolvimento ativo, com:
- Zero tolerância a corrupção de dados por design
- WAL (Write-Ahead Logging) garante durabilidade mesmo em crash
- Point-in-Time Recovery (PITR) para restauração precisa após falhas
- Streaming replication nativa para futuras necessidades de read replicas

Para um sistema de auth onde a perda de dados de assinatura pode resultar em disputas comerciais, confiabilidade não é opcional.

### 6. Extensibilidade para casos de uso avançados

O ecossistema de extensões do PostgreSQL cobre necessidades futuras:

- `pgcrypto` — criptografia simétrica e assimétrica no banco
- `pg_trgm` — busca por similaridade de texto (útil para busca de tenants)
- `uuid-ossp` ou `gen_random_uuid()` — geração de UUIDs para primary keys
- `pg_stat_statements` — monitoramento de queries para identificar gargalos
- `timescaledb` — caso o audit log cresça e necessite de time-series management

---

## Modelagem Multi-tenant Recomendada

### Estratégia: Schema Shared, Row-Level Isolation

Para o `wisa-crm-service`, a estratégia recomendada é **tabelas compartilhadas com `tenant_id` em todas as entidades**, combinada com Row-Level Security:

```
┌─────────────────────────────────────────────────────┐
│                  wisa_crm_db                        │
│                                                     │
│  tenants (id, name, status, plan, created_at)       │
│  subscriptions (id, tenant_id, expires_at, status)  │
│  users (id, tenant_id, email, password_hash, ...)   │
│  refresh_tokens (id, user_id, tenant_id, expires_at)│
│  audit_logs (id, tenant_id, user_id, event, ts)     │
└─────────────────────────────────────────────────────┘
```

**Por que não schema-per-tenant?**
- O `wisa-crm-service` pode ter centenas ou milhares de tenants
- Schema-per-tenant cria overhead de manutenção de migrations (cada migration precisa rodar N vezes)
- Connection pooling é ineficiente com schema-per-tenant (cada schema requer `SET search_path`)
- Row-level isolation com RLS é suficiente e mais elegante para este caso

### Índices críticos

```sql
-- Consulta mais frequente: buscar usuário por email dentro de um tenant
CREATE INDEX CONCURRENTLY idx_users_tenant_email 
  ON users (tenant_id, email);

-- Verificação de assinatura ativa (executada em cada login)
CREATE INDEX CONCURRENTLY idx_subscriptions_tenant_status 
  ON subscriptions (tenant_id, status) 
  WHERE status = 'active';

-- Invalidação de refresh tokens (executada em logout e renovação)
CREATE INDEX CONCURRENTLY idx_refresh_tokens_token_hash 
  ON refresh_tokens (token_hash) 
  WHERE revoked_at IS NULL;
```

---

## Consequências

### Positivas
- ACID garante consistência em operações de auth + billing
- RLS fornece isolamento entre tenants como segurança de profundidade
- MVCC permite alta concorrência de leituras sem bloqueios
- Extensões cobrem necessidades presentes e futuras
- Maturidade e confiabilidade comprovadas em produção

### Negativas
- PostgreSQL tem limite físico de conexões simultâneas (~100–200 por padrão) — requer PgBouncer
- Operações em instância única têm limite de escalabilidade horizontal (mitigado com read replicas)
- Manutenção de migrations e `VACUUM` requer atenção operacional
- Configuração segura (`pg_hba.conf`, SSL, `postgresql.conf`) requer conhecimento específico

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| Connection pool exhaustion em pico de logins | Alta | Alto | Alta |
| SQL Injection via queries não parametrizadas | Baixa | Crítico | Alta |
| Exposição de dados de outro tenant por bug de aplicação | Baixa | Crítico | Alta |
| Corrupção de dados por falha de hardware sem WAL | Baixa | Crítico | Alta |
| Crescimento descontrolado do audit_log bloqueando tabela | Média | Médio | Média |
| Conexão não-SSL interceptada em rede | Baixa | Alto | Média |

---

## Mitigações

### Connection pool exhaustion
- Instalar e configurar **PgBouncer** em modo `transaction pooling` entre o backend Go e o PostgreSQL
- Configurar `max_connections` no PostgreSQL para 200, com PgBouncer gerenciando pool de 10–20 conexões reais para centenas de conexões virtuais
- Monitorar via `pg_stat_activity` com alert quando conexões ativas > 80% do limite

### SQL Injection
- **Somente queries parametrizadas** — enforçado pelo GORM por padrão
- Configurar usuário de banco com privilégios mínimos (`SELECT`, `INSERT`, `UPDATE` nas tabelas necessárias — nunca `DROP`, `CREATE`, `TRUNCATE`)
- Revisar qualquer uso de `db.Raw()` ou `db.Exec()` em code review

### Isolamento multi-tenant
- Habilitar Row-Level Security em todas as tabelas com `tenant_id`
- Testes de integração explícitos que tentam acessar dados de outro tenant e verificam bloqueio
- Audit log de todas as queries cross-tenant via `pgaudit`

### Durabilidade
- Configurar `fsync = on` e `synchronous_commit = on` no `postgresql.conf` (padrão, não alterar)
- Configurar backup automático com `pg_dump` ou `pg_basebackup` + WAL archiving
- Testar restore a partir do backup mensalmente

### Crescimento do audit_log
- Implementar particionamento por range de data: `PARTITION BY RANGE (created_at)`
- Configurar rotação automática de partições antigas (ex: manter 12 meses, arquivar o restante)
- Criar índices na partição corrente, não na tabela pai

### Conexões seguras
- Configurar `ssl = on` no `postgresql.conf`
- No `pg_hba.conf`, exigir `hostssl` para todas as conexões remotas
- Gerar certificados TLS para o PostgreSQL e rotacioná-los anualmente

---

## Configuração de Produção Recomendada

### `postgresql.conf` (parâmetros críticos)

```ini
# Memória (ajustar para VPS com 4GB RAM como exemplo)
shared_buffers = 1GB                    # 25% da RAM
effective_cache_size = 3GB              # 75% da RAM
work_mem = 16MB                         # por operação de sort/hash
maintenance_work_mem = 256MB            # para VACUUM e índices

# Conexões
max_connections = 200                   # gerenciado via PgBouncer
superuser_reserved_connections = 3

# WAL e durabilidade
wal_level = replica                     # habilita streaming replication futura
fsync = on                              # não alterar
synchronous_commit = on                 # não alterar
checkpoint_completion_target = 0.9

# Logging
log_min_duration_statement = 1000       # loga queries > 1s
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on

# SSL
ssl = on
ssl_cert_file = 'server.crt'
ssl_key_file = 'server.key'
```

---

## Alternativas Consideradas

### MySQL/MariaDB
- **Prós:** Amplamente usado, fácil de operar, boa performance de leitura
- **Contras:** Row-Level Security ausente (crítico para multi-tenant), suporte a transações historicamente inferior ao PostgreSQL, sem `SKIP LOCKED` nativo (versões antigas), menor expressividade de tipos e constraints

### SQLite
- **Prós:** Zero configuração, sem servidor separado
- **Contras:** Sem suporte a concorrência real para escrita, inadequado para multi-tenant com múltiplos usuários simultâneos, sem Row-Level Security, sem suporte a conexões remotas seguras

### MongoDB
- **Prós:** Flexibilidade de schema
- **Contras:** Sem ACID em transações multi-documento até versão 4+, sem joins nativos eficientes, modelo eventual consistency inadequado para dados financeiros e de auth, sem Row-Level Security

### CockroachDB / Distributed SQL
- **Prós:** Escala horizontal nativa
- **Contras:** Overhead de latência em ambiente single-node, complexidade operacional desnecessária para o estágio atual, custo de licença nas versões avançadas

**PostgreSQL é a única escolha que combina ACID, Row-Level Security, maturidade, ecossistema e compatibilidade com GORM de forma completa.**

---

## Referências

- [PostgreSQL Row Security Policies](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [PgBouncer Documentation](https://www.pgbouncer.org/config.html)
- [PostgreSQL Security Hardening](https://www.postgresql.org/docs/current/auth-pg-hba-conf.html)
- [pgaudit Extension](https://github.com/pgaudit/pgaudit)
- [OWASP Database Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Database_Security_Cheat_Sheet.html)
