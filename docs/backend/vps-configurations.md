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

### Schema e search_path

As tabelas do sistema estão no schema `wisa_crm_db`. A aplicação deve usar `search_path=wisa_crm_db` na conexão ou executar `SET search_path TO wisa_crm_db` após conectar. Para operações multi-tenant, o middleware deve executar `SET LOCAL app.current_tenant_id = '<uuid>'` no início de cada request que acessa dados por tenant.

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

## Referências

- [ADR-003 — PostgreSQL como Banco de Dados](../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [ADR-009 — Infraestrutura VPS Linux](../adrs/ADR-009-infraestrutura-vps-linux.md)
