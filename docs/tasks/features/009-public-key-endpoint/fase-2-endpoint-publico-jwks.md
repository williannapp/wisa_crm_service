# Fase 2 — Endpoint Público GET /.well-known/jwks.json

## Objetivo

Implementar o handler e registrar a rota `GET /.well-known/jwks.json` que retorna o JWKS (chaves públicas) em formato JSON. O endpoint não exige autenticação, retorna Content-Type application/json e inclui header Cache-Control para cache de 24 horas. O kid das chaves deve corresponder ao kid dos JWTs emitidos.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O path `/.well-known/jwks.json` é o padrão da indústria (RFC 7517, OpenID Connect, utilizado por Auth0, Keycloak, etc.). O Gin router suporta paths com pontos e barras. A rota deve ser registrada **fora** do grupo `/api/v1/auth` para não passar por middlewares de autenticação (se houver no futuro). Atualmente, não há middleware de auth nas rotas — todas são públicas exceto pela lógica interna dos handlers de login/token/refresh que validam credenciais no body. O endpoint JWKS é de leitura (GET) e não requer nenhuma validação.

### Ação 2.1

Registrar a rota no nível raiz do router, antes ou depois do `router.GET("/health", ...)`:

```go
router.GET("/.well-known/jwks.json", jwksHandler.GetJWKS)
```

A rota deve estar disponível **independentemente** de haver banco de dados ou não. O `/health` funciona sem DB. O JWKS depende da chave privada — e a chave privada só é carregada quando `jwtPrivateKeyPath != ""` (quando há auth). Portanto, o endpoint JWKS só deve ser registrado quando o JWT service estiver configurado (mesmo bloco onde estão as rotas de auth).

### Observação 2.1

Conformidade com o main.go atual: as rotas de auth estão dentro de `if db != nil && jwtPrivateKeyPath != ""`. O JWKS deve estar no mesmo bloco — sem chave privada, não há chave pública para expor. Retornar 503 ou não registrar a rota. Se não houver chave, não registrar o handler JWKS — o Gin retornará 404 para `/.well-known/jwks.json`. Alternativa: registrar sempre e retornar 503 se o provider não estiver disponível. Preferir: registrar apenas quando houver chave, para simplificar.

### Pensamento 2

O handler precisa do JWKSProvider injetado. Criar `JWKSHandler` em `internal/delivery/http/handler/jwks_handler.go` ou adicionar método ao `AuthHandler`. O JWKS não é estritamente parte do fluxo de auth — é um endpoint de descoberta. Criar handler separado `JWKSHandler` mantém a separação de responsabilidades.

### Ação 2.2

Criar `JWKSHandler` em `internal/delivery/http/handler/jwks_handler.go`:

```go
type JWKSHandler struct {
    provider service.JWKSProvider
}

func NewJWKSHandler(provider service.JWKSProvider) *JWKSHandler {
    return &JWKSHandler{provider: provider}
}

func (h *JWKSHandler) GetJWKS(c *gin.Context) {
    keys, err := h.provider.GetKeys(c.Request.Context())
    if err != nil {
        // Log e responder 500
        return
    }
    c.Header("Content-Type", "application/json")
    c.Header("Cache-Control", "public, max-age=86400")
    c.JSON(200, JWKSResponse{Keys: keys})
}
```

### Observação 2.2

O `Content-Type: application/json` é definido automaticamente por `c.JSON()`. O header `Cache-Control` deve ser adicionado explicitamente antes do `c.JSON`. A struct de resposta pode ser `map[string]interface{}{"keys": keys}` ou uma struct `JWKSResponse` com tag JSON.

### Pensamento 3

O requisito especifica que o endpoint deve estar disponível **exclusivamente via HTTPS**. O backend Go escuta em `127.0.0.1:8080` (ou `:8080`) — HTTP. A terminação TLS é feita pelo NGINX (ADR-007). O backend não precisa implementar verificação de HTTPS — em produção, o NGINX recebe HTTPS e faz proxy para o backend em HTTP local. O requisito "exclusivamente via HTTPS" é atendido pela arquitetura: o cliente nunca acessa o backend diretamente; acessa via NGINX que só aceita HTTPS na porta 443. Documentar isso na Fase 3.

### Ação 2.3

Nenhuma alteração no handler para forçar HTTPS. O backend não verifica `X-Forwarded-Proto` para este endpoint — não é necessário.

### Observação 2.3

Conformidade com ADR-007: NGINX termina TLS; backend em loopback.

### Pensamento 4

Erro no GetKeys: se o provider falhar (ex.: arquivo de chave corrompido após startup), retornar 500. Não expor detalhes do erro ao cliente. Usar o middleware Recovery para panics. O handler deve usar `RespondWithError` ou equivalente? O `RespondWithError` mapeia erros de domínio para códigos HTTP. Para erro interno do provider (não é erro de domínio típico), logar e retornar 500 genérico. O pacote `pkg/errors` tem `INTERNAL_ERROR`. Verificar se o handler deve usar `httperrors.RespondWithError(c, err)` — o ErrorMapper pode não ter mapeamento para um erro genérico de provider. Criar `ErrJWKSUnavailable` em domain ou usar `fmt.Errorf` e mapear para 500. Para simplicidade: `log.Error` e `c.JSON(500, map[string]string{"error": "Service temporarily unavailable"})` — ou usar um AppError com código INTERNAL. O code_guidelines diz para usar RespondWithError. Adicionar um erro catalogado `JWKS_UNAVAILABLE` ou reutilizar `INTERNAL_ERROR`.

### Ação 2.4

Em caso de erro no provider:
```go
log.Printf("JWKS provider error: %v", err)
c.JSON(500, gin.H{"error": "Service temporarily unavailable"})
```
Ou mapear para AppError existente. O `RespondWithError` espera um `*apperrors.AppError` ou erro que o `MapToAppError` reconheça. Para erro genérico 500, podemos usar `apperrors.NewInternalError("")` ou similar. Verificar o catálogo em pkg/errors.

### Observação 2.4

Não revelar stack trace ou path de arquivo em produção. Mensagem genérica.

### Pensamento 5

Wiring no main.go:
- Instanciar `RSAJWKSProvider` com o mesmo `RSAJWTConfig` usado pelo `RSAJWTService` (PrivateKeyPath, KeyID)
- Criar `JWKSHandler` com o provider
- Registrar `router.GET("/.well-known/jwks.json", jwksHandler.GetJWKS)` dentro do bloco onde o jwtSvc é criado

O provider e o JWT service compartilham a mesma chave. Ambos usam `JWT_PRIVATE_KEY_PATH` e `JWT_KEY_ID`. Garantir que o kid no JWK seja idêntico ao kid no header do JWT.

### Ação 2.5

No main.go, após criar jwtSvc:
```go
jwksProvider, err := crypto.NewRSAJWKSProvider(crypto.RSAJWTConfig{
    PrivateKeyPath: jwtPrivateKeyPath,
    KeyID:          getEnv("JWT_KEY_ID", "key-2026-v1"),
})
if err != nil {
    log.Fatalf("JWKS provider initialization failed: %v", err)
}
jwksHandler := handler.NewJWKSHandler(jwksProvider)
router.GET("/.well-known/jwks.json", jwksHandler.GetJWKS)
```

A rota deve ser registrada **antes** do `authGroup` para que a ordem de registro seja clara. Ou registrar no mesmo nível que `/health` — o `/.well-known/jwks.json` é uma rota especial. Em Gin, a ordem de registro define a precedência. Não há conflito com outras rotas.

### Observação 2.5

O RSAJWTConfig tem também Issuer e ExpMinutes — o JWKSProvider não precisa deles. Podemos criar um config reduzido ou usar o mesmo. Usar o mesmo RSAJWTConfig com os campos não utilizados ignorados é mais simples.

### Pensamento 6

Rota com prefixo `/api/` no NGINX: O ADR-007 mostra `location /api/` para o proxy. O path `/.well-known/jwks.json` **não** está sob `/api/`. O NGINX precisa de um `location` específico para `/.well-known/jwks.json`. A ADR-007 já documenta:
```nginx
location /.well-known/jwks.json {
    proxy_pass http://127.0.0.1:8080;
    add_header Cache-Control "public, max-age=86400";
}
```
O backend também adiciona Cache-Control no handler — redundante mas inofensivo. O NGINX pode sobrescrever com `add_header` — o último ganha. Para garantir que o header esteja na resposta, o backend deve defini-lo. Se o NGINX fizer proxy e não modificar headers, o header do backend será repassado. Manter o header no handler do Go.

### Observação 2.6

Conformidade com ADR-007. A documentação da Fase 3 deve garantir que a config NGINX esteja aplicada.

### Pensamento 7

Teste manual: após implementação, `curl http://localhost:8080/.well-known/jwks.json` deve retornar:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-2026-v1",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```
O valor `e` para exponent 65537 em base64url é `AQAB` (bytes 0x01, 0x00, 0x01).

### Observação 2.7

Validar que o cliente possa usar essa chave para verificar um JWT emitido pelo serviço.

---

## Checklist de Implementação

- [ ] 1. Criar `internal/delivery/http/handler/jwks_handler.go` com JWKSHandler e GetJWKS
- [ ] 2. Adicionar header Cache-Control: public, max-age=86400
- [ ] 3. Tratar erro do provider com 500 e mensagem genérica
- [ ] 4. Criar RSAJWKSProvider (Fase 1) e instanciar no main
- [ ] 5. Registrar rota GET `/.well-known/jwks.json` no main.go (dentro do bloco com JWT)
- [ ] 6. Garantir Content-Type application/json (c.JSON já define)
- [ ] 7. Teste: curl local e verificar estrutura da resposta

---

## Referências

- ADR-006 — JWKS endpoint
- ADR-007 — NGINX, location /.well-known/jwks.json
- RFC 7517 — JWKS structure
- docs/code_guidelines/backend.md — handlers, middlewares
