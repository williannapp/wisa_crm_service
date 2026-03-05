# Fase 5 — Handler HTTP e Rota de Login

## Objetivo

Criar o handler HTTP para o endpoint de login, DTOs de request/response, rota e integração com o ErrorMapper existente. Garantir validação de input, tratamento de erros padronizado e resposta JSON conforme especificação.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O endpoint recebe POST com body JSON:
```json
{
  "slug": "cliente1",
  "product_slug": "crm",
  "user_email": "usuario@empresa.com",
  "password": "senha123"
}
```
Url de acesso: `https://slug.domain.com.br/<product_slug>`

Resposta de sucesso:
```json
{
  "token": "eyJhbGciOiJSUzI1NiIs..."
}
```

O code guidelines e a feature 004 já definem RespondWithError para mapear domain errors em AppError. O handler deve usar esse helper.

### Ação 1.1

Criar DTO de request em `delivery/http/dto/` ou no package do handler:
```go
type LoginRequest struct {
    Slug        string `json:"slug" binding:"required"`
    ProductSlug string `json:"product_slug" binding:"required"`
    UserEmail   string `json:"user_email" binding:"required,email"`
    Password    string `json:"password" binding:"required,min=1"`
}
```

O binding "required" e "email" do Gin devem validar. Para password, "required" e talvez min length (definido em guidelines ou ADR). A spec não define tamanho mínimo — usar "required" como mínimo. Considerar min=8 para segurança; se não houver regra, manter flexível.

### Observação 1.1

O Gin usa `ShouldBindJSON` que retorna erro se JSON inválido ou se required falhar. O handler deve tratar erro de binding e retornar 400 com AppError apropriado (ex: INVALID_REQUEST ou similar). Verificar se existe código para validation errors no mapper.

---

### Pensamento 2

Validação de binding: quando `c.ShouldBindJSON(&req)` falha, o erro não é um domain error. O ErrorMapper pode não ter tratamento para isso. Opções: (a) retornar 400 genérico com mensagem "Dados inválidos"; (b) criar um wrapper ou tipo específico para validation errors que mapeie para 400. O pkg/errors pode ter CodeInvalidRequest ou CodeValidationError. Verificar codes.go.

### Ação 1.2

Se não existir, adicionar em codes.go: `CodeInvalidRequest` ou `CodeValidationError` com HTTP 400. Criar função `NewInvalidRequest(message string)` ou similar. Para binding errors do Gin, o handler pode chamar `RespondWithError(c, ErrValidation)` passando um erro que o mapper trate, ou responder diretamente com AppError quando for erro de binding (não domain error). A abordagem mais limpa: o handler verifica `err := c.ShouldBindJSON`; se err != nil, chama `c.JSON(400, apperrors.NewInvalidRequest("Dados inválidos. Verifique os campos enviados."))` diretamente, sem passar pelo mapper. Ou criar `errors.ErrValidation` no domain e o use case não é chamado — o handler trata binding antes. Assim o mapper continua focado em domain errors. O handler faz:
```go
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, apperrors.NewInvalidRequest("Dados inválidos."))
    return
}
```

### Observação 1.2

Erros de binding são responsabilidade da delivery. Responder com 400 e AppError diretamente, sem passar pelo MapToAppError (que é para domain errors). O RespondWithError pode ser usado apenas para erros do use case. Para binding, resposta direta.

---

### Pensamento 3

A rota: conforme convenção REST e ADR-010, o login é parte do fluxo de auth. A rota pode ser `POST /api/v1/auth/login` ou `POST /api/v1/login`. O README e convenções do projeto podem definir. Adotar `POST /api/v1/auth/login` como padrão para auth.

### Ação 1.3

Registrar rota no router (main.go ou onde as rotas são configuradas):
```go
authGroup := router.Group("/api/v1/auth")
authGroup.POST("/login", authHandler.Login)
```

### Observação 1.3

O handler será AuthHandler com método Login. O AuthHandler recebe o use case AuthenticateUser via construtor.

---

### Pensamento 4

Estrutura do handler:
```go
type AuthHandler struct {
    authenticateUser *auth.AuthenticateUserUseCase
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req dto.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, apperrors.NewInvalidRequest("Dados inválidos. Verifique slug, product_slug, email e senha."))
        return
    }
    out, err := h.authenticateUser.Execute(c.Request.Context(), auth.LoginInput{
        Slug:        req.Slug,
        ProductSlug: req.ProductSlug,
        UserEmail:   req.UserEmail,
        Password:    req.Password,
    })
    if err != nil {
        deliveryerrors.RespondWithError(c, err)
        return
    }
    c.JSON(200, gin.H{"token": out.Token})
}
```

### Observação 1.4

O RespondWithError já existe em delivery/http/errors/respond.go. Usar esse helper. A resposta de sucesso pode ser um DTO estruturado: `LoginResponse{Token: out.Token}` com tag JSON, ou gin.H. Preferir struct para consistência.

---

### Pensamento 5

Resposta de sucesso: o corpo deve ser `{"token": "..."}`. O Content-Type será application/json. O handler usa `c.JSON(200, LoginResponse{Token: out.Token})`. Criar `LoginResponse` em dto:
```go
type LoginResponse struct {
    Token string `json:"token"`
}
```

### Ação 1.5

Criar `delivery/http/dto/login.go` com LoginRequest e LoginResponse.

### Observação 1.5

Mantém DTOs organizados. O handler importa o package dto.

---

### Pensamento 6

Conformidade com as respostas de erro da especificação: SUBSCRIPTION_SUSPENDED e SUBSCRIPTION_CANCELED têm formato exato. A Fase 1 já ajustou NewSubscriptionSuspended e NewSubscriptionCanceled com os details corretos. O ErrorMapper mapeia ErrSubscriptionSuspended e ErrSubscriptionCanceled. O AppError serializa com status, error_code, message, details. Verificar se o formato final é:
```json
{
  "status": "error",
  "error_code": "SUBSCRIPTION_SUSPENDED",
  "message": "Acesso suspenso por pendência financeira.",
  "details": "Sua assinatura não está ativa devido a pagamentos em aberto. Por favor, atualize sua forma de pagamento para acessar o software."
}
```

O AppError já tem esses campos. Garantir que NewSubscriptionSuspended inclua o details na Fase 1.

### Ação 1.6

Verificar na implementação da Fase 1 que os códigos SUBSCRIPTION_SUSPENDED e SUBSCRIPTION_CANCELED tenham os details exatos da especificação.

### Observação 1.6

Já documentado na Fase 1. O handler apenas chama RespondWithError — o mapper e os códigos fazem o resto.

---

### Decisão final Fase 5

**Implementar:**

1. Criar `delivery/http/dto/login.go` com LoginRequest (slug, user_email, password) e LoginResponse (token)
2. Criar `delivery/http/handler/auth_handler.go` com método Login
3. Handler: bind JSON, validar, chamar use case, RespondWithError em caso de erro, c.JSON(200, response) em sucesso
4. Registrar rota `POST /api/v1/auth/login`
5. Adicionar CodeInvalidRequest em pkg/errors se não existir, para erros de binding
6. Garantir que resposta de sucesso seja `{"token": "..."}`

---

### Checklist de Implementação

1. [ ] Criar DTOs LoginRequest e LoginResponse
2. [ ] Criar AuthHandler com método Login
3. [ ] Tratar erro de binding com 400
4. [ ] Usar RespondWithError para erros do use case
5. [ ] Registrar rota POST /api/v1/auth/login
6. [ ] Adicionar CodeInvalidRequest se necessário
7. [ ] Testar manualmente com curl ou similar

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Especificação | Request com slug, product_slug, user_email, password |
| Resposta sucesso | JSON com token |
| Resposta erro | Via AppError padronizado |
| Code guidelines §6.3 | RespondWithError para domain errors |
| Validação de input | Gin binding (required, email) |

---

## Referências

- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [fase-4-usecase-authenticate.md](./fase-4-usecase-authenticate.md)
- [Feature 004 - ErrorMapper](../../004-package-erro-padronizado/fase-3-integracao-delivery.md)
