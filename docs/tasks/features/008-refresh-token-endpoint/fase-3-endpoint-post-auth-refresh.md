# Fase 3 — Endpoint POST /api/v1/auth/refresh

## Objetivo

Implementar o endpoint `POST /api/v1/auth/refresh` que recebe `refresh_token`, `product_slug` e `tenant_slug`, valida o token no banco, revoga o atual, gera novo par (access + refresh) e retorna ambos. Em caso de token inválido (inexistente, revogado ou expirado), retorna 401 Unauthorized com mensagem genérica, sem revelar o motivo específico.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O cliente, ao receber 401 em uma requisição protegida (access token expirado), deve chamar `POST /api/v1/auth/refresh` com:
- `refresh_token`: string (token em texto plano recebido na troca de code)
- `product_slug`: string
- `tenant_slug`: string

O backend valida usando hash do refresh_token + tenant_id + product_id (resolvidos dos slugs).

### Ação 3.1

Criar DTO de request em `internal/delivery/http/dto/refresh.go`:

```go
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	ProductSlug  string `json:"product_slug" binding:"required"`
	TenantSlug   string `json:"tenant_slug" binding:"required"`
}
```

### Observação 3.1

Todos os campos obrigatórios. Ausência retorna 400 via binding. O binding "required" garante validação antes de chamar o Use Case.

---

### Pensamento 2

O Use Case `RefreshTokenUseCase` (ou `RefreshTokenRotationUseCase`) deve:
1. Resolver `tenant_slug` → tenant_id (TenantRepository.FindBySlug)
2. Resolver `product_slug` → product_id (ProductRepository.FindBySlug)
3. Se tenant ou product não encontrado → retornar erro que mapeia para 401 genérico
4. Computar hash SHA-256 do refresh_token
5. Buscar em refresh_tokens WHERE token_hash = hash AND tenant_id = X AND product_id = Y AND revoked_at IS NULL AND expires_at > NOW()
6. Se não encontrado ou inválido → retornar ErrRefreshTokenInvalid (401 genérico)
7. Se válido, em transação:
   - Revogar token atual: UPDATE revoked_at = NOW()
   - Gerar novo access token (JWT 15 min)
   - Gerar novo refresh token (plain + hash)
   - Inserir novo refresh token no banco
   - Retornar access_token, expires_in, refresh_token, refresh_expires_in

### Ação 3.2

Criar `RefreshTokenUseCase` em `internal/usecase/auth/refresh_token.go`:

```go
type RefreshTokenInput struct {
	RefreshToken string
	ProductSlug  string
	TenantSlug   string
}

type RefreshTokenOutput struct {
	AccessToken      string
	ExpiresIn        int
	RefreshToken     string
	RefreshExpiresIn int
}

type RefreshTokenUseCase struct {
	tenantRepo       repository.TenantRepository
	productRepo      repository.ProductRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtSvc           service.JWTService
	refreshTokenGen  service.RefreshTokenGenerator
}
```

### Observação 3.2

O Use Case precisa de TenantRepository e ProductRepository para resolver slugs. O JWTService para assinar o novo access token. O RefreshTokenGenerator para o novo refresh token. O RefreshTokenRepository para Find, Revoke e Create.

---

### Pensamento 3

Erro de domínio: `ErrRefreshTokenInvalid`. Usado para qualquer falha de validação (token não encontrado, revogado, expirado). Nunca diferenciar na resposta HTTP.

### Ação 3.3

Adicionar em `internal/domain/errors.go`:

```go
ErrRefreshTokenInvalid = errors.New("refresh token invalid")
```

Mapear no ErrorMapper: `ErrRefreshTokenInvalid` → HTTP 401 com mensagem genérica "Não autorizado" ou "Credenciais inválidas".

### Observação 3.3

Requisito explícito: "Não revelar o motivo específico". A mensagem deve ser idêntica para token inexistente, revogado ou expirado.

---

### Pensamento 4

O RefreshTokenRepository precisa de método para buscar por hash + tenant_id + product_id. Já especificado na Fase 2: `FindByHashAndTenantAndProduct`. Também precisa revogar (UPDATE revoked_at) e criar. O `RevokeByID` ou `RevokeByHashAndTenantAndProduct` — para rotação, temos o ID do registro encontrado, então `RevokeByID` é suficiente. O repositório já foi planejado na Fase 2.

Para a busca, o repositório pode retornar nil se não encontrar. O Use Case trata nil como ErrRefreshTokenInvalid. Ou o repositório retorna erro específico `ErrRefreshTokenNotFound` e o Use Case converte para ErrRefreshTokenInvalid antes de retornar (para manter a abstração de "invalid" genérico). Preferir: repositório retorna (nil, nil) quando não encontra; Use Case trata como ErrRefreshTokenInvalid.

### Ação 3.4

Assinatura do repositório (já na Fase 2):

```go
FindByHashAndTenantAndProduct(ctx, tokenHash, tenantID, productID) (*entity.RefreshToken, error)
```

Retornar (nil, nil) quando não encontrar — o Use Case interpreta como inválido. Ou retornar (nil, ErrNotFound) e o Use Case sempre retorna ErrRefreshTokenInvalid para o handler.

### Observação 3.4

Para não revelar motivo: mesmo em caso de tenant_slug ou product_slug inexistente, retornar ErrRefreshTokenInvalid (401 genérico). Não retornar 404 para "tenant não encontrado" — isso poderia enumerar tenants. Seguir timing constante: executar hash e query mesmo quando tenant/product não existe, para evitar timing attack. Se tenant não encontrado, fazer query com tenant_id inválido (uuid.Nil) que não retornará resultado — retornar 401.

### Pensamento 4b (Anti-enumeration)

Para evitar enumeração de tenants e products:
- Se tenant_slug não encontrado → não fazer query no refresh_tokens (evitaria vazamento). Retornar 401 com timing normalizado.
- Para timing constante: mesmo em "tenant não encontrado", executar uma operação de tempo similar (ex.: hash dummy + query que não retorna). Ou aceitar que a diferença de tempo entre "tenant inválido" e "token inválido" é mínima (uma query a mais). O requisito prioriza "não revelar motivo" sobre timing. Retornar 401 em todos os casos.

### Observação 4b

Conformidade com ADR-010 e code guidelines (credenciais: nunca diferenciar).

---

### Pensamento 5

Transação atômica: revogar token antigo + inserir novo. Se a inserção falhar após revogar, o usuário perde o refresh token (precisa fazer login novamente). O ideal é fazer em transação:
1. BEGIN
2. SELECT ... FOR UPDATE (bloquear linha do refresh token)
3. Verificar revoked_at IS NULL e expires_at > NOW()
4. UPDATE revoked_at = NOW()
5. Gerar novo access token e refresh token
6. INSERT novo refresh token
7. COMMIT

O `FindByHashAndTenantAndProduct` pode receber um `*gorm.DB` de transação para executar dentro dela. Ou o Use Case chama `db.Transaction` e passa o tx para os métodos do repositório. O GORM suporta `tx.Transaction`. O repositório precisa aceitar transação ou o Use Case gerencia a transação e chama múltiplos métodos do repo (Create, Update) — cada um usa o mesmo *gorm.DB. A abordagem mais limpa: o RefreshTokenRepository tem método `FindAndRevoke` que faz SELECT FOR UPDATE + UPDATE em uma transação interna, ou o Use Case inicia transação e passa contexto com tx. Em Go com GORM, o padrão é `db.Transaction(func(tx *gorm.DB) error { ... })` e os repositórios têm `WithDB(tx)` ou recebem o tx. Para simplicidade, o repositório pode ter `FindByHashAndTenantAndProduct` e `RevokeByID` separados; o Use Case chama `db.Transaction` e dentro da transaction chama repo.Find (que usa db do repo) — mas o repo está vinculado ao db global. Para transação, precisamos que o repo use o tx. A interface do repositório não recebe tx; a implementação GORM usa o db injetado no construtor. Para transações, temos duas opções:
a) Repositório com método `RunInTransaction(ctx, fn func(repo RefreshTokenRepository) error)` — complexo
b) Use Case recebe *gorm.DB e inicia transação; cria "transaction-scoped" repository que usa tx
c) Repositório tem `RevokeAndCreate(ctx, oldID uuid.UUID, newToken *entity.RefreshToken) error` que executa ambos em transação interna

A opção (c) mantém a atomicidade no repositório. Criar método `RotateRefreshToken(ctx, oldToken *entity.RefreshToken, newToken *entity.RefreshToken) error` que: revoga old, cria new, em transação.

### Ação 3.5

Adicionar ao `RefreshTokenRepository`:

```go
Rotate(ctx context.Context, oldTokenID uuid.UUID, newToken *entity.RefreshToken) error
```

Implementação: dentro de `db.Transaction`, executar `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = ?` e `INSERT` do novo token. Retorna erro se qualquer operação falhar (rollback automático).

### Observação 3.5

Conformidade com ADR-003 (transações ACID) e ADR-010 (race condition: SELECT FOR UPDATE).

---

### Pensamento 6

Para gerar o JWT no refresh, precisamos dos claims: Subject (user_id), Audience, TenantID, UserAccessProfile. Esses vêm do refresh token encontrado (que tem user_id, tenant_id, product_id). Precisamos de UserAccessProfile — não está na tabela refresh_tokens. A tabela tem user_id, tenant_id, product_id. O UserAccessProfile vem de user_product_access. O Use Case deve buscar o UserProductAccess para obter o access_profile ao gerar o JWT. Ou armazenar UserAccessProfile na tabela refresh_tokens no momento da criação (desnormalização) para evitar join na renovação. A tabela refresh_tokens não tem access_profile. Opções:
1. Adicionar coluna `access_profile` em refresh_tokens (migration adicional)
2. Buscar user_product_access no Use Case de refresh

A opção 2 não requer migration. O Use Case já depende de UserProductAccessRepository ou similar. Verificar se existe. Sim, o AuthenticateUser tem userProductAccRepo. O RefreshTokenUseCase pode depender de UserProductAccessRepository e chamar FindByUserAndProduct para obter o profile. Default "view" se não encontrar.

### Ação 3.6

Adicionar `UserProductAccessRepository` ao RefreshTokenUseCase. Após encontrar o refresh token, chamar `userProductAccRepo.FindByUserAndProduct(ctx, rt.UserID, rt.ProductID)` para obter access_profile (default "view"). Montar JWTClaims com Subject=rt.UserID, TenantID=rt.TenantID, Audience (montado a partir de tenant_slug e product — precisamos do tenant e product para montar aud). O Audience no login é `tenant_slug.base_domain`; no JWT a aud pode incluir o product. Verificar ADR-006: aud é "domínio do cliente". O redirectBaseDomain e tenant_slug montam a aud. Para o refresh, temos tenant_id e product_id; precisamos de tenant.Slug e product.Slug para montar aud — ou podemos obter do product/tenant já resolvidos pelos slugs. Temos tenant e product ao resolver os slugs. Então: audience = fmt.Sprintf("%s.%s", tenantSlug, redirectBaseDomain) — igual ao login. O product_slug pode fazer parte do path, não da aud. Seguir o mesmo formato do AuthCodeData no login.

### Observação 3.6

Consistência com o formato de audience existente.

---

### Pensamento 7

Handler: receber RefreshRequest, chamar Use Case, retornar 200 com TokenResponse (ou RefreshResponse com mesmos campos). Em caso de erro: RespondWithError que mapeia ErrRefreshTokenInvalid para 401.

### Ação 3.7

Criar `AuthHandler.Refresh` ou adicionar método ao AuthHandler existente:

```go
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, apperrors.NewInvalidRequest("Dados inválidos. Verifique refresh_token, product_slug e tenant_slug."))
		return
	}
	out, err := h.refreshTokenUseCase.Execute(c.Request.Context(), auth.RefreshTokenInput{
		RefreshToken: req.RefreshToken,
		ProductSlug:  req.ProductSlug,
		TenantSlug:   req.TenantSlug,
	})
	if err != nil {
		deliveryerrors.RespondWithError(c, err)
		return
	}
	c.JSON(200, dto.RefreshResponse{
		AccessToken:      out.AccessToken,
		ExpiresIn:        out.ExpiresIn,
		RefreshToken:     out.RefreshToken,
		RefreshExpiresIn: out.RefreshExpiresIn,
	})
}
```

Registrar rota: `authGroup.POST("/refresh", authHandler.Refresh)`

### Observação 3.7

O DTO TokenResponse já tem access_token e expires_in. O RefreshResponse pode reutilizar ou ter os mesmos campos + refresh_token e refresh_expires_in. Criar RefreshResponse para clareza ou estender TokenResponse. Como o TokenResponse da Fase 2 já inclui refresh_token e refresh_expires_in, podemos ter um único DTO `TokenResponse` usado tanto em /token quanto em /refresh. Sim — ambos retornam a mesma estrutura.

---

### Pensamento 8 (Validação da assinatura)

O requisito e ADR-010 mencionam verificar se a assinatura do tenant está ativa na renovação. O RefreshTokenUseCase deve, após validar o token, verificar a subscription (tenant + product) antes de emitir novos tokens. Se assinatura vencida/suspensa → retornar 402 Payment Required (conforme ADR-010) ou 401? O requisito do usuário diz "caso inválido → 401". O ADR-010 diz "Se assinatura vencida → 402". Incluir verificação de assinatura no Use Case: após validar refresh token, buscar subscription por tenant+product; se status != active/pending, retornar ErrSubscriptionExpired ou similar que mapeia para 402. Isso permite que o cliente redirecione para renovação de plano. Documentar na Fase 4.

### Ação 3.8

Incluir verificação de assinatura no RefreshTokenUseCase:
- Após validar refresh token e antes de gerar novos tokens
- Chamar subscriptionRepo.FindByTenantAndProduct
- Se status suspended/canceled ou expirado → retornar ErrSubscriptionExpired (402)
- Erros de refresh token → 401; assinatura vencida → 402

### Observação 3.8

O requisito do usuário foca em "token inválido → 401". A verificação de assinatura é adicional e segue ADR-010. Incluir no plano.

---

### Pensamento 9 (Rate limiting)

O endpoint /refresh pode ser alvo de tentativas de adivinhação de refresh tokens. Rate limiting via NGINX (ADR-007). Documentar na Fase 4.

### Observação 3.9

Sem alteração de código Go para rate limiting; documentar em docs.

---

## Checklist de Implementação

- [ ] 1. Criar ErrRefreshTokenInvalid em domain/errors.go
- [ ] 2. Adicionar Rotate ao RefreshTokenRepository (ou RevokeByID + Create em transação no Use Case)
- [ ] 3. Criar RefreshTokenUseCase com toda a lógica
- [ ] 4. Incluir verificação de assinatura (SubscriptionRepository) antes de renovar
- [ ] 5. Criar DTOs RefreshRequest e RefreshResponse
- [ ] 6. Criar AuthHandler.Refresh
- [ ] 7. Registrar rota POST /api/v1/auth/refresh
- [ ] 8. Mapear ErrRefreshTokenInvalid → 401 no ErrorMapper
- [ ] 9. Mapear ErrSubscriptionExpired → 402 no ErrorMapper (se não existir)
- [ ] 10. Wiring no main.go
- [ ] 11. Documentar endpoint em docs/backend/README.md

---

## Referências

- ADR-006 — JWT, Refresh Token rotation
- ADR-010 — Fluxo de renovação, verificação de assinatura
- docs/code_guidelines/backend.md — Tratamento de erros, segurança
- Requisito da feature: validação por refresh_token, product_slug, tenant_slug
