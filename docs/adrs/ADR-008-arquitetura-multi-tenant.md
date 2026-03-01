# ADR-008 — Estratégia de Arquitetura Multi-Tenant

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (modelagem de dados e isolamento)

---

## Contexto

O `wisa-crm-service` é um sistema SaaS que serve múltiplos clientes (tenants) usando a mesma infraestrutura centralizada. Cada tenant é uma empresa ou cliente que contratou o serviço, e possui seus próprios usuários, planos e regras de assinatura.

O modelo de negócio implica que:

- **Dados de tenants diferentes nunca devem ser acessíveis um pelo outro** — vazamento de dados de usuários de um cliente para outro é uma falha catastrófica de negócio e compliance
- O número de tenants pode crescer de dezenas para centenas ao longo do tempo
- Cada tenant pode ter dezenas a milhares de usuários
- O sistema precisa identificar corretamente o tenant em cada requisição para aplicar as regras corretas de validação de assinatura

Existem três estratégias principais de multi-tenancy em bancos de dados:

1. **Database-per-tenant:** banco de dados separado para cada tenant
2. **Schema-per-tenant:** schema PostgreSQL separado para cada tenant, no mesmo banco
3. **Row-level tenancy:** tabelas compartilhadas com `tenant_id` em cada linha, mesmo schema

A escolha impacta diretamente: isolamento de dados, custo operacional, complexidade de migrations, escalabilidade e segurança.

---

## Decisão

**A estratégia adotada é Row-Level Tenancy: tabelas compartilhadas com `tenant_id` em todas as entidades, reforçado por Row-Level Security (RLS) no PostgreSQL.**

A identificação do tenant em cada requisição será feita via:
1. **Subdomínio HTTP** para o portal de login (ex: `cliente1.auth.wisa-crm.com`)
2. **Claim `tenant_id` no JWT** para endpoints autenticados
3. **Parâmetro de request** para o endpoint de login (campo `tenant_slug`)

---

## Justificativa

### 1. Por que não Database-per-tenant?

**Prós:** Máximo isolamento, possibilidade de diferentes versões de schema por tenant, backup e restore isolados.

**Contras para este contexto:**
- Centenas de tenants = centenas de conexões de banco distintas — impraticável para uma VPS single-server
- Migrations precisam ser executadas N vezes (uma por banco) — operação complexa propensa a inconsistências
- PostgreSQL tem custo de overhead por database (processo de background por banco)
- Connection pooling eficiente é inviável (PgBouncer não suporta bem múltiplos databases dinamicamente)

### 2. Por que não Schema-per-tenant?

**Prós:** Bom isolamento, migrations relativamente isoladas, `SET search_path` permite mesmas queries em schemas diferentes.

**Contras para este contexto:**
- `SET search_path = tenant_schema` precisa ser executado em cada conexão — comportamento não confiável com connection pooling (PgBouncer pode reusar conexão com search_path do tenant anterior)
- Migrations ainda precisam ser executadas por schema — mais manageable que por banco, mas ainda complexo
- PostgreSQL tem limite prático de ~10.000 schemas por banco, mas schemas muito numerosos impactam desempenho do catalog
- Não há benefício de performance sobre Row-Level com índices adequados

### 3. Por que Row-Level Tenancy?

**Prós:**
- Migrations executadas **uma única vez** — simplicidade operacional máxima
- Connection pooling funciona perfeitamente — todas as conexões usam o mesmo banco/schema
- Row-Level Security do PostgreSQL fornece isolamento enforçado pelo banco, não apenas pela aplicação
- Índices compostos `(tenant_id, ...)` garantem performance equivalente a schemas separados
- Escalabilidade horizontal futura (read replicas) funciona de forma transparente

**Contras mitigados:**
- Risco de vazamento de dados → mitigado por RLS (ADR-003) e Global Scopes no GORM (ADR-004)
- Backup granular por tenant não é trivial → mitigado com backup completo + restore seletivo via `COPY ... WHERE tenant_id = $1`

### 4. Identificação do tenant por requisição

O tenant deve ser identificado **antes** de qualquer operação de banco. As estratégias por contexto:

#### Contexto de login (usuário ainda não autenticado)

O tenant precisa ser identificado sem JWT. Estratégias possíveis:

**Opção A — Subdomínio:**
```
https://cliente1.auth.wisa-crm.com/login
→ NGINX extrai "cliente1" e passa via header X-Tenant-Slug
```

**Opção B — Campo no formulário de login:**
```json
{
  "tenant_slug": "cliente1",
  "email": "usuario@empresa.com",
  "password": "senha"
}
```

**Decisão:** Usar **campo `tenant_slug` no request de login** como método primário, com suporte a subdomínio como método alternativo. O campo no formulário é mais flexível (funciona em qualquer domínio), mais simples de implementar e permite que o mesmo endpoint de login sirva múltiplos tenants sem configuração DNS por tenant.

#### Contexto autenticado (JWT presente)

O `tenant_id` é lido diretamente do claim JWT validado:

```go
func ExtractTenantFromJWT(claims jwt.MapClaims) (uuid.UUID, error) {
    tenantID, ok := claims["tenant_id"].(string)
    if !ok {
        return uuid.Nil, ErrInvalidToken
    }
    return uuid.Parse(tenantID)
}
```

### 5. Modelagem de dados completa

```sql
-- Tenant: entidade raiz do sistema multi-tenant
CREATE TABLE tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(63) UNIQUE NOT NULL,  -- ex: "cliente1"
    name        VARCHAR(255) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, suspended, cancelled
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Assinatura: controla acesso ao sistema
CREATE TABLE subscriptions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    plan            VARCHAR(50) NOT NULL,  -- free, starter, pro, enterprise
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    expires_at      TIMESTAMPTZ NOT NULL,
    grace_period_end TIMESTAMPTZ,  -- período de tolerância após vencimento
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Usuários vinculados a um tenant
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id),
    email           VARCHAR(320) NOT NULL,
    password_hash   VARCHAR(72) NOT NULL,  -- bcrypt output max length
    full_name       VARCHAR(255),
    role            VARCHAR(50) NOT NULL DEFAULT 'user',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until    TIMESTAMPTZ,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, email)
);

-- Refresh tokens: controle de sessões ativas
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    token_hash  CHAR(64) NOT NULL UNIQUE,  -- SHA-256 do token
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address  INET,
    user_agent  TEXT
);

-- Audit log: rastreabilidade de eventos de segurança
CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id),
    user_id     UUID REFERENCES users(id),  -- nullable para tentativas de login com email inválido
    event_type  VARCHAR(50) NOT NULL,  -- login_success, login_failed, token_refreshed, logout, account_locked
    ip_address  INET,
    user_agent  TEXT,
    metadata    JSONB,  -- dados adicionais específicos do evento
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);  -- particionamento por data

-- Índices para performance multi-tenant
CREATE INDEX CONCURRENTLY idx_users_tenant_email    ON users (tenant_id, email);
CREATE INDEX CONCURRENTLY idx_users_tenant_status   ON users (tenant_id, status);
CREATE INDEX CONCURRENTLY idx_subscriptions_tenant  ON subscriptions (tenant_id, status, expires_at);
CREATE INDEX CONCURRENTLY idx_refresh_tokens_hash   ON refresh_tokens (token_hash) WHERE revoked_at IS NULL;
CREATE INDEX CONCURRENTLY idx_audit_tenant_date     ON audit_logs (tenant_id, created_at DESC);

-- Row-Level Security
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
```

### 6. Hierarquia de isolamento (defense in depth)

```
Camada 1: NGINX                → Identifica tenant pelo subdomínio/request (opcional)
Camada 2: Middleware Go        → Extrai e valida tenant_id do request/JWT
Camada 3: Use Case             → Passa tenant_id explicitamente para todos os repositórios
Camada 4: Repository (GORM)    → Global Scope injeta tenant_id em todas as queries
Camada 5: PostgreSQL RLS       → Políticas de banco enforçam tenant_id mesmo se camadas acima falharem
```

Cinco camadas de isolamento garantem que um bug em qualquer camada isolada não resulte em vazamento de dados entre tenants.

---

## Consequências

### Positivas
- Migrations únicas — operação simples e sem risco de inconsistência entre tenants
- Connection pooling funciona perfeitamente com PgBouncer
- Defense in depth: 5 camadas de isolamento de tenant
- Performance com índices compostos equivalente a schemas separados
- Escalabilidade para centenas de tenants sem mudança de arquitetura

### Negativas
- Backup granular por tenant requer SQL específico (`pg_dump` com cláusula WHERE)
- Uma query sem filtro de `tenant_id` (bug) pode retornar dados de múltiplos tenants — mitigado por RLS e Global Scopes
- Crescimento da tabela `users` com muitos tenants requer monitoramento de índices

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| Query sem filtro tenant_id retornando dados cross-tenant | Baixa | Crítico | Alta |
| Tenant spoofing: usuário manipula tenant_id no request de login | Média | Alto | Alta |
| Crescimento descontrolado de audit_logs por tenant abusivo | Média | Médio | Média |
| Tenant com assinatura vencida ainda acessa sistema via token válido | Alta | Médio | Média |
| Enumeração de tenants via timing attack no login | Média | Médio | Média |

---

## Mitigações

### Query sem filtro tenant_id
- Row-Level Security no PostgreSQL (última linha de defesa)
- Global Scope GORM em todos os models com tenant_id
- Testes de integração que verificam explicitamente o isolamento entre tenants
- Code review: qualquer query sem `tenant_id` em tabelas multi-tenant é bloqueante

### Tenant spoofing
- O `tenant_slug` recebido no login é validado contra o banco — só aceitar slugs existentes
- **Nunca** confiar em `tenant_id` recebido no body do request para endpoints autenticados — sempre usar o `tenant_id` do JWT validado
- Log e alerta para tentativas de login com `tenant_slug` inválido

### Crescimento de audit_logs
- Particionamento por range de data (já modelado acima)
- Política de retenção: partições com mais de 12 meses podem ser movidas para cold storage ou dropadas
- Monitorar tamanho das partições com alerta para crescimento anormal por tenant

### Token válido de tenant com assinatura vencida
- Aceitar a janela de 15 minutos como trade-off de design (access token de vida curta)
- O `wisa-crm-service` recusa emissão de novo token na renovação se a assinatura venceu
- Opcionalmente, manter cache Redis de tenants com assinatura vencida para validação no middleware (custo de complexidade vs benefício de bloqueio imediato)

### Timing attack na identificação de tenant
- Retornar sempre a mesma mensagem de erro para tenant inválido, email inválido e senha inválida: "Credenciais inválidas"
- Usar `time.Sleep` de duração constante em caso de tenant não encontrado para evitar diferença de tempo de resposta
- Aplicar bcrypt mesmo quando o usuário não é encontrado (compare com hash dummy) para normalizar o tempo de resposta

---

## Alternativas Consideradas

### Database-per-tenant
- **Rejeitada:** inviável para centenas de tenants em VPS single-server; migrations complexas; connection pooling ineficiente

### Schema-per-tenant
- **Rejeitada:** migrations ainda complexas; problema de search_path com connection pooling; sem benefício de performance sobre row-level com índices adequados

### Hybrid: schema-per-tenant para grandes clientes + row-level para pequenos
- **Rejeitada:** complexidade operacional excessiva; dois modelos de dados a manter; sem justificativa clara de negócio no estágio atual

**Row-Level Tenancy com RLS é a estratégia mais adequada para o estágio atual e previsão de crescimento do sistema.**

---

## Referências

- [Multi-tenant SaaS patterns in PostgreSQL](https://www.citusdata.com/blog/2016/10/03/designing-your-saas-database-for-high-scalability/)
- [PostgreSQL Row-Level Security](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [Citus: Multi-tenant Data Isolation](https://docs.citusdata.com/en/stable/use_cases/multi_tenant.html)
- [OWASP Insecure Direct Object Reference Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet.html)
