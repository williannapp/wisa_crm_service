# Test App — Auth Code Flow Integration

Aplicação de teste (Angular + Go) para validar a integração completa com o fluxo de Authorization Code do **wisa-crm-service**.

## Restrições

- **Não alterar** código em `backend/` ou `frontend/` do projeto principal.
- Esta aplicação reside em pasta separada (`test-app/`).

## Estrutura

```
test-app/
├── backend/    # Go — API, auth flow, JWT validation
├── frontend/   # Angular — UI e chamadas à API
└── README.md
```

## Variáveis de Ambiente

Consulte `backend/.env.example`. Principais variáveis:

| Variável | Descrição | Default |
|----------|-----------|---------|
| `AUTH_SERVER_URL` | URL do wisa-crm-service | `https://auth.wisa.labs.com.br` |
| `TENANT_SLUG` | Slug do tenant de teste | `cliente1` |
| `PRODUCT_SLUG` | Slug do produto | `gestao-pocket` |
| `APP_URL` | URL base da test-app (para callback) | `http://localhost:8081` |
| `PORT` | Porta do backend | `8081` |

## Execução

### Desenvolvimento (separado)

1. **Backend:**
   ```bash
   cd test-app/backend
   cp .env.example .env   # ajuste se necessário
   # Se frontend em :4201, adicione: FRONTEND_URL=http://localhost:4201
   go run ./cmd/api
   ```
   Servidor em `http://localhost:8081`

2. **Frontend:**
   ```bash
   cd test-app/frontend
   npm install
   ng serve --port 4201
   ```
   Aplicação em `http://localhost:4201`. O proxy encaminha /api e /login ao backend.

### Docker Compose

**Apenas test-app:**
```bash
docker compose up -d test-app-backend test-app-frontend
```
- **test-app-backend:** `http://localhost:8081`
- **test-app-frontend:** `http://localhost:4201` (proxy para backend em /api, /login, /callback)

**Stack completa com NGINX (testes E2E com subdomínios):**
```bash
docker compose up -d
```
Adicione ao `/etc/hosts`: `127.0.0.1 auth.wisa.labs.com.br` e `127.0.0.1 lingerie-maria.wisa.labs.com.br`. Acesse `http://lingerie-maria.wisa.labs.com.br`. Ver [nginx/README.md](../nginx/README.md).

### Produção (backend serve frontend)

Após `ng build`, copie os arquivos de `dist/test-app-frontend/browser/` para um diretório estático servido pelo backend (ver backend).

## Fluxo de Autenticação

1. Usuário acessa a aplicação.
2. Se não autenticado → redirect para `GET /login` (backend) → redirect para auth server.
3. Usuário faz login no auth server.
4. Auth redirect para `GET /callback?code=...&state=...`.
5. Backend troca `code` por tokens e armazena em cookies HttpOnly.
6. Usuário é redirecionado para a aplicação com acesso liberado.

## Endpoints

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/health` | Status do servidor |
| GET | `/login` | Inicia fluxo de login (redirect ao auth) |
| GET | `/callback` ou `/:product/callback` | Recebe code e state do auth |
| GET | `/api/hello` | Rota protegida — requer JWT válido |
| POST | `/api/auth/refresh` | Renova access token (proxy ao auth server; lê refresh_token do cookie) |

## Referências

- [docs/integration/auth-code-flow-integration.md](../docs/integration/auth-code-flow-integration.md) — Guia de integração
- [docs/context.md](../docs/context.md) — Objetivo do sistema
