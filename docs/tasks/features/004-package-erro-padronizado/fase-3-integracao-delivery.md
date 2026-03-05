# Fase 3 — Integração do ErrorMapper na Camada Delivery

## Objetivo

Integrar o ErrorMapper na camada de delivery HTTP, garantindo que todos os handlers e middlewares utilizem respostas de erro padronizadas via AppError, com conformidade Clean Architecture e consistência na API.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A camada delivery inclui handlers Gin e middlewares. Cada handler que chama use cases recebe `error` e deve:
1. Se err != nil, chamar MapToAppError(err) para obter *AppError
2. Responder com c.JSON(appErr.HTTPStatus, appErr) ou equivalente
3. Nunca fazer c.JSON(500, gin.H{"error": err.Error()}) — vazaria informações internas

Além disso, um middleware de **recovery** deve capturar panics e responder com AppError genérico (INTERNAL_ERROR, 500), evitando que o Gin exiba stack trace ao cliente.

### Ação 1.1

Definir integração:
1. **Helper de resposta:** Função `RespondWithError(c *gin.Context, err error)` que chama MapToAppError e c.JSON
2. **Middleware de Recovery:** Captura panic, loga internamente, responde 500 com AppError genérico
3. **Atualizar health_handler** (se houver erros) ou criar handler de exemplo que use RespondWithError

### Observação 1.1

O helper `RespondWithError` centraliza a lógica: um único ponto que aplica o mapper e serializa. Qualquer handler que retorne erro deve usar esse helper. O middleware de recovery evita vazamento em caso de panic inesperado.

---

### Pensamento 2

O helper RespondWithError precisa decidir o Content-Type e o formato do corpo. O AppError deve ser serializado para JSON. O gin.Context.JSON() usa application/json por padrão. O AppError precisa ter formato de resposta padronizado:

```json
{
  "status": "error",
  "error_code": "INVALID_CREDENTIALS",
  "message": "Credenciais inválidas.",
  "details": ""
}
```

O struct AppError da Fase 1 deve ter um tipo de "resposta" que inclua o campo "status" para serialização. Ou um wrapper:

```go
type ErrorResponse struct {
    Status   string `json:"status"`   // sempre "error"
    Code     string `json:"error_code"`
    Message  string `json:"message"`
    Details  string `json:"details,omitempty"`
}
```

O AppError pode ter método ToResponse() que retorna ErrorResponse, ou o MarshalJSON do AppError já inclui "status". Revisar Fase 1: o AppError deve serializar com "status":"error" no JSON.

### Ação 1.2

Garantir que a struct de resposta HTTP inclua `status: "error"` sempre. O AppError da Fase 1 pode ter:
- Campo `Status string` com valor fixo "error", ou
- MarshalJSON customizado que adiciona o campo

O helper RespondWithError fará:
```go
appErr := mapper.MapToAppError(err)
c.JSON(appErr.HTTPStatus, appErr)  // AppError já serializa corretamente
```

### Observação 1.2

Se AppError tiver os campos corretos com tags JSON (status, error_code, message, details), a serialização será automática. O campo Status pode ser definido no NewAppError como "error" constantemente.

---

### Pensamento 3

O middleware de recovery do Gin já existe (gin.Recovery()). Porém, ele pode retornar formato não padronizado. Devemos usar um custom recovery que:
1. Recupera do panic (recover())
2. Loga o panic (com stack trace) internamente
3. Responde com AppError INTERNAL_ERROR (500) sem expor o panic

O Gin permite `gin.CustomRecovery(handler)` ou similar. Verificar a API do Gin para custom recovery.

### Ação 1.3

Criar `internal/infrastructure/http/middleware/recovery.go` ou `internal/delivery/http/middleware/recovery.go`:
- Função que retorna gin.HandlerFunc
- No defer/recover, capturar panic, logar, escrever resposta JSON com AppError INTERNAL_ERROR
- Usar o mapper ou construir AppError diretamente (o panic não é um domain error)

Para panic, não há err para mapear — criar AppError INTERNAL_ERROR diretamente via pkg/errors.

### Observação 1.3

O middleware de recovery é infraestrutura (lida com panics, logging). O code guidelines indicam middlewares em `infrastructure/http/middleware/`. Porém, a estrutura atual do projeto pode ter `internal/delivery/http`. Verificar layout: feature 001 criou `internal/delivery/http/handler` e guidelines citam `infrastructure/http/middleware`. O ADR-005 mostra middlewares em infrastructure. Adotar `internal/infrastructure/http/middleware/recovery.go` para o custom recovery.

### Verificação da estrutura atual

O TRACKER e glob mostram: `internal/delivery/http/handler/`, `internal/infrastructure/persistence/`. As guidelines citam `infrastructure/http/middleware/`. A estrutura seria `internal/infrastructure/http/middleware/recovery.go`. Se não existir o diretório infrastructure/http, criá-lo. O main.go provavelmente usa Gin; o recovery é registrado no engine.

### Observação 1.4

A Fase 1 da feature 001 criou estrutura em delivery. O ADR-005 mostra tanto delivery quanto infrastructure. Para middlewares de infraestrutura (recovery, rate limit, tenant extraction), `infrastructure/http/middleware` é o local correto. Se o projeto ainda não tem esse diretório, a Fase 3 inclui criá-lo.

---

### Pensamento 4

O main.go atual (feature 001) configura o router Gin. A Fase 3 deve:
1. Registrar o custom recovery middleware como primeiro middleware (antes de qualquer handler)
2. Documentar que novos handlers devem usar RespondWithError ao invés de c.JSON manual para erros

O health_handler atual não retorna erros (sempre 200 OK). A integração é preparatória para handlers futuros (auth, etc.). Podemos adicionar um handler de exemplo que demonstra o uso, ou apenas documentar.

### Ação 1.4

- Registrar recovery middleware no main.go (ou no setup do router)
- Criar helper `RespondWithError` em `internal/delivery/http/errors/` ou junto ao mapper
- Documentar no código e na fase o padrão de uso

### Observação 1.4

A integração fica completa quando: (1) recovery está ativo, (2) helper existe e está documentado, (3) qualquer handler futuro usa o helper. Não é obrigatório modificar health_handler nesta fase, pois ele não retorna erros.

---

### Pensamento 5

Conformidade com ADRs e segurança:
- ADR-005: ErrorMapper centralizado — atendido na Fase 2; Fase 3 é a integração
- Code guidelines §12 (Credenciais): Nunca diferenciar mensagens de auth — o mapper garante
- Code guidelines §7 (Logs): Incluir request_id quando disponível; o recovery pode logar com contexto
- Evitar vazamento: Recovery nunca expõe stack trace ao cliente

### Ação 1.5

Checklist de conformidade:
- [ ] Recovery middleware nunca expõe panic ou stack ao cliente
- [ ] Helper RespondWithError é o único ponto de resposta de erro nos handlers
- [ ] Consistência: todos os erros retornam formato AppError padronizado

### Observação 1.5

A consistência nas respostas da API é atingida quando todos os pontos de saída de erro passam pelo mesmo fluxo: MapToAppError → RespondWithError → c.JSON.

---

### Decisão final Fase 3

**Implementar:**

1. Criar helper `RespondWithError(c *gin.Context, err error)` em `internal/delivery/http/errors/` (ou package apropriado)
2. Criar middleware de recovery customizado em `internal/infrastructure/http/middleware/recovery.go`
3. Registrar recovery no router (main.go ou onde o engine é configurado)
4. Documentar o padrão: handlers devem usar RespondWithError para erros; nunca c.JSON com err.Error()

---

### Checklist de Implementação

1. [ ] Criar `RespondWithError(c *gin.Context, err error)`
2. [ ] Criar middleware Recovery que responde com AppError INTERNAL_ERROR em caso de panic
3. [ ] Registrar Recovery no router
4. [ ] Garantir que Recovery logue panic internamente sem expor ao cliente
5. [ ] Atualizar ou criar documentação do padrão de tratamento de erros

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Consistência nas respostas | Todos os erros via RespondWithError |
| Impedir vazamento | Recovery com resposta genérica |
| Clean Architecture | Helper e recovery na camada correta |
| Code guidelines §6.3 | ErrorMapper centralizado + uso na delivery |
| Padrão para handlers | Documentação e helper obrigatório |

---

## Referências

- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../adrs/ADR-005-clean-architecture.md)
- [Gin Recovery](https://pkg.go.dev/github.com/gin-gonic/gin#Recovery)
