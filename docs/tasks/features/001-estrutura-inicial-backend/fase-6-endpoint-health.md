# Fase 6 — Endpoint Health

## Objetivo

Criar um endpoint `GET /health` (ou `/healthz`) para validar se a aplicação está rodando corretamente. Esse endpoint é usado por load balancers, Kubernetes probes, monitoramento e scripts de deploy.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O endpoint health deve:
- Retornar HTTP 200 quando a aplicação estiver saudável
- Ser leve (sem consultas ao banco ou operações pesadas nesta fase)
- Não exigir autenticação
- Ter baixa latência

Para a estrutura inicial (sem banco), o health check é apenas "o processo está vivo e respondendo".

### Ação 6.1

Definir contrato do endpoint:
- **Método:** GET
- **Path:** `/health` (convenção comum) ou `/healthz` (padrão Kubernetes)
- **Resposta:** 200 OK com corpo JSON opcional, ex: `{"status":"ok"}`

### Observação 6.1

Em fases futuras, o health pode incluir verificações de conectividade com o banco ( readiness vs liveness ). Para agora, um 200 OK é suficiente.

---

### Pensamento 2

Segundo a Clean Architecture (ADR-005) e code guidelines, o handler HTTP pertence à camada `delivery/http/handler/`. O health é um handler simples que não depende de use cases ou repositórios.

### Ação 6.2

Criar o handler em `backend/internal/delivery/http/handler/health_handler.go`:
- Função que implementa `http.HandlerFunc` ou um handler com assinatura `func(w http.ResponseWriter, r *http.Request)`
- Registrar a rota `/health` no `main.go`

### Observação 6.2

O `main.go` em `cmd/api/` é o ponto de wiring. O handler de health será registrado no router **Gin** (framework HTTP padrão do projeto, conforme Fase 2).

---

### Pensamento 3

O código guidelines (backend.md) menciona `delivery/http/handler/` e `health_handler.go` explicitamente. A estrutura já está alinhada.

### Ação 6.3

Estrutura do handler com **Gin**:
```go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func HealthHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
```

### Observação 6.3

- O Gin registra apenas GET por padrão quando se usa `router.GET("/health", ...)`
- `c.JSON()` define automaticamente Content-Type `application/json`
- Resposta mínima e clara

---

### Pensamento 4

O `main.go` deve:
1. Carregar variáveis de ambiente (godotenv)
2. Obter PORT (default 8080)
3. Registrar handler em `/health`
4. Iniciar servidor HTTP
5. Opcional: graceful shutdown

### Ação 6.4

Estrutura do `main.go` com **Gin**:
```go
package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    "wisa-crm-service/backend/internal/delivery/http/handler"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found, using system environment")
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    router := gin.Default()
    router.GET("/health", handler.HealthHandler)

    addr := ":" + port
    log.Printf("Server starting on %s", addr)
    if err := router.Run(addr); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
```

### Observação 6.4

- `gin.Default()` inclui os middlewares `Logger` e `Recovery`
- O import path `wisa-crm-service/backend/...` deve corresponder ao `module` do `go.mod`

---

### Pensamento 5

Com Gin, o handler segue a assinatura `func(c *gin.Context)`. O registro é `router.GET("/health", handler.HealthHandler)` — passando a função diretamente, sem chamada. O Gin encaminha apenas requisições GET para essa rota.

### Ação 6.5

O handler `HealthHandler` recebe `*gin.Context` e utiliza `c.JSON()` para responder. Métodos diferentes de GET retornarão 404 automaticamente, pois a rota está registrada apenas para GET. Se for necessário 405 explícito, pode-se usar `router.Any("/health", ...)` com verificação de método — porém 404 é aceitável para health checks.

---

### Pensamento 6

Segurança: o endpoint `/health` é público. Não deve expor informações internas (versão do Go, paths, stack traces). O corpo `{"status":"ok"}` é seguro. Não incluir detalhes sensíveis.

### Ação 6.6

Manter resposta mínima. Em fases futuras, pode-se adicionar `version` ou `build_date` se forem não sensíveis. Evitar informações que auxiliem ataques.

### Observação 6.6

Conforme guidelines de segurança: não expor detalhes internos.

---

### Checklist de Implementação

1. [ ] Criar `backend/internal/delivery/http/handler/health_handler.go`
2. [ ] Implementar handler Gin que responde GET com 200 e `{"status":"ok"}`
3. [ ] Registrar rota `router.GET("/health", handler.HealthHandler)` (métodos não-GET retornam 404)
4. [ ] Criar ou atualizar `backend/cmd/api/main.go` com:
   - Carregamento de godotenv
   - Leitura de PORT (default 8080)
   - Inicialização do Gin (`gin.Default()`)
   - Registro da rota `GET /health`
   - Inicialização do servidor com `router.Run(addr)`
5. [ ] Ajustar imports conforme module path do go.mod
6. [ ] Validar com `go run backend/cmd/api/main.go` ou `cd backend && go run ./cmd/api`
7. [ ] Testar com `curl http://localhost:8080/health` — deve retornar 200 e JSON

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-005 Clean Architecture | Handler na camada delivery |
| Code Guidelines | Estrutura handler/ e convenções |
| Segurança | Resposta mínima, sem dados sensíveis |
| Dependency Rule | Handler não importa domain/usecase para health simples |

---

## Possíveis Extensões Futuras

- **Readiness:** Incluir verificação de conexão com PostgreSQL
- **Liveness:** Manter como está (processo vivo)
- **Version:** Adicionar campo `version` no JSON (via ldflags no build)

---

## Referências

- [Fase 2 — Importar bibliotecas (Gin como framework padrão)](./fase-2-importar-bibliotecas.md)
- [docs/code_guidelines/backend.md](../../../code_guidelines/backend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../../adrs/ADR-005-clean-architecture.md)
