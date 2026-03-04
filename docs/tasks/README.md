# Rastreamento de Tasks

> Guia para a estrutura de diretórios e criação de novas features e fixes.

---

## Estrutura Atual

```
docs/tasks/
├── TRACKER.md                    # Ponto central — visão geral do projeto
├── README.md                     # Este arquivo — guia de uso
├── features/                     # Diretório de features
│   └── (vazio — para futuras features)
└── fixes/                        # Diretório de correções/bugs
    └── (vazio — para futuras correções)
```

---

## Como Criar uma Nova Feature

### 1. Criar o diretório

Use o padrão de nomenclatura: `XXX-nome-descritivo` (ex: `001-mvp-implementacao`, `002-relatorios-pdf`, `003-perfis-acesso`)

```
docs/tasks/features/001-nome-da-feature/
```

### 2. Criar os arquivos obrigatórios

| Arquivo | Descrição |
|---------|-----------|
| `TRACKER.md` | Tracker específico da feature com status, resumo das tasks e referências aos arquivos de tasks |
| `*.md` | Arquivos de tasks (ex: `fase-1-xxx.md`, `tasks-iniciais.md`, etc.) |

### 3. Estrutura mínima do TRACKER.md da feature

```markdown
# Feature XXX — Nome da Feature — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda
- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature
| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| ... | ... | 0/X | Pendente |

## Arquivos de Tasks
- [nome-do-arquivo.md](./nome-do-arquivo.md)

## Resumo das Tasks
- [ ] Task 1 — Descrição
- [ ] Task 2 — Descrição
```

### 4. Atualizar o TRACKER central

Adicione a nova feature em `docs/tasks/TRACKER.md` na tabela "Status Geral" e, se relevante, no "Diário de Sessões".

---

## Como Criar um Novo Fix (Correção/Bug)

### 1. Criar o diretório

Use o padrão: `XXX-descricao-curta` (ex: `001-bug-login-mobile`, `002-correcao-validacao-cpf`)

```
docs/tasks/fixes/001-descricao-do-fix/
```

### 2. Criar os arquivos obrigatórios

| Arquivo | Descrição |
|---------|-----------|
| `TRACKER.md` | Tracker do fix com problema, solução e checklist |
| `tasks.md` ou `checklist.md` | (Opcional) Detalhamento das tasks do fix |

### 3. Estrutura mínima do TRACKER.md do fix

```markdown
# Fix XXX — Descrição do Fix

## Problema
Descrição do bug ou correção necessária.

## Solução Proposta
Resumo da abordagem.

## Status
- [ ] Task 1 — Descrição
- [ ] Task 2 — Descrição

## Arquivos Afetados
- `caminho/arquivo.extensao`
- `caminho/componente.extensao`
```

### 4. Atualizar o TRACKER central

Adicione o fix em `docs/tasks/TRACKER.md` na tabela "Status Geral".

---

## Convenções

- **Numeração:** Use 3 dígitos (001, 002, 003...) para facilitar ordenação
- **Nomes:** Use kebab-case, descritivo e conciso
- **Legenda:** Mantenha consistência (`[ ]`, `[~]`, `[x]`, `[-]`)
- **Referências:** O TRACKER central deve sempre referenciar features/fixes ativos
- **Diário de Sessões:** Registre atividades significativas no TRACKER central

---

## Referências

- [TRACKER.md](./TRACKER.md) — Tracker central do projeto
- `docs/context.md` — (Opcional) Contexto e objetivos do projeto
- `docs/adrs/` — (Opcional) Decisões arquiteturais