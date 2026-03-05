# Fase 2 — Refresh Token no fluxo de troca de code (POST /auth/token)

## Objetivo

Estender o endpoint `POST /api/v1/auth/token` para gerar, persistir e retornar um refresh token além do access token. O refresh token será armazenado como hash SHA-256 no banco, com expiração de 7 dias, conforme ADR-006 e requisito da feature. Depende da Fase 1 (product_id em refresh_tokens).

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O `AuthCodeData` atualmente contém `Subject`, `Audience`, `TenantID`, `UserAccessProfile`. Para criar um refresh token no momento da troca de code, precisamos de `UserID`, `TenantID` e `ProductID`. O Subject já é o user_id. O TenantID já existe. Falta **ProductID** — o product foi validado no login, mas não está armazenado no AuthCodeData. O Audience é montado como `tenant_slug.base_domain`, não contém product. Devemos adicionar `ProductID` ao AuthCodeData.

### Ação 2.1

Adicionar `ProductID string` ao `AuthCodeData` em `internal/domain/service/auth_code_store.go`:

```go
type AuthCodeData struct {
	Subject           string // user_id (ULID)
	Audience          string
	TenantID          string
	ProductID         string // product_id (UUID) — para criar refresh token escopado
	UserAccessProfile string
}
```

### Observação 2.1

O `AuthenticateUserUseCase` já tem acesso a `product.ID` no momento do login. Deve incluir `ProductID: product.ID.String()` ao montar o `AuthCodeData`.

---

### Pensamento 2

Precisamos de um repositório para `refresh_tokens`. Seguindo Clean Architecture: interface em `domain/repository`, implementação em `infrastructure/persistence`. A entidade RefreshToken pode ser um struct simples em domain/entity ou usarmos um model GORM. O code guideline recomenda entidades puras no domínio; o repositório abstrai a persistência.

### Ação 2.2

Criar interface `RefreshTokenRepository` em `internal/domain/repository/refresh_token_repository.go`:

```go
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	FindByHashAndTenantAndProduct(ctx context.Context, tokenHash string, tenantID, productID uuid.UUID) (*entity.RefreshToken, error)
	RevokeByID(ctx context.Context, id uuid.UUID) error
}
```

E entidade `RefreshToken` em `internal/domain/entity/refresh_token.go`:

```go
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TenantID  uuid.UUID
	ProductID uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}
```

### Observação 2.2

O `FindByHashAndTenantAndProduct` será usado no endpoint de refresh. O `RevokeByID` será usado ao revogar o token atual durante a rotação. O `Create` insere novos tokens. Alternativa: `UpdateRevokedAt` em vez de método específico; `RevokeByID` deixa a semântica clara.

---

### Pensamento 3

Precisamos gerar um refresh token aleatório e computar seu hash SHA-256. O token em texto plano (ex.: 32 bytes hex = 64 caracteres) é retornado ao cliente; o hash é armazenado no banco. Usar `crypto/rand` para geração segura. Criar função utilitária ou serviço em `internal/infrastructure/crypto` ou `pkg` — evitar que o domínio dependa de crypto. O Use Case pode receber uma interface `RefreshTokenGenerator` ou uma função. Para simplicidade, criar `GenerateRefreshToken() (plain string, hash string, err)` em um pacote interno (ex.: `internal/infrastructure/crypto/refresh_token.go`).

### Ação 2.3

Criar `internal/infrastructure/crypto/refresh_token.go`:

```go
// GenerateRefreshToken creates a cryptographically random refresh token and its SHA-256 hash.
// Returns (plainToken, hashHex, error). hashHex is 64 chars (SHA-256 in hex).
func GenerateRefreshToken() (plain string, hashHex string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	plain = hex.EncodeToString(bytes)
	h := sha256.Sum256([]byte(plain))
	hashHex = hex.EncodeToString(h[:])
	return plain, hashHex, nil
}
```

### Observação 2.3

O token em texto plano tem 64 caracteres (32 bytes em hex). O hash SHA-256 também 64 caracteres. A coluna `token_hash` é CHAR(64), compatível.

---

### Pensamento 4

O `ExchangeCodeForTokenUseCase` precisa:
1. Receber `RefreshTokenRepository` como dependência
2. Após obter AuthCodeData e gerar JWT, gerar refresh token
3. Persistir o refresh token no banco (user_id, tenant_id, product_id, token_hash, expires_at = NOW() + 7 dias)
4. Retornar refresh_token (plain) e refresh_expires_in no output

O Use Case não deve conhecer detalhes de infraestrutura (crypto). Injetar uma interface ou função que gera o token. Para manter o Use Case enxuto, podemos ter:
- `type RefreshTokenGenerator interface { Generate() (plain, hash string, err error) }`
- Ou passar a função diretamente

Para pragmatismo (conforme code guideline: interface apenas quando ≥2 implementações), usar função concreta no wiring. O Use Case receberá `RefreshTokenRepository` e uma função `generateRefreshToken func() (string, string, error)`. Ou criar um serviço `RefreshTokenService` com método `GenerateAndHash`. Seguindo padrão do `BcryptPasswordService`, criar `RefreshTokenGenerator` como interface mínima no domain/service e implementação em infrastructure/crypto.

### Ação 2.4

Criar interface em `internal/domain/service/refresh_token_generator.go`:

```go
type RefreshTokenGenerator interface {
	Generate() (plain string, hash string, err error)
}
```

Implementação: `internal/infrastructure/crypto/refresh_token_generator.go` com a lógica do GenerateRefreshToken. O Use Case dependerá da interface.

### Observação 2.4

Dependency Inversion: Use Case depende de interface, não da implementação concreta. Permite mock em testes.

---

### Pensamento 5

O `ExchangeCodeForTokenUseCase` deve:
1. GetAndDelete(code) — já existe
2. Se data == nil ou err → return ErrCodeInvalidOrExpired
3. Sign JWT — já existe
4. **Novo:** Gerar refresh token (plain, hash)
5. **Novo:** Calcular expires_at = time.Now().UTC().Add(7 * 24 * time.Hour)
6. **Novo:** Criar entidade RefreshToken e persistir via repository
7. **Novo:** Incluir RefreshToken (plain) e RefreshExpiresIn no output

O AuthCodeData precisa ter ProductID. Se a Fase 1 não tiver rodado ainda, o AuthenticateUser precisa ser alterado nesta mesma fase para incluir ProductID no Store. A ordem de implementação: primeiro migration (Fase 1), depois código. O AuthenticateUser já tem product no escopo — basta adicionar `ProductID: product.ID.String()` ao data.

### Ação 2.5

Fluxo no `ExchangeCodeForTokenUseCase.Execute`:
- Após Sign JWT
- `plainRT, hashRT, err := uc.refreshTokenGen.Generate()`
- `userID, _ := uuid.Parse(data.Subject)`; `tenantID, _ := uuid.Parse(data.TenantID)`; `productID, _ := uuid.Parse(data.ProductID)`
- `expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)`
- `rt := &entity.RefreshToken{ UserID: userID, TenantID: tenantID, ProductID: productID, TokenHash: hashRT, ExpiresAt: expiresAt, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC() }`
- `uc.refreshTokenRepo.Create(ctx, rt)`
- Retornar `ExchangeCodeOutput{ AccessToken: token, ExpiresIn: 900, RefreshToken: plainRT, RefreshExpiresIn: 7*24*3600 }`

### Observação 2.5

RefreshExpiresIn em segundos (604800 = 7 dias) para o cliente saber quando o refresh expira. O DTO TokenResponse deve incluir `refresh_token` e `refresh_expires_in`.

---

### Pensamento 6 (Segurança)

- Refresh token NUNCA logado
- Hash SHA-256 armazenado; plain retornado apenas na resposta
- Transação: Create do refresh token e retorno do JWT — se Create falhar, retornar erro (o code já foi consumido; cliente precisará novo login). O code é single-use, então não há rollback do Redis. A persistência do refresh token deve suceder; caso contrário, retornar 500 e o cliente refaz o fluxo de login.
- ProductID vem do AuthCodeData gerado no login — não é fornecido pelo cliente na troca de code, reduzindo superfície de ataque.

### Observação 2.6

Conformidade com ADR-006 (refresh token 7 dias, hash SHA-256) e code guidelines (não logar tokens).

---

### Pensamento 7 (Implementação GORM)

O repositório GORM precisará de um model. Criar `RefreshTokenModel` em `internal/infrastructure/persistence/model` ou similar, com tags GORM mapeando para `wisa_crm_db.refresh_tokens`. O schema é `wisa_crm_db`; a tabela foi criada com prefixo.

### Ação 2.7

Criar `gorm_refresh_token_repository.go` que implementa `RefreshTokenRepository`. O model GORM usa `gorm.Model` ou campos explícitos. A tabela é `wisa_crm_db.refresh_tokens`; configurar `TableName()` para retornar `wisa_crm_db.refresh_tokens`. Ou usar `db.Table("wisa_crm_db.refresh_tokens")` nas queries. Verificar como outras tabelas estão configuradas no projeto (search_path ou tabela qualificada).

### Observação 2.7

Consultar implementações existentes (GormUserRepository, etc.) para manter consistência.

---

### Pensamento 8 (Wiring)

O `main.go` deve:
- Construir `GormRefreshTokenRepository`
- Construir `RefreshTokenGenerator` (implementação crypto)
- Passar ambos para `NewExchangeCodeForTokenUseCase`
- Atualizar assinatura de `NewExchangeCodeForTokenUseCase` para aceitar esses parâmetros

O `AuthenticateUserUseCase` deve incluir `ProductID` ao montar `AuthCodeData`. Nenhuma alteração de assinatura necessária — apenas adicionar um campo ao struct literal.

### Observação 2.8

Múltiplas dependências novas no Use Case; manter ordem lógica de parâmetros.

---

## Checklist de Implementação

- [ ] 1. Adicionar `ProductID` ao `AuthCodeData` em domain/service
- [ ] 2. Alterar `AuthenticateUserUseCase` para incluir `ProductID` no AuthCodeData
- [ ] 3. Criar entidade `RefreshToken` em domain/entity
- [ ] 4. Criar interface `RefreshTokenRepository` em domain/repository
- [ ] 5. Criar `RefreshTokenGenerator` (interface + implementação crypto)
- [ ] 6. Criar `GormRefreshTokenRepository` em infrastructure/persistence
- [ ] 7. Alterar `ExchangeCodeForTokenUseCase`: injetar repo e generator, gerar e persistir refresh token
- [ ] 8. Alterar `ExchangeCodeOutput` e `TokenResponse` para incluir refresh_token e refresh_expires_in
- [ ] 9. Wiring no main.go
- [ ] 10. Atualizar documentação do endpoint em docs/backend/README.md

---

## Referências

- ADR-006 — JWT e Refresh Token (7 dias, hash SHA-256)
- docs/code_guidelines/backend.md — Clean Architecture, repositórios
- backend/internal/domain/service/auth_code_store.go
- backend/internal/usecase/auth/exchange_code_for_token.go
