# Task Tracker

> **Este arquivo é o ponto central de controle do projeto.**  
> Consulte-o ao início de cada sessão de desenvolvimento para saber exatamente onde parou.

---

## Estrutura de Diretórios

O diretório `docs/tasks/` está organizado por **features** e **fixes**:

```
docs/tasks/
├── TRACKER.md              ← Você está aqui (visão geral)
├── README.md               ← Guia para criar novas features/fixes
├── features/               ← Features em desenvolvimento
│   ├── 001-estrutura-inicial-backend/
│   ├── 002-configuracao-banco-dados/
│   ├── 003-estrutura-tabelas-banco-dados/
│   └── 004-package-erro-padronizado/
└── fixes/                  ← Correções e bugs
    └── (vazio — para futuras correções)
```

---

## Legenda

- `[ ]` Pendente
- `[~]` Em andamento
- `[x]` Concluída
- `[-]` Cancelada

---

## Status Geral

| Feature/Fix | Descrição | Progresso | Status |
|-------------|-----------|-----------|--------|
| [001-estrutura-inicial-backend](features/001-estrutura-inicial-backend/TRACKER.md) | Estrutura inicial do backend: diretórios, libs, .gitignore, env, Dockerfile, health | 6/6 | Concluída |
| [002-configuracao-banco-dados](features/002-configuracao-banco-dados/TRACKER.md) | Configuração do banco: estrutura base, env, containers, documentação, ORM/migrations | 5/5 | Concluída |
| [003-estrutura-tabelas-banco-dados](features/003-estrutura-tabelas-banco-dados/TRACKER.md) | Estrutura de tabelas: schema wisa_crm_db, tenants, products, subscriptions, payments, users, user_product_access, refresh_tokens, audit_logs | 6/6 | Concluída |
| [004-package-erro-padronizado](features/004-package-erro-padronizado/TRACKER.md) | Package de erro padronizado: estrutura pkg/errors, AppError, catálogo de códigos, ErrorMapper, integração na delivery | 3/3 | Concluída |

---

## Diário de Sessões

*(Registre aqui as atividades significativas de cada sessão de desenvolvimento)*

### Sessão 1 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 001 — Estrutura inicial do backend
  - Criação de documentos de planejamento para as 6 fases
- **Features/fixes criados:** 001-estrutura-inicial-backend (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-estrutura-diretorios.md](features/001-estrutura-inicial-backend/fase-1-estrutura-diretorios.md)

### Sessão 2 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 001 — Estrutura inicial do backend
  - Fase 1: Estrutura de diretórios em `backend/` (Clean Architecture)
  - Fase 2: go.mod com Gin v1.9.1 e godotenv v1.5.1
  - Fase 3: .gitignore na raiz e em backend/
  - Fase 4: .env.example com PORT e APP_ENV
  - Fase 5: Dockerfile multi-stage (golang:1.25-alpine, alpine:3.19)
  - Fase 6: Endpoint GET /health com handler Gin
- **Features/fixes concluídos:** 001-estrutura-inicial-backend
- **Tasks concluídas:** 6/6 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

### Sessão 3 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 002 — Configuração do banco de dados
  - Criação de documentos de planejamento para as 5 fases
- **Features/fixes criados:** 002-configuracao-banco-dados (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-estrutura-base-banco.md](features/002-configuracao-banco-dados/fase-1-estrutura-base-banco.md)

### Sessão 4 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 002 — Configuração do banco de dados
  - Fase 1: persistence/database.go com GORM, pool configurado; main.go com DATABASE_URL
  - Fase 2: .env.example com DATABASE_URL documentada
  - Fase 3: backend/docker/database/Dockerfile, backend/docker/backend/Dockerfile, docker-compose.yml (postgres + backend)
  - Fase 5: migrations 000001_init_schema, cmd/migrate, Makefile
  - Fase 4: docs/backend/README.md, docs/backend/vps-configurations.md
- **Features/fixes concluídos:** 002-configuracao-banco-dados
- **Tasks concluídas:** 5/5 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

### Sessão 5 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 003 — Estrutura de tabelas do banco de dados
  - Criação de documentos de planejamento para as 6 fases (schema, tabelas, índices, RLS, triggers)
- **Features/fixes criados:** 003-estrutura-tabelas-banco-dados (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-schema-e-enums.md](features/003-estrutura-tabelas-banco-dados/fase-1-schema-e-enums.md)

### Sessão 6 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 003 — Estrutura de tabelas do banco de dados
  - Fase 1: Migration 000002 — schema wisa_crm_db e 8 tipos ENUM
  - Fase 2: Migration 000003 — tabelas tenants e products
  - Fase 3: Migration 000004 — tabelas subscriptions e payments
  - Fase 4: Migration 000005 — tabelas users e user_product_access
  - Fase 5: Migration 000006 — tabelas refresh_tokens e audit_logs (particionamento)
  - Fase 6: Migration 000007 — índices, RLS e triggers set_updated_at
  - Atualização de backend/.env.example e docs/backend/vps-configurations.md com search_path
- **Features/fixes concluídos:** 003-estrutura-tabelas-banco-dados
- **Tasks concluídas:** 6/6 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

### Sessão 7 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 004 — Package de Erro Padronizado
  - Criação de documentos de planejamento para as 3 fases (estrutura + AppError, catálogo + mapper, integração delivery)
- **Features/fixes criados:** 004-package-erro-padronizado (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-estrutura-diretorios-tipo-erro.md](features/004-package-erro-padronizado/fase-1-estrutura-diretorios-tipo-erro.md)

### Sessão 8 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 004 — Package de Erro Padronizado
  - Fase 1: pkg/errors com AppError, NewAppError, codes.go, MarshalJSON
  - Fase 2: Catálogo completo (INVALID_CREDENTIALS, ACCOUNT_LOCKED, etc.), domain/errors.go, MapToAppError em delivery/http/errors
  - Fase 3: RespondWithError, Recovery middleware em infrastructure/http/middleware, registro no router
  - Documentação em docs/backend/README.md
- **Features/fixes concluídos:** 004-package-erro-padronizado
- **Tasks concluídas:** 3/3 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

---

## Como Usar Este Tracker

1. **Ao iniciar uma sessão:** Leia este arquivo para saber o status atual
2. **Para detalhes de uma feature:** Consulte o `TRACKER.md` dentro da pasta da feature e os arquivos de tasks
3. **Durante o trabalho:** Atualize o checkbox da task (`[ ]` → `[~]` → `[x]`) no arquivo correspondente
4. **Ao finalizar a sessão:** Atualize a tabela "Status Geral" e adicione uma entrada no "Diário de Sessões"
5. **Para criar nova feature ou fix:** Consulte o [README.md](README.md) com o guia de criação