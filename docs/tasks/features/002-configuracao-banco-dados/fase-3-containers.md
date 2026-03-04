# Fase 3 — Containers (Dockerfile do Banco + docker-compose)

## Objetivo

Criar um `Dockerfile` para o banco de dados PostgreSQL e um `docker-compose.yml` na pasta raiz do projeto `wisa_crm_service`, permitindo subir o banco (e opcionalmente o backend) via containers para desenvolvimento local.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O requisito solicita "Dockerfile para o banco de dados". O PostgreSQL oficial já possui imagem Docker mantida. Duas abordagens:

1. **Usar imagem oficial diretamente no docker-compose** — sem Dockerfile customizado
2. **Criar Dockerfile que estende a imagem oficial** — para adicionar scripts de init, configurações customizadas (postgresql.conf), ou extensões

Para desenvolvimento local, a abordagem 1 é suficiente. O usuário explicitamente pediu "Dockerfile para o banco de dados". Portanto, criar um Dockerfile que:
- Use `FROM postgres:16-alpine` (ADR-003: PostgreSQL 16+)
- Possa copiar scripts de inicialização (ex: `init.sql` para criar extensões como `uuid-ossp`)
- Manter minimalista para dev; em produção na VPS o PostgreSQL roda nativo (não em container) conforme ADR-009

### Ação 3.1

Criar `docker/postgres/Dockerfile`:

```dockerfile
FROM postgres:16-alpine

# Opcional: copiar scripts de init ou config customizada
# COPY init.sql /docker-entrypoint-initdb.d/

# O entrypoint padrão da imagem postgres já executa scripts em /docker-entrypoint-initdb.d/
# Se houver init.sql no futuro, descomentar a linha acima
```

Ou: manter o Dockerfile mínimo, apenas `FROM postgres:16-alpine`, para que a imagem possa ser customizada futuramente. O docker-compose pode usar a imagem oficial diretamente ou buildar a partir deste Dockerfile.

### Observação 3.1

A pasta `docker/postgres/` centraliza artefatos relacionados ao PostgreSQL. Alternativa: `database/Dockerfile` ou `db/Dockerfile`. Seguir convenção comum: `docker/` para Dockerfiles de serviços auxiliares.

---

### Pensamento 2

O `docker-compose.yml` deve ficar na **raiz do projeto** `wisa_crm_service`, conforme requisito. Serviços típicos:
1. **postgres** — banco de dados
2. **backend** (opcional) — aplicação Go, usando o Dockerfile existente em `backend/`

O backend precisa de `DATABASE_URL` apontando para o host do PostgreSQL. Em Docker Compose, o hostname do serviço postgres é o nome do serviço (`postgres`). A URL seria `postgres://user:pass@postgres:5432/wisa_crm_db`.

### Ação 3.2

Estrutura do `docker-compose.yml` na raiz:

```yaml
services:
  postgres:
    build:
      context: ./docker/postgres
      dockerfile: Dockerfile
    # ou: image: postgres:16-alpine
    environment:
      POSTGRES_USER: wisa_crm
      POSTGRES_PASSWORD: wisa_crm_secret
      POSTGRES_DB: wisa_crm_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U wisa_crm -d wisa_crm_db"]
      interval: 5s
      timeout: 5s
      retries: 5

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      PORT: 8080
      APP_ENV: development
      DATABASE_URL: postgres://wisa_crm:wisa_crm_secret@postgres:5432/wisa_crm_db?sslmode=disable
    depends_on:
      postgres:
        condition: service_healthy

volumes:
  postgres_data:
```

### Observação 3.2

Credenciais no docker-compose são para **desenvolvimento local**. Em produção (VPS), o PostgreSQL roda nativo, não em container. O `depends_on` com `condition: service_healthy` garante que o backend só inicie após o PostgreSQL estar pronto. O volume `postgres_data` persiste dados entre reinícios.

---

### Pensamento 3

Sobre o Dockerfile do banco: se usarmos apenas `FROM postgres:16-alpine` sem customizações, o `docker-compose` pode usar `image: postgres:16-alpine` diretamente, sem build. O requisito pede "Dockerfile para o banco de dados". Para atender literalmente, criar um Dockerfile que estende a imagem e pode ser expandido no futuro (ex: adicionar extensão `uuid-ossp`, `pgcrypto`).

### Ação 3.3

Conteúdo final do `docker/postgres/Dockerfile`:

```dockerfile
FROM postgres:16-alpine

# PostgreSQL 16 Alpine - imagem oficial
# Para adicionar extensões ou scripts de init, descomentar:
# RUN apk add --no-cache postgresql-contrib
# COPY init.sql /docker-entrypoint-initdb.d/
```

O `postgres-contrib` inclui extensões adicionais. Para uuid-ossp, a extensão já pode estar disponível na imagem base. Verificar: `postgres:16-alpine` inclui suporte a `uuid-ossp` via pacote. Por simplicidade, manter o Dockerfile minimalista. Scripts de init para criar extensões podem ser adicionados em `docker/postgres/init/` na fase de ORM/migrations.

### Observação 3.3

A migration do GORM/golang-migrate criará as tabelas. O init do PostgreSQL pode apenas criar a extensão `uuid-ossp` se necessário. Adiar para Fase 5 se as migrations usarem `gen_random_uuid()` (nativo no PG 13+).

---

### Pensamento 4

O backend atual (Dockerfile em `backend/`) não passa variáveis de ambiente no build. O CMD executa o binário diretamente. Para que o backend em container use `DATABASE_URL`, ela deve ser passada no `environment` do docker-compose (já incluído na Ação 3.2).

O backend precisa ser capaz de rodar com a conexão apontando para `postgres:5432` quando em container. O hostname `postgres` resolve internamente na rede Docker.

### Ação 3.4

Garantir que o backend leia `DATABASE_URL` do ambiente (Fase 1 e 2). O docker-compose injeta a variável. Nenhuma alteração adicional no backend para esta fase.

### Observação 3.4

Consistência entre Fase 1, 2 e 3.

---

### Pensamento 5

Conformidade com ADR-009: em produção, PostgreSQL roda na VPS, não em container. O docker-compose é **apenas para desenvolvimento local**. Documentar isso claramente na Fase 4.

### Ação 3.5

Não expor o PostgreSQL para a internet no docker-compose. A porta 5432 mapeada para o host (`5432:5432`) é para acesso local (localhost) do desenvolvedor. Em produção, o backend e o PostgreSQL rodam no mesmo servidor (loopback).

### Observação 3.5

Segurança: desenvolvimento local usa localhost; produção usa VPS com PostgreSQL em loopback.

---

### Decisão final Fase 3

**Implementar:**
1. Criar `docker/postgres/Dockerfile` com `FROM postgres:16-alpine` (possibilidade de init scripts futuros)
2. Criar `docker-compose.yml` na raiz com serviços `postgres` e `backend`
3. Configurar healthcheck no postgres e `depends_on` no backend
4. Volume `postgres_data` para persistência
5. Credenciais alinhadas com `.env.example`: `wisa_crm`, `wisa_crm_secret`, `wisa_crm_db`
6. O serviço `backend` é opcional para quem quiser rodar tudo via compose; desenvolvedores podem rodar apenas `postgres` e o backend localmente com `go run`

---

### Checklist de Implementação

1. [ ] Criar diretório `docker/postgres/`
2. [ ] Criar `docker/postgres/Dockerfile` com `FROM postgres:16-alpine`
3. [ ] Criar `docker-compose.yml` na raiz do projeto
4. [ ] Configurar serviço `postgres` com env vars, healthcheck e volume
5. [ ] Configurar serviço `backend` (opcional) com dependência do postgres e DATABASE_URL
6. [ ] Validar: `docker compose up postgres -d` sobe o banco; `docker compose up` sobe postgres + backend
7. [ ] Documentar na Fase 4 o uso do docker-compose

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-003 | PostgreSQL 16 |
| ADR-009 | Docker para dev; produção usa PostgreSQL nativo na VPS |
| Segurança | Credenciais apenas para dev local; não para produção |
| Requisito | Dockerfile para DB + docker-compose na raiz |

---

## Referências

- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [docs/adrs/ADR-009-infraestrutura-vps-linux.md](../../../adrs/ADR-009-infraestrutura-vps-linux.md)
- [PostgreSQL Docker Official](https://hub.docker.com/_/postgres)
