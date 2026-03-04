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
│   └── 002-configuracao-banco-dados/
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
| [002-configuracao-banco-dados](features/002-configuracao-banco-dados/TRACKER.md) | Configuração do banco: estrutura base, env, containers, documentação, ORM/migrations | 0/5 | Pendente |

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

---

## Como Usar Este Tracker

1. **Ao iniciar uma sessão:** Leia este arquivo para saber o status atual
2. **Para detalhes de uma feature:** Consulte o `TRACKER.md` dentro da pasta da feature e os arquivos de tasks
3. **Durante o trabalho:** Atualize o checkbox da task (`[ ]` → `[~]` → `[x]`) no arquivo correspondente
4. **Ao finalizar a sessão:** Atualize a tabela "Status Geral" e adicione uma entrada no "Diário de Sessões"
5. **Para criar nova feature ou fix:** Consulte o [README.md](README.md) com o guia de criação