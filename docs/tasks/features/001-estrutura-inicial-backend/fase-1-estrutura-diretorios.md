# Fase 1 — Estrutura de Diretórios do Backend

## Objetivo

Criar a estrutura de diretórios do backend dentro da pasta `backend/`, em conformidade com Clean Architecture (ADR-005), code guidelines e ADRs do projeto.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O documento `docs/code_guidelines/backend.md` e o ADR-005 definem uma estrutura específica. A estrutura original está descrita na raiz do `wisa-crm-service/`, mas o requisito solicita que o backend fique dentro de `backend/`. Portanto, deve-se replicar a estrutura recomendada, prefixando com `backend/`.

### Ação 1.1

Consultar a estrutura exata definida nas guidelines:

- `cmd/api/main.go` — Entrypoint
- `internal/domain/entity/` — Entidades
- `internal/domain/repository/` — Interfaces de repositório
- `internal/usecase/` — Casos de uso
- `internal/infrastructure/persistence/` — Implementações GORM
- `internal/infrastructure/crypto/` — Serviços criptográficos
- `internal/infrastructure/http/middleware/` — Middlewares HTTP
- `internal/delivery/http/handler/` — Handlers HTTP
- `internal/delivery/http/dto/` — DTOs
- `internal/app/` — Wiring de dependências (opcional)
- `pkg/logger/` — Utilitários compartilhados
- `migrations/` — Migrations SQL versionadas

### Observação 1.1

A estrutura segue a Dependency Rule: domínio no centro, dependências apontando para dentro. Nenhum diretório de infraestrutura é importado pelo domínio.

---

### Pensamento 2

Na fase inicial, não haverá entidades, repositórios ou use cases implementados. A tarefa é criar apenas a **estrutura de pastas vazia** (com arquivos placeholder se necessário para que o Go reconheça os pacotes, ou simplesmente diretórios vazios). Em Go, um diretório sem arquivos `.go` não é um pacote válido. Para manter a estrutura navegável, podemos criar arquivos `.gitkeep` ou arquivos `.go` com apenas `package X` e um comentário.

### Ação 1.2

Definir o conjunto mínimo de diretórios para a fase inicial:

```
backend/
├── cmd/
│   └── api/
│       └── main.go          # Entrypoint — será populado em fases posteriores
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   ├── repository/
│   │   └── errors.go       # Será criado em fase posterior
│   ├── usecase/
│   ├── infrastructure/
│   │   ├── persistence/
│   │   ├── crypto/
│   │   └── http/
│   │       └── middleware/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       └── dto/
│   └── app/
├── pkg/
│   └── logger/
└── migrations/
```

### Observação 1.2

O `migrations/` está na raiz do backend conforme guidelines. O `internal/app/wire.go` pode ficar vazio ou ser criado apenas quando o `main.go` crescer (ADR-005).

---

### Pensamento 3

Em Go, um diretório deve conter pelo menos um arquivo `.go` para ser reconhecido como pacote. Diretórios vazios podem existir, mas ao rodar `go build ./...` pode haver avisos. A abordagem pragmática para a fase inicial é criar apenas os diretórios que terão código imediatamente: `cmd/api/` com `main.go` básico (apenas declaração de pacote e função main vazia ou com `log.Print` mínimo). Os demais diretórios podem ter um arquivo `doc.go` ou ficar vazios até as próximas fases.

### Ação 1.3

Plano de criação:

1. Criar `backend/` na raiz do projeto
2. Criar cada subdiretório conforme árvore acima
3. Criar `backend/cmd/api/main.go` com conteúdo mínimo para validar que a estrutura compila (ex: `package main` + `func main() {}`)
4. Opcional: criar `backend/go.mod` com `go mod init` (ou fazer na Fase 2)

### Observação 1.3

A Fase 2 tratará do `go.mod` e imports. Na Fase 1, o foco é a criação das pastas. Se `main.go` for criado, ele precisará de um `go.mod` para compilar. Portanto, a ordem ideal é: criar pastas primeiro, e na Fase 2 incluir `go mod init` e o `main.go` básico. **Ajuste:** na Fase 1, criar apenas os diretórios e um `main.go` placeholder que será completado na Fase 2. Ou: Fase 1 cria diretórios + `go mod init` mínimo para que a estrutura seja validável. O requisito diz "Não implemente nenhum código" — então a Fase 1 deve criar apenas diretórios, sem arquivos de código. Os diretórios vazios são suficientes; o `main.go` será criado na Fase 6 (endpoint health) ou em fase de bibliotecas.

### Decisão final Fase 1

**Criar exclusivamente a estrutura de diretórios.** Não criar arquivos `.go` nesta fase. O `main.go` e o `go.mod` serão criados nas fases subsequentes. Para que pastas vazias sejam versionadas no Git, criar arquivos `.gitkeep` em cada diretório vazio (convenção comum).

---

### Checklist de Implementação

1. [ ] Criar `backend/` na raiz do projeto
2. [ ] Criar `backend/cmd/api/`
3. [ ] Criar `backend/internal/domain/entity/`
4. [ ] Criar `backend/internal/domain/repository/`
5. [ ] Criar `backend/internal/usecase/`
6. [ ] Criar `backend/internal/infrastructure/persistence/`
7. [ ] Criar `backend/internal/infrastructure/crypto/`
8. [ ] Criar `backend/internal/infrastructure/http/middleware/`
9. [ ] Criar `backend/internal/delivery/http/handler/`
10. [ ] Criar `backend/internal/delivery/http/dto/`
11. [ ] Criar `backend/internal/app/`
12. [ ] Criar `backend/pkg/logger/`
13. [ ] Criar `backend/migrations/`
14. [ ] Adicionar `.gitkeep` em cada diretório que ficará vazio inicialmente (opcional, para versionamento)

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-005 Clean Architecture | Estrutura de camadas (domain, usecase, infrastructure, delivery) |
| Code Guidelines backend.md | Layout recomendado seguido |
| Dependency Rule | Domínio isolado; sem imports de infra no domain |
| Segurança | Nenhum arquivo sensível; estrutura não expõe dados |

---

## Referências

- [docs/code_guidelines/backend.md](../../../code_guidelines/backend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../../adrs/ADR-005-clean-architecture.md)
