# Fase 2 — Importar Bibliotecas Necessárias

## Objetivo

Inicializar o módulo Go do backend e importar as bibliotecas necessárias para dar suporte à estrutura inicial, ao endpoint health e ao carregamento de variáveis de ambiente.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

Para a estrutura inicial (sem banco, sem JWT, sem autenticação), as bibliotecas necessárias são mínimas. O ADR-001 especifica Go 1.22+. As dependências devem ser enxutas, conforme princípio de supply chain (evitar dependências desnecessárias).

### Ação 2.1

Identificar dependências para esta fase:

| Biblioteca | Finalidade | Justificativa |
|------------|------------|---------------|
| `github.com/gin-gonic/gin` | Framework HTTP, roteamento, middlewares | Padrão adotado do projeto; alto desempenho, ampla adoção, API ergonômica |
| `github.com/joho/godotenv` | Carregar variáveis de `.env` | Padrão de mercado para Go; usada na Fase 4 |

### Observação 2.1

O **Gin** é definido como framework HTTP padrão do projeto. Oferece roteamento expressivo, suporte nativo a middlewares (rate limit, recovery, CORS), validação de requests (binding) e performance adequada para APIs REST. É um dos frameworks Go mais utilizados e mantidos, alinhado aos requisitos de APIs de autenticação com múltiplos endpoints.

### Decisão

Usar **`github.com/gin-gonic/gin`** como framework HTTP padrão e **`github.com/joho/godotenv`** para variáveis de ambiente. O Gin será utilizado para todos os endpoints, incluindo o health da Fase 6 e os demais endpoints futuros (login, refresh, JWKS, etc.).

---

### Pensamento 2

O `go.mod` deve ser criado em `backend/`. O module path deve seguir convenção: normalmente o repositório (ex: `github.com/wisa-labs/wisa-crm-service`) ou, para projeto local, um path como `wisa-crm-service/backend` ou `github.com/owner/repo/backend`.

### Ação 2.2

Executar na pasta `backend/`:
```bash
cd backend
go mod init <module-path>
```

O `<module-path>` deve ser definido conforme o repositório real. Exemplo: `github.com/wisa-labs/wisa-crm-service/backend` ou o path que o projeto utilizar.

### Observação 2.2

Verificar se existe um `go.mod` na raiz do projeto. O `docs/` e a estrutura sugerem um monorepo. Se o backend for submódulo, o path pode ser `wisa-crm-service/backend`. Manter consistência com a estrutura de imports que outros pacotes usarão.

---

### Pensamento 3

As bibliotecas `gin` e `godotenv` devem ser adicionadas com versão fixa. Gin está em v1.9.x; godotenv em v1.5.x.

### Ação 2.3

```bash
cd backend
go get github.com/gin-gonic/gin@v1.9.1
go get github.com/joho/godotenv@v1.5.1
go mod tidy
```

### Observação 2.3

`go mod tidy` remove dependências não utilizadas e atualiza `go.sum`. Garante reprodutibilidade do build. O Gin possui dependências transitivas (ex.: `go-playground/validator`); serão resolvidas automaticamente pelo `go get`.

---

### Pensamento 4

A versão mínima do Go no `go.mod` deve ser 1.25.2. Declarar no `go.mod` na linha de versão.

### Ação 2.4

No `go.mod`, garantir:
```
go 1.25.2
```

### Observação 2.4

O comando `go mod init` pode incluir uma versão diferente. Ajustar manualmente para `go 1.25.2` e validar que a máquina de desenvolvimento usa Go 1.25.2 ou superior.

---

### Checklist de Implementação

1. [ ] Navegar até `backend/`
2. [ ] Executar `go mod init <module-path>` (definir path conforme repositório)
3. [ ] Garantir `go 1.25.2` no `go.mod`
4. [ ] Executar `go get github.com/gin-gonic/gin@v1.9.1`
5. [ ] Executar `go get github.com/joho/godotenv@v1.5.1`
6. [ ] Executar `go mod tidy`
7. [ ] Validar que `go build ./...` executa sem erros (requer `main.go` da Fase 6 ou placeholder)
8. [ ] Documentar o module path no README da feature ou no TRACKER

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-001 Go 1.22+ | go 1.25.2 no go.mod (compatível) |
| Supply chain | Mínimo de dependências; godotenv amplamente utilizada |
| Segurança | go mod verify na CI (fase futura) |

---

## Observação sobre Ordem de Fases

Esta fase pode ser executada após a Fase 1 (estrutura) e antes ou em conjunto com a Fase 6 (endpoint health). O `main.go` criado na Fase 6 utilizará essas dependências.

---

## Referências

- [ADR-001 — Go como linguagem de backend](../../../adrs/ADR-001-golang-como-linguagem-de-backend.md)
- [Go Modules — golang.org](https://go.dev/ref/mod)
