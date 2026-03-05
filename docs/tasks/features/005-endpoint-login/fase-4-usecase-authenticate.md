# Fase 4 — Use Case AuthenticateUser

## Objetivo

Implementar o caso de uso de autenticação que orquestra as validações (tenant, email, senha, status do usuário, assinatura) e emite o JWT quando todas passam. Em conformidade com ADR-010 (mensagem genérica para auth, timing constante) e ADR-006 (estrutura do JWT).

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O fluxo do use case, na ordem das validações:

1. **Tenant** — FindBySlug(slug). Se não encontrar → ErrInvalidCredentials (ADR-010: nunca revelar se tenant existe)
2. **Product** — FindBySlug(product_slug). Se não encontrar → ErrInvalidCredentials. Se status != "active" (inactive ou blocked) → ErrProductUnavailable
3. **User** — FindByEmailAndTenantID(tenantID, email). Se não encontrar → executar bcrypt com hash dummy (timing), retornar ErrInvalidCredentials
3. **Password** — Compare(password, user.PasswordHash). Se inválido → ErrInvalidCredentials
4. **User Status** — Se user.Status != "active" → ErrUserBlocked
5. **Subscription** — FindByTenantAndProduct(tenantID, productID). Se não encontrar → ErrSubscriptionExpired. Se status == "suspended" → ErrSubscriptionSuspended. Se status == "canceled" → ErrSubscriptionCanceled. Se active ou pending → ok
6. **Access Profile** — FindByUserAndProduct(userID, productID). Se não encontrar → usar "view" como default
7. **JWT** — Construir claims, chamar JWTService.Sign, retornar token

### Ação 1.1

Definir Input e Output do use case:
```go
type LoginInput struct {
    Slug        string
    ProductSlug string  // identificador do produto na URL (ex: https://slug.domain.com.br/product_slug)
    UserEmail   string
    Password    string
}

type LoginOutput struct {
    Token string
}
```

### Observação 1.1

O use case não conhece HTTP. Input vem do handler após parse do body. Output é o token para o handler encapsular na resposta JSON.

---

### Pensamento 2

Proteção contra timing attack (ADR-010): quando o usuário não existe, ainda assim executar bcrypt.Compare com um hash dummy. O hash dummy deve ser um bcrypt válido para não causar panic. Exemplo: `$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.G4jSH.7r8K9KQm` (bcrypt válido de uma string conhecida). Usar constante no use case.

### Ação 1.2

```go
const dummyBcryptHash = "$2a$12$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
```
Ou gerar um hash estático e armazenar como constante. O importante: sempre chamar Compare, com user.PasswordHash ou dummyBcryptHash.

### Observação 1.2

O tempo de resposta será similar para "user not found" e "wrong password". Protege contra enumeração de usuários via timing.

---

### Pensamento 3

Ordem de validação e early return: validar tenant primeiro (sem isso, não há tenantID para buscar user). Depois user. Para user, a estratégia: buscar user; se err != nil, definir hashToCompare = dummy, chamar Compare(password, hashToCompare), retornar ErrInvalidCredentials. Se user encontrado, hashToCompare = user.PasswordHash, chamar Compare. Se !valid, retornar ErrInvalidCredentials.

### Ação 1.3

Pseudocódigo:
```go
tenant, err := uc.tenantRepo.FindBySlug(ctx, input.Slug)
if err != nil {
    return nil, domain.ErrInvalidCredentials  // nunca ErrTenantNotFound para cliente
}
product, err := uc.productRepo.FindBySlug(ctx, input.ProductSlug)
if err != nil {
    return nil, domain.ErrInvalidCredentials  // nunca ErrProductNotFound para cliente
}
if product.Status != "active" {
    return nil, domain.ErrProductUnavailable  // inactive ou blocked
}
user, err := uc.userRepo.FindByEmailAndTenantID(ctx, tenant.ID, input.UserEmail)
hashToCompare := dummyHash
if err == nil {
    hashToCompare = user.PasswordHash
}
if !uc.passwordSvc.Compare(input.Password, hashToCompare) {
    return nil, domain.ErrInvalidCredentials
}
if err != nil {
    return nil, domain.ErrInvalidCredentials  // user não existe
}
if user.Status != "active" {
    return nil, domain.ErrUserBlocked
}
subscription, err := uc.subscriptionRepo.FindByTenantAndProduct(ctx, tenant.ID, product.ID)
// ... validar status da assinatura, access profile, JWT
```

### Observação 1.3

A ordem garante que bcrypt sempre execute quando há tentativa de login, independente de user existir. Correto para timing.

---

### Pensamento 4

Subscription: FindByTenantAndProduct(tenantID, productID). Se err != nil → ErrSubscriptionExpired. Se subscription.Status == "suspended" → ErrSubscriptionSuspended. Se "canceled" → ErrSubscriptionCanceled. Se "active" ou "pending" → prosseguir.

O tenant pode ter múltiplas subscriptions para produtos diferentes. O product_id vem do Product encontrado por product_slug.

### Ação 1.4

Após validar subscription (tenant_id + product_id):
```go
accessProfile := "view"
upa, err := uc.userProductAccessRepo.FindByUserAndProduct(ctx, user.ID, product.ID)
if err == nil {
    accessProfile = upa.AccessProfile
}
```

### Observação 1.4

Default "view" quando não há registro em user_product_access. Valores possíveis: admin, operator, view (conforme banco).

---

### Pensamento 5

Claim `aud`: construir sempre com `tenant.Slug` + domínio base (ex: `cliente1.app.wisa-crm.com`). O slug já vem no redirect do sistema do cliente; usar `JWT_AUD_BASE_DOMAIN` do config para compor o aud completo.

### Ação 1.5

Lógica:
```go
// Construir aud: slug + domínio base (ex: cliente1.app.wisa-crm.com)
aud := tenant.Slug + "." + uc.audienceBaseDomain
```

O use case recebe no construtor: `audienceBaseDomain` (ex: "app.wisa-crm.com") para compor o aud: `tenant.Slug + "." + audienceBaseDomain`.

### Observação 1.5

Decisão: aud = `tenant.Slug + "." + JWT_AUD_BASE_DOMAIN`. O slug identifica o tenant e o subdomínio do sistema do cliente; sem necessidade de campo extra no banco.

---

### Pensamento 6

Estrutura do JWT conforme especificação:
- iss: do config (JWT_ISSUER)
- sub: user.ID.String()
- aud: definido acima
- tenant_id: tenant.ID.String()
- user_access_profile: accessProfile (admin, operator, view)
- jti: uuid.New().String()
- iat: time.Now().Unix()
- exp: iat + 15*60 (ou configurável)
- nbf: iat

O JWTService.Sign recebe esses valores. O use case monta a struct de claims e chama Sign.

### Ação 1.6

O use case depende de: TenantRepository, UserRepository, SubscriptionRepository, UserProductAccessRepository, PasswordService, JWTService. E de config para issuer e audienceBaseDomain. O aud é construído no use case como slug + base domain.

### Observação 1.6

O JWTService pode receber issuer no construtor. O use case passa aud, sub, tenant_id, etc. nas claims. Assim o use case tem controle sobre os valores dinâmicos.

---

### Pensamento 7

Tratamento de ErrTenantNotFound, ErrProductNotFound e ErrUserNotFound do repositório: o use case deve convertê-los em ErrInvalidCredentials antes de retornar (ADR-010). Assim o ErrorMapper nunca recebe esses erros no contexto de login — apenas ErrInvalidCredentials.

### Ação 1.7

Em todos os pontos onde o repositório pode retornar ErrTenantNotFound ou ErrUserNotFound no fluxo de auth, converter para ErrInvalidCredentials antes de retornar.

### Observação 1.7

Conformidade com ADR-010. O cliente nunca sabe se falhou por tenant, email ou senha.

---

### Decisão final Fase 4

**Implementar:**

1. Criar `usecase/auth/authenticate_user.go`
2. Input: LoginInput (Slug, ProductSlug, UserEmail, Password); Output: LoginOutput (Token)
3. Dependências: TenantRepo, ProductRepo, UserRepo, SubscriptionRepo, UserProductAccessRepo, PasswordService, JWTService, audienceBaseDomain (config).
4. Fluxo:
   - Buscar tenant por slug → ErrInvalidCredentials se não encontrar
   - Buscar product por product_slug → ErrInvalidCredentials se não encontrar; ErrProductUnavailable se status != active
   - Buscar user por tenant+email
   - Sempre executar bcrypt.Compare (com hash real ou dummy)
   - Se senha inválida ou user não existe → ErrInvalidCredentials
   - Se user.Status != "active" → ErrUserBlocked
   - Buscar subscription por tenant+product → validar status (suspended, canceled, ou ausente → erros específicos)
   - Buscar user_product_access → default "view" se não encontrar
   - Montar claims e chamar JWTService.Sign
   - Retornar LoginOutput{Token: token}
5. Constante dummyBcryptHash para timing normalization
6. Converter ErrTenantNotFound e ErrUserNotFound em ErrInvalidCredentials no fluxo de auth

---

### Checklist de Implementação

1. [ ] Criar LoginInput e LoginOutput
2. [ ] Implementar fluxo com ordem correta de validações
3. [ ] Usar dummy hash para timing quando user não existe
4. [ ] Tratar suspended e canceled com erros específicos
5. [ ] Montar claims JWT corretamente (iss, sub, aud, tenant_id, user_access_profile, jti, iat, exp, nbf)
6. [ ] Retornar ErrInvalidCredentials para tenant/user not found e wrong password
7. [ ] Teste unitário com mocks para cenários: credenciais inválidas, user blocked, product unavailable (inactive/blocked), subscription suspended, subscription canceled, sucesso

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-010 | Mensagem genérica, timing constante |
| ADR-006 | Claims JWT corretos |
| Code guidelines §12 | Nunca diferenciar auth failures |
| Clean Architecture | Use case orquestra, sem detalhes de infra |
| Especificação usuário | Validações: email, senha, status, assinatura |

---

## Referências

- [docs/adrs/ADR-006-jwt-com-assinatura-assimetrica.md](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [fase-3-servicos-crypto.md](./fase-3-servicos-crypto.md)
