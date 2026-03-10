# Fase 1 — Estrutura Base e Tela Hello World

## Objetivo

Criar a pasta `test-app/` na raiz do projeto com backend em Go e frontend em Angular, contendo uma tela simples "Hello World", sem alterar nenhum código nas pastas `backend/` ou `frontend/` existentes.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A aplicação de teste deve residir em uma pasta isolada. O requisito explícito é "Crie uma folder separada para essa aplicação de teste" e "Não altere nenhum código da folder de frontend ou backend". Portanto, a estrutura será:

```
wisa_crm_service/
├── backend/           ← NÃO ALTERAR
├── frontend/         ← NÃO ALTERAR
├── test-app/         ← NOVA
│   ├── backend/      ← Go
│   └── frontend/     ← Angular
└── ...
```

### Ação 1.1

Criar `test-app/backend/` com estrutura mínima de Clean Architecture (conforme code_guidelines/backend.md e ADR-005):
- `cmd/api/main.go` — entrypoint
- `internal/` — domain, delivery, infrastructure (estrutura básica)
- `go.mod` — dependências
- `.env.example` — variáveis (AUTH_SERVER_URL, PORT, etc.)
- Endpoint GET /health para validar que o servidor está rodando

### Observação 1.1

A aplicação de teste é um **cliente** do wisa-crm-service; não precisa de banco de dados, Redis ou migrations. A estrutura pode ser mais enxuta que o backend principal. Seguir convenções de nomenclatura (backend.md §10).

---

### Pensamento 2

O frontend Angular deve ser Angular v20+ (ADR-002, frontend.md). Uma tela "Hello World" pode ser um componente único que exibe texto estático. Não há necessidade de rotas complexas ou lazy loading nesta fase.

### Ação 1.2

Criar `test-app/frontend/` com `ng new` (ou equivalente). Estrutura mínima:
- Standalone components (padrão Angular v20)
- Um componente principal ou página que exiba "Hello World"
- `app.config.ts` com provideHttpClient se necessário
- Sem dependências ao backend da test-app nesta fase (apenas UI estática)

### Observação 1.2

O frontend da test-app é independente do frontend principal. Não compartilhar código entre eles. Usar Signals conforme guidelines (frontend.md §4) se houver estado; para "Hello World" puro, um template estático é suficiente.

---

### Pensamento 3

O backend da test-app precisa servir o frontend estático para que a aplicação funcione como um todo. Duas abordagens:
- (A) Backend serve os arquivos compilados do Angular (build output)
- (B) Em desenvolvimento, rodar frontend e backend separadamente (proxy ou CORS)

Para simplificar o plano, na Fase 1 o backend pode apenas expor GET /health e, opcionalmente, servir arquivos estáticos da pasta `dist/` ou `www/` após build do Angular. Ou manter separados em dev e documentar como executar.

### Ação 1.3

Definir que o backend da test-app:
- Exponha GET /health retornando `{"status":"ok"}`
- Em produção/deploy único: sirva os arquivos estáticos do Angular em `/` (após `ng build`)
- Documentar em README ou .env.example as portas (ex.: backend :8081, frontend :4201)

### Observação 1.3

A convenção do projeto principal usa backend na 8080 e frontend na 4200. A test-app deve usar portas diferentes para evitar conflito quando ambos estiverem rodando: ex. backend 8081, frontend 4201.

---

### Pensamento 4

Variáveis de ambiente necessárias para as fases futuras (redirect, callback):
- `AUTH_SERVER_URL` — `https://auth.wisa.labs.com.br` (fixo conforme requisito)
- `TENANT_SLUG` — slug do tenant de teste (ex.: `cliente1`)
- `PRODUCT_SLUG` — slug do produto (ex.: `gestao-pocket` ou valor de teste)
- `CALLBACK_URL` ou `APP_URL` — URL base da test-app para montar o callback (ex.: `http://localhost:8081` em dev)
- `PORT` — porta do backend

### Ação 1.4

Criar `test-app/backend/.env.example` com:
```
AUTH_SERVER_URL=https://auth.wisa.labs.com.br
TENANT_SLUG=cliente1
PRODUCT_SLUG=gestao-pocket
APP_URL=http://localhost:8081
PORT=8081
```

Incluir no .gitignore do test-app os arquivos .env (nunca commitar credenciais).

### Observação 1.4

Conformidade com docs/integration/auth-code-flow-integration.md pré-requisitos.

---

### Pensamento 5

Docker e docker-compose: o projeto principal tem docker-compose. A test-app pode ter seu próprio docker-compose ou ser documentada para rodar com `go run` e `ng serve`. Para manter o plano enxuto, Docker pode ser opcional ou fase posterior.

### Ação 1.5

Não incluir Docker na Fase 1. Documentar no README da test-app como rodar:
- `cd test-app/backend && go run ./cmd/api`
- `cd test-app/frontend && ng serve --port 4201`

### Observação 1.5

Se o projeto exigir Docker para a test-app, pode ser adicionado em fase posterior. O foco da Fase 1 é estrutura e Hello World.

---

## Decisão Final Fase 1

**Entregáveis:**

1. **test-app/backend/**
   - Estrutura: cmd/api/main.go, internal/ (delivery/http/handler, infrastructure se necessário)
   - Endpoint GET /health
   - go.mod com dependências mínimas (ex.: Gin ou net/http)
   - .env.example com AUTH_SERVER_URL, TENANT_SLUG, PRODUCT_SLUG, APP_URL, PORT
   - .gitignore

2. **test-app/frontend/**
   - Angular v20+ com `ng new`
   - Componente/página que exibe "Hello World"
   - Configuração para porta 4201 (ou documentar)

3. **test-app/README.md**
   - Como executar backend e frontend
   - Variáveis de ambiente
   - Restrição: não alterar backend/ e frontend/ do projeto principal

---

## Checklist de Implementação

1. [ ] Criar pasta test-app/
2. [ ] Criar test-app/backend com go.mod, main.go, GET /health
3. [ ] Criar test-app/backend/.env.example
4. [ ] Criar test-app/backend/.gitignore
5. [ ] Criar test-app/frontend com Angular (ng new)
6. [ ] Implementar tela "Hello World" no frontend
7. [ ] Criar test-app/README.md
8. [ ] Validar que backend e frontend rodam sem alterar código de backend/ ou frontend/

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Pasta separada | test-app/ fora de backend/ e frontend/ |
| Não alterar backend/frontend | Nenhuma modificação nesses diretórios |
| Code guidelines backend | Estrutura Clean Architecture enxuta |
| Code guidelines frontend | Angular v20+, Standalone, Signals se aplicável |
| ADR-005 | Clean Architecture |
| ADR-002 | Angular como framework |

---

## Referências

- [docs/code_guidelines/backend.md](../../../code_guidelines/backend.md)
- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../../adrs/ADR-005-clean-architecture.md)
