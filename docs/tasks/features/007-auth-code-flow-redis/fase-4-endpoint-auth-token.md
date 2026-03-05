# Fase 4 — Endpoint POST /api/v1/auth/token

## Objetivo

Implementar o endpoint `POST /api/v1/auth/token` que recebe o authorization code, valida existência e expiração, remove o code do Redis (single-use), gera o JWT e retorna `{ "access_token": "JWT", "expires_in": 900 }`. O code é trocado por token de forma atômica.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O cliente (aplicação que implementa GET /callback) recebe o redirect com `?code=XYZ&state=123`. O backend do cliente deve chamar o servidor de autenticação para trocar o code pelo token. O request pode ser:
- Body JSON: `{ "code": "XYZ" }`
- Ou form-urlencoded: `code=XYZ` (conforme exemplo do usuário)

O conteúdo deve ser `application/x-www-form-urlencoded` ou `application/json`. OAuth 2.0 usa form-urlencoded; JSON é mais comum em APIs REST modernas. Documentar suporte a ambos ou escolher um. Para simplicidade, aceitar JSON `{ "code": "..." }`.

### Ação 1.1

DTO de request:

```go
type TokenRequest struct {
    Code string `json:"code" binding:"required"`
}
```

### Observação 1.1

O binding `required` garante que code ausente retorne 400. O Gin valida automaticamente com `ShouldBindJSON`.

---

### Pensamento 2

O Use Case `ExchangeCodeForToken` (ou `ExchangeAuthCodeForToken`) deve:
1. Receber o code
2. Chamar `authCodeStore.GetAndDelete(ctx, code)`
3. Se erro (code não existe, expirou ou já usado) → retornar `ErrCodeInvalidOrExpired`
4. Se sucesso, converter AuthCodeData em JWTClaims
5. Chamar `jwtSvc.Sign(ctx, claims)` para gerar o JWT
6. Retornar o token e expires_in (900 = 15 minutos em segundos, conforme ADR-006)

### Ação 1.2

Criar Use Case em `internal/usecase/auth/exchange_code_for_token.go`:

```go
type ExchangeCodeForTokenUseCase struct {
    authCodeStore service.AuthCodeStore
    jwtSvc        service.JWTService
}

type ExchangeCodeInput struct {
    Code string
}

type ExchangeCodeOutput struct {
    AccessToken string
    ExpiresIn   int
}
```

Execute:
- GetAndDelete(code)
- Se nil data ou err → return ErrCodeInvalidOrExpired
- Build JWTClaims from AuthCodeData
- Sign with jwtSvc
- Return ExchangeCodeOutput{AccessToken: token, ExpiresIn: 900}

### Observação 1.2

O `expires_in` é fixo em 900 (15 min) conforme ADR-006. O JWT já contém `exp`; o cliente usa `expires_in` para saber quando renovar (refresh token) antes de fazer requisições.

---

### Pensamento 3

O JWTClaims precisa de Issuer (iss). O JWTService existente provavelmente preenche iss a partir de config. Verificar como o Sign() funciona — se recebe claims parciais e preenche iss, iat, exp, nbf, jti automaticamente. O AuthCodeData tem Subject, Audience, TenantID, UserAccessProfile. O Sign deve receber esses claims e preencher o restante.

### Ação 1.3

Construir JWTClaims a partir de AuthCodeData:
```go
claims := service.JWTClaims{
    Subject:           data.Subject,
    Audience:          data.Audience,
    TenantID:          data.TenantID,
    UserAccessProfile: data.UserAccessProfile,
    // iss, iat, exp, nbf, jti preenchidos pelo jwtSvc.Sign
}
```

### Observação 1.3

Se o JWTService.Sign exigir todos os campos, o AuthCodeData deve incluí-los ou o Sign deve ter overload que aceita claims parciais. Verificar implementação existente do RSAJWTService.

---

### Pensamento 4

O Handler para POST /api/v1/auth/token:
- Recebe TokenRequest
- Chama exchangeCodeForToken.Execute(ctx, ExchangeCodeInput{Code: req.Code})
- Se err == ErrCodeInvalidOrExpired → mapear para 401
- Se sucesso → c.JSON(200, TokenResponse{AccessToken: out.AccessToken, ExpiresIn: out.ExpiresIn})

### Ação 1.4

Criar ou estender AuthHandler com método `ExchangeCode` (ou `Token`). Registrar rota:

```go
authGroup.POST("/token", authHandler.ExchangeCode)
```

DTO de response:

```go
type TokenResponse struct {
    AccessToken string `json:"access_token"`
    ExpiresIn   int    `json:"expires_in"`
}
```

### Observação 1.4

O nome do campo `access_token` segue a especificação do usuário. O OAuth 2.0 usa `access_token` e `expires_in` como padrão.

---

### Pensamento 5

Mapeamento de erros no ErrorMapper (delivery):
- `domain.ErrCodeInvalidOrExpired` → HTTP 401 com mensagem genérica "Código de autorização inválido ou expirado"
- Não diferenciar "não encontrado", "expirado" ou "já usado" na resposta ao cliente

### Ação 1.5

Adicionar entrada no MapToAppError (ou equivalente) para ErrCodeInvalidOrExpired → 401.

### Observação 1.5

O código HTTP 401 é apropriado — o cliente deve obter um novo code através do fluxo de login novamente.

---

### Pensamento 6

Rate limiting: o endpoint /auth/token pode ser alvo de tentativas de adivinhação de codes. O NGINX (ADR-007) já faz rate limiting por IP. O endpoint deve ter limite similar ao login — ex.: 10 req/min por IP. Verificar configuração do NGINX e adicionar se necessário.

### Ação 1.6

Documentar que o endpoint /api/v1/auth/token deve estar sob rate limiting no NGINX. Implementação na infraestrutura (docs ou config) — não no código Go. A Fase 5 pode incluir essa orientação para deploy.

### Observação 1.6

O código Go pode incluir rate limiting interno se o projeto tiver middleware para isso. Por ora, delegar ao NGINX conforme ADR-007.

---

### Pensamento 7 (Segurança)

- **Code no body, não na URL:** O code é enviado no body do POST, nunca na query string — evita vazamento em logs de servidor e Referer.
- **Single-use:** GetAndDelete garante que o code seja removido. Segunda tentativa com o mesmo code retorna 401.
- **TTL 40s:** Codes expirados já foram removidos pelo Redis; GetAndDelete retorna nil.
- **Sem informações extras:** Na resposta 401, não revelar se o code foi usado, expirado ou inexistente.

### Observação 1.7

Alinhado com OWASP e OAuth 2.0 Security BCP.

---

### Pensamento 8 (Wiring)

O main.go deve:
- Construir ExchangeCodeForTokenUseCase com authCodeStore e jwtSvc
- Passar para o AuthHandler
- Registrar rota POST /api/v1/auth/token

### Ação 1.8

Adicionar método `Token` ou `ExchangeCode` ao AuthHandler. O handler precisa da referência ao ExchangeCodeForTokenUseCase. Atualizar NewAuthHandler para receber esse use case (ou criar handler separado — prefira extender AuthHandler para manter coesão de auth).

### Observação 1.8

O AuthHandler pode ter dois use cases: AuthenticateUser (login) e ExchangeCodeForToken (token). Ambos são operações de auth.

---

## Checklist de Implementação

- [ ] 1. Criar ExchangeCodeForTokenUseCase em internal/usecase/auth/
- [ ] 2. Criar DTOs TokenRequest e TokenResponse
- [ ] 3. Implementar AuthHandler.Token (ou ExchangeCode) que chama o use case
- [ ] 4. Registrar rota POST /api/v1/auth/token
- [ ] 5. Mapear ErrCodeInvalidOrExpired para HTTP 401 no ErrorMapper
- [ ] 6. Resposta 200: { "access_token": "JWT", "expires_in": 900 }
- [ ] 7. Wiring no main.go
- [ ] 8. Documentar endpoint em docs/backend/README.md

---

## Conclusão

A Fase 4 completa o fluxo de troca de code por token. O endpoint /auth/token é stateless em relação ao code — cada code é consumido uma vez. O JWT retornado segue ADR-006 (RS256, 15 min). O cliente usa o access_token para chamadas autenticadas e o expires_in para planejar renovação via refresh token (futuro).
