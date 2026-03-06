# Como Rodar o Backend

## Pré-requisitos

- Go 1.25 ou superior
- PostgreSQL 16+ (ou Docker)

## Desenvolvimento Local

### Opção 1: Com Docker (PostgreSQL e Redis em container)

1. Subir o banco e Redis: `docker compose up postgres redis -d`
2. Copiar variáveis: `cp backend/.env.example backend/.env`
3. Ajustar `DATABASE_URL` e `REDIS_URL` se necessário (padrão: localhost)
4. Rodar migrations: `make migrate-up`
5. Na pasta backend: `go run ./cmd/api` ou `make run-backend` (a partir da raiz)

### Opção 2: Tudo em Docker

1. `docker compose up`
2. Backend disponível em http://localhost:8080
3. Frontend disponível em http://localhost:4200
4. Health check: http://localhost:8080/health

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
| `JWT_AUD_BASE_DOMAIN`    | Não                    | Domínio base para `aud` e redirect (ex: app.wisa-crm.com) |
| `REDIS_URL`              | Sim (com auth)         | URL Redis para authorization codes (ex: redis://localhost:6379/0) |

## Migrações

- **Aplicar todas:** `make migrate-up` (ou `cd backend && go run ./cmd/migrate`)
- **Rollback última:** `make migrate-down`
- **Rollback todas:** `make migrate-down-all` (cuidado em produção)

O comando de migrate usa `DATABASE_URL` do ambiente. Carregue `backend/.env` antes de executar (ex.: rodando de dentro de `backend/` ou com `.env` no diretório atual).

## Health Check

- `GET /health` retorna `{"status":"ok"}` com HTTP 200.

## Endpoints de Descoberta

- **GET /.well-known/jwks.json** — Retorna chaves públicas em formato JWKS (RFC 7517) para validação da assinatura RS256 dos JWTs. Não exige autenticação. Cache 24h (`Cache-Control: public, max-age=86400`). Ver [Obtenção da Chave Pública (JWKS)](../integration/auth-code-flow-integration.md#obtenção-da-chave-pública-jwks).

## Fluxo de Autenticação (Authorization Code)

O login utiliza o **Authorization Code Flow** (OAuth 2.0). O JWT não é retornado diretamente — o cliente recebe um code na URL de redirect e o troca pelo token via `POST /auth/token`.

### Endpoints

- **POST /api/v1/auth/login** — Autentica usuário e responde HTTP 302 redirect para callback do cliente com `?code=...&state=...`
- **POST /api/v1/auth/token** — Troca o authorization code por JWT (retorna `access_token`, `expires_in`, `refresh_token`, `refresh_expires_in`)
- **POST /api/v1/auth/refresh** — Renova sessão com refresh token (retorna novo par access + refresh)

### Pré-requisitos

- `DATABASE_URL` configurada
- `REDIS_URL` configurada (ex: `redis://localhost:6379/0`)
- `JWT_PRIVATE_KEY_PATH` apontando para arquivo `.pem` da chave privada RSA 4096 bits (gerar com: `openssl genrsa -out private.pem 4096`)
- Dados de teste no banco: tenant, product, user (com senha bcrypt), subscription ativa, user_product_access

### POST /api/v1/auth/login — Request

```json
{
  "slug": "cliente1",
  "product_slug": "crm",
  "user_email": "usuario@empresa.com",
  "password": "senha123",
  "state": "token_csrf_gerado_pelo_cliente"
}
```

### Resposta de sucesso (302)

`Location: https://{tenant_slug}.{JWT_AUD_BASE_DOMAIN}/{product_slug}/callback?code={code}&state={state}`

O code expira em **40 segundos** e é de uso único.

### POST /api/v1/auth/token — Request

```json
{
  "code": "code_recebido_na_url_do_callback"
}
```

### Resposta de sucesso (200)

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_in": 900,
  "refresh_token": "abc123...",
  "refresh_expires_in": 604800
}
```

### POST /api/v1/auth/refresh — Request

```json
{
  "refresh_token": "token_recebido_na_troca_de_code",
  "tenant_slug": "cliente1",
  "product_slug": "crm"
}
```

Resposta igual à do `/auth/token`. O refresh token antigo é invalidado (rotação).

### Teste manual (curl)

Para testar o fluxo completo localmente (redirect não será seguido automaticamente):

```bash
# Login (com -L para seguir redirect; em dev, verá a URL de callback com code na resposta)
curl -v -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"slug":"cliente1","product_slug":"crm","user_email":"usuario@empresa.com","password":"senha123","state":"xyz"}'

# Trocar code por token (use o code retornado na Location do redirect)
curl -X POST http://localhost:8080/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"code":"<code_da_url>"}'
```

O token pode ser decodificado em [jwt.io](https://jwt.io) para validar os claims (iss, sub, aud, tenant_id, user_access_profile, etc.).

### Documentação de integração para clientes

Ver [docs/integration/auth-code-flow-integration.md](../integration/auth-code-flow-integration.md).

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
- **JWKS:** `GET http://localhost:8080/.well-known/jwks.json`
- **Login:** `POST http://localhost:8080/api/v1/auth/login`
- **Token:** `POST http://localhost:8080/api/v1/auth/token`

## Tratamento de Erros

O backend utiliza um package de erro padronizado (`pkg/errors`) com:

- **AppError**: struct com código, mensagem, detalhe e status HTTP
- **ErrorMapper**: mapeia erros de domínio para AppError (em `internal/delivery/http/errors`)
- **RespondWithError**: helper que handlers devem usar para responder erros — nunca `c.JSON` com `err.Error()`
- **Recovery middleware**: captura panics e responde com INTERNAL_ERROR (500) sem expor stack trace

Handlers devem usar `RespondWithError(c, err)` — importar com alias, ex.: `import httperrors "wisa-crm-service/backend/internal/delivery/http/errors"` e chamar `httperrors.RespondWithError(c, err)`.

## Produção (VPS)

Ver [docs/backend/vps-configurations.md](./vps-configurations.md).
