# Configurações VPS para o Backend

Este documento complementa as ADRs com passos operacionais para configurar o backend na VPS. Em produção, o PostgreSQL roda nativamente no servidor (não em container) — ver [ADR-009](../adrs/ADR-009-infraestrutura-vps-linux.md).

---

## 1. PostgreSQL na VPS

### Instalação

```bash
apt update && apt install -y postgresql-16
```

### Criação de usuário e banco

```bash
sudo -u postgres psql -c "CREATE USER wisa_crm WITH PASSWORD 'SECURE_PASSWORD_HERE';"
sudo -u postgres psql -c "CREATE DATABASE wisa_crm_db OWNER wisa_crm;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE wisa_crm_db TO wisa_crm;"
```

### pg_hba.conf

Restringir conexões ao loopback (backend e PostgreSQL na mesma máquina):

```
# /etc/postgresql/16/main/pg_hba.conf
# Conexões locais com senha
host    wisa_crm_db    wisa_crm    127.0.0.1/32    scram-sha-256
```

Após alteração: `systemctl reload postgresql`.

### SSL (ADR-003)

Em produção, habilitar SSL no PostgreSQL. Ver [ADR-003](../adrs/ADR-003-postgresql-como-banco-de-dados.md) para parâmetros em `postgresql.conf`.

---

## 2. Variáveis de ambiente do backend

O serviço systemd usa `EnvironmentFile=/etc/wisa-crm/env`:

```bash
# /etc/wisa-crm/env (exemplo - substituir credenciais)
PORT=8080
APP_ENV=production
DATABASE_URL=postgres://wisa_crm:SECURE_PASSWORD@localhost:5432/wisa_crm_db?sslmode=require
```

- Proteger o arquivo: `chmod 600 /etc/wisa-crm/env`
- Em produção usar `sslmode=require` ou `verify-full` (ADR-003)

---

## 3. Migrações em produção

**Ordem obrigatória:** rodar migrations **antes** de iniciar o novo binário.

```bash
# 1. Exportar DATABASE_URL (ou source de /etc/wisa-crm/env)
export DATABASE_URL="postgres://wisa_crm:PASSWORD@localhost:5432/wisa_crm_db?sslmode=require"

# 2. Aplicar migrations (a partir da raiz do projeto)
make migrate-up
# Ou: cd backend && go run ./cmd/migrate

# 3. Iniciar o serviço
systemctl start wisa-crm
```

### Rollback (quando necessário)

```bash
# Rollback da última migration
cd backend && go run ./cmd/migrate -down

# Rollback de todas (cuidado!)
cd backend && go run ./cmd/migrate -down-all
```

---

## 4. PgBouncer (produção)

Conforme [ADR-003](../adrs/ADR-003-postgresql-como-banco-de-dados.md), em ambientes com alta concorrência deve-se usar PgBouncer em modo `transaction pooling` entre o backend e o PostgreSQL. A `DATABASE_URL` deve apontar para o PgBouncer, não diretamente para o PostgreSQL.

---

## 5. Manutenção: partições do audit_logs

> **Lembrete operacional:** A tabela `audit_logs` usa particionamento por `created_at`. A partição inicial (conforme [fase-5-tabelas-refresh-tokens-audit-logs](../../tasks/features/003-estrutura-tabelas-banco-dados/fase-5-tabelas-refresh-tokens-audit-logs.md)) cobre o range **1970-01-01 até 2030-01-01**.
>
> **Antes de 2030**, criar novas partições para cobrir os períodos futuros. Caso contrário, INSERTs com `created_at` após o fim do range da última partição falharão.
>
> **Ação recomendada:** Agendar (cron/job) a criação de novas partições mensais ou anuais conforme política de retenção (ADR-003 sugere manter 12 meses de dados ativos). Ver a task fase-5 para detalhes do modelo de particionamento.

### Como criar uma nova partição

Conectar ao banco e executar (ajustar datas e nome da partição conforme o período desejado):

```bash
sudo -u postgres psql -d wisa_crm_db -c "
CREATE TABLE wisa_crm_db.audit_logs_2030 PARTITION OF wisa_crm_db.audit_logs
FOR VALUES FROM ('2030-01-01') TO ('2031-01-01');
"
```

**Exemplo — partição anual** (para ano 2030):

```sql
CREATE TABLE wisa_crm_db.audit_logs_2030 PARTITION OF wisa_crm_db.audit_logs
FOR VALUES FROM ('2030-01-01') TO ('2031-01-01');
```

**Exemplo — partição mensal** (para janeiro de 2030):

```sql
CREATE TABLE wisa_crm_db.audit_logs_2030_01 PARTITION OF wisa_crm_db.audit_logs
FOR VALUES FROM ('2030-01-01') TO ('2030-02-01');
```

**Regras:**
- Os ranges devem ser **contíguos e não sobrepostos** — o `TO` de uma partição é o `FROM` da próxima.
- Usar `TIMESTAMPTZ` no `created_at` implica datas em UTC; ajustar limites se usar timezone local.
- Partições antigas podem ser removidas com `DROP TABLE wisa_crm_db.audit_logs_YYYY;` após arquivamento (se aplicável).

**Verificar partições existentes:**

```sql
SELECT inhrelid::regclass AS partition_name
FROM pg_inherits
WHERE inhparent = 'wisa_crm_db.audit_logs'::regclass
ORDER BY partition_name;
```

---

## Referências

- [ADR-003 — PostgreSQL como Banco de Dados](../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [ADR-009 — Infraestrutura VPS Linux](../adrs/ADR-009-infraestrutura-vps-linux.md)
