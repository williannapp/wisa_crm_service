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

| Variável     | Obrigatória (produção) | Descrição                          |
|--------------|------------------------|------------------------------------|
| `PORT`       | Não                    | Porta HTTP (padrão: 8080)          |
| `APP_ENV`    | Não                    | `development` ou `production`      |
| `DATABASE_URL` | Sim                  | Connection string PostgreSQL       |

## Migrações

- **Aplicar todas:** `make migrate-up` (ou `cd backend && go run ./cmd/migrate`)
- **Rollback última:** `make migrate-down`
- **Rollback todas:** `make migrate-down-all` (cuidado em produção)

O comando de migrate usa `DATABASE_URL` do ambiente. Carregue `backend/.env` antes de executar (ex.: rodando de dentro de `backend/` ou com `.env` no diretório atual).

## Health Check

- `GET /health` retorna `{"status":"ok"}` com HTTP 200.

## Produção (VPS)

Ver [docs/backend/vps-configurations.md](./vps-configurations.md).
