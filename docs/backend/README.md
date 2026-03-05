# Como Rodar o Backend

## Pré-requisitos

- Go 1.25 ou superior
- PostgreSQL 16+ (ou Docker)

## Desenvolvimento Local

### Opção 1: Com Docker (PostgreSQL em container)

1. Subir o banco: `docker compose up postgres -d`
2. Copiar variáveis: `cp backend/.env.example backend/.env`
3. Ajustar `DATABASE_URL` se necessário (padrão: `localhost:5432`)
4. Rodar migrations: `make migrate-up`
5. Na pasta backend: `go run ./cmd/api` ou `make run-backend` (a partir da raiz)

### Opção 2: Tudo em Docker

1. `docker compose up`
2. Backend disponível em http://localhost:8080
3. Health check: http://localhost:8080/health

### Opção 3: PostgreSQL local (sem Docker)

1. Instalar PostgreSQL 16
2. Criar banco `wisa_crm_db` e usuário `wisa_crm` com permissões
3. Configurar `backend/.env` com `DATABASE_URL`
4. Rodar migrations: `make migrate-up`
5. `go run ./cmd/api` (de dentro de `backend/`)

## Variáveis de Ambiente

Ver `backend/.env.example`.

| Variável                 | Obrigatória (produção) | Descrição                                                |
|--------------------------|------------------------|----------------------------------------------------------|
| `PORT`                   | Não                    | Porta HTTP (padrão: 8080)                               |
| `APP_ENV`                | Não                    | `development` ou `production`                           |
| `DATABASE_URL`           | Sim                    | Connection string PostgreSQL                            |
| `JWT_PRIVATE_KEY_PATH`   | Sim (com DB)           | Caminho para chave privada RSA 4096 bits (.pem)        |
| `JWT_ISSUER`             | Não                    | Claim `iss` do JWT (padrão: wisa-crm-service)             |
| `JWT_EXPIRATION_MINUTES` | Não                    | Duração do token em minutos (padrão: 15)                |
| `JWT_KEY_ID`             | Não                    | `kid` no header do JWT (padrão: key-2026-v1)             |
| `JWT_AUD_BASE_DOMAIN`    | Não                    | Domínio base para `aud` (ex: app.wisa-crm.com)          |

## Migrações

- **Aplicar todas:** `make migrate-up` (ou `cd backend && go run ./cmd/migrate`)
- **Rollback última:** `make migrate-down`
- **Rollback todas:** `make migrate-down-all` (cuidado em produção)

O comando de migrate usa `DATABASE_URL` do ambiente. Carregue `backend/.env` antes de executar (ex.: rodando de dentro de `backend/` ou com `.env` no diretório atual).

## Health Check

- `GET /health` retorna `{"status":"ok"}` com HTTP 200.

## Endpoint de Login

- **POST /api/v1/auth/login** — Autentica usuário e retorna JWT RS256.

### Pré-requisitos

- `DATABASE_URL` configurada
- `JWT_PRIVATE_KEY_PATH` apontando para arquivo `.pem` da chave privada RSA 4096 bits (gerar com: `openssl genrsa -out private.pem 4096`)
- Dados de teste no banco: tenant, product, user (com senha bcrypt), subscription ativa, user_product_access

### Request

```json
{
  "slug": "cliente1",
  "product_slug": "crm",
  "user_email": "usuario@empresa.com",
  "password": "senha123"
}
```

### Resposta de sucesso (200)

```json
{
  "token": "eyJhbGciOiJSUzI1NiIs..."
}
```

### Teste manual (curl)

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"slug":"cliente1","product_slug":"crm","user_email":"usuario@empresa.com","password":"senha123"}'
```

O token pode ser decodificado em [jwt.io](https://jwt.io) para validar os claims (iss, sub, aud, tenant_id, user_access_profile, etc.).

## Como Testar na sua Máquina

### Pré-requisitos

- **Docker e Docker Compose** — para Postgres e backend
- Ou **Go 1.25+** e **PostgreSQL 16+** — para rodar localmente sem Docker

### Opção A: Tudo em Docker (mais simples)

1. **Gerar chave privada RSA** (se ainda não existir):
   ```bash
   cd backend
   openssl genrsa -out private.pem 4096
   chmod 644 private.pem
   cd ..
   ```

2. **Subir os serviços:**
   ```bash
   docker compose up -d
   ```

3. **Rodar migrations** (se o banco estiver vazio):
   ```bash
   docker run --rm --network wisa_crm_service_default \
     -v "$(pwd)/backend/migrations:/migrations" \
     migrate/migrate \
     -path=/migrations \
     -database "postgres://wisa_crm:wisa_crm_secret@postgres:5432/wisa_crm_db?sslmode=disable" \
     up
   ```

4. **Popular banco com dados de teste:**
   ```bash
   docker exec -i wisa_crm_service-postgres-1 psql -U wisa_crm -d wisa_crm_db -f - < backend/scripts/seed_login_test.sql
   ```

5. **Testar login:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"slug":"cliente1","product_slug":"crm","user_email":"usuario@empresa.com","password":"senha123"}'
   ```

### Opção B: Backend local (Go) + Postgres em Docker

1. **Subir apenas o Postgres:**
   ```bash
   docker compose up postgres -d
   ```

2. **Rodar migrations** (comando do passo 3 da Opção A).

3. **Popular banco com dados de teste** (comando do passo 4 da Opção A).

4. **Configurar `.env` no backend:**
   ```bash
   cp backend/.env.example backend/.env
   ```
   Editar e ajustar:
   - `DATABASE_URL=postgres://wisa_crm:wisa_crm_secret@localhost:5432/wisa_crm_db?sslmode=disable&options=-c%20search_path=wisa_crm_db`
   - `JWT_PRIVATE_KEY_PATH=./private.pem`

5. **Gerar chave e rodar o backend:**
   ```bash
   cd backend
   openssl genrsa -out private.pem 4096
   go mod tidy
   go run ./cmd/api
   ```

6. **Testar login** (comando do passo 5 da Opção A).

### Dados de teste do seed

| Campo         | Valor                 |
|---------------|-----------------------|
| slug          | cliente1              |
| product_slug  | crm                   |
| user_email    | usuario@empresa.com   |
| password      | senha123              |

### Endpoints para validação

- **Health:** `GET http://localhost:8080/health`
- **Login:** `POST http://localhost:8080/api/v1/auth/login`

## Tratamento de Erros

O backend utiliza um package de erro padronizado (`pkg/errors`) com:

- **AppError**: struct com código, mensagem, detalhe e status HTTP
- **ErrorMapper**: mapeia erros de domínio para AppError (em `internal/delivery/http/errors`)
- **RespondWithError**: helper que handlers devem usar para responder erros — nunca `c.JSON` com `err.Error()`
- **Recovery middleware**: captura panics e responde com INTERNAL_ERROR (500) sem expor stack trace

Handlers devem usar `RespondWithError(c, err)` — importar com alias, ex.: `import httperrors "wisa-crm-service/backend/internal/delivery/http/errors"` e chamar `httperrors.RespondWithError(c, err)`.

## Produção (VPS)

Ver [docs/backend/vps-configurations.md](./vps-configurations.md).
