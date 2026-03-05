# Fase 1 — Entidades, Interfaces de Repositório e Models de Persistência

## Objetivo

Criar entidades de domínio puras, interfaces de repositório e models de persistência GORM necessários para o endpoint de login, em conformidade com Clean Architecture e ADR-005. O claim `aud` do JWT será construído a partir do `slug` + domínio base (config), sem necessidade de novos campos no banco. A URL de acesso inclui `product_slug`: `https://slug.domain.com.br/<product_slug>`.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O endpoint de login recebe `slug`, `product_slug`, `user_email` e `password`. A URL de acesso segue o padrão `https://slug.domain.com.br/<product_slug>`. O fluxo requer:
1. **Tenant** — buscar por slug para obter tenant_id (o aud do JWT será construído com slug + domínio base do config)
2. **Product** — buscar por product_slug para obter product_id; validar se status é `active` (se `inactive` ou `blocked`, retornar erro)
3. **User** — buscar por tenant_id + email para validar credenciais e status
4. **Subscription** — validar que o tenant possui assinatura **para o produto informado** com status `active` ou `pending` (tenant_id + product_id)
5. **UserProductAccess** — obter o perfil de acesso do usuário para o produto

O domínio não conhece GORM (code guidelines §3.1). As entidades devem ser structs puras. As interfaces de repositório vivem em `domain/repository`. Os models GORM ficam em `infrastructure/persistence` (ou em um package de models que a persistence usa).

### Ação 1.1

Definir entidades em `internal/domain/entity/`:
- `tenant.go`: Tenant com ID, Slug, Name, Status
- `product.go`: Product com ID, Slug, Name, Status (para lookup por product_slug)
- `user.go`: User com ID, TenantID, Email, PasswordHash, Status
- `subscription.go`: Subscription com ID, TenantID, ProductID, Status
- `user_product_access.go`: UserProductAccess com UserID, ProductID, TenantID, AccessProfile

### Observação 1.1

O claim `aud` do JWT será construído no use case como `slug` + domínio base (via `JWT_AUD_BASE_DOMAIN`). O login é chamado via redirect do sistema do cliente que já envia o `slug`; não é necessário armazenar subdomínio no banco.

---

### Pensamento 2

O schema do banco (migrations 000002–000007) define:
- `wisa_crm_db.tenants`: slug, name, tax_id, type, status
- `wisa_crm_db.products`: slug, name, status (active, inactive, blocked) — **login deve rejeitar se status != active**
- `wisa_crm_db.users`: tenant_id, email, password_hash, status (user_status: active, blocked)
- `wisa_crm_db.subscriptions`: tenant_id, product_id, status (pending, active, suspended, canceled) — **a assinatura é por tenant + produto**
- `wisa_crm_db.user_product_access`: user_id, product_id, tenant_id, access_profile (admin, operator, view)

O `access_profile` considera os valores: `admin`, `operator`, `view` (conforme enum do banco). O claim `user_access_profile` no JWT usará esses valores.

**Nenhuma migration adicional** é necessária para o login. O schema atual de tenants já contém o `slug`, suficiente para construir o `aud` do JWT.

---

### Pensamento 3

As entidades de domínio são puras e não usam tags GORM. Os tipos de status e profile podem ser strings no domínio. A persistência terá structs separadas com tags GORM que mapeiam para as tabelas. O repositório converte model → entity ao retornar.

Alternativa: usar um único struct com tags GORM na camada de persistência e mapear manualmente para a entidade no repositório. Isso evita duplicação de structs similares.

### Ação 1.3

Definir entidades com tipos mínimos:
```go
// domain/entity/tenant.go
type Tenant struct {
    ID     uuid.UUID
    Slug   string
    Name   string
    Status string  // active, inactive, blocked
}

// domain/entity/product.go
type Product struct {
    ID     uuid.UUID
    Slug   string
    Name   string
    Status string  // active, inactive, blocked
}

// domain/entity/user.go
type User struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    Email        string
    PasswordHash string
    Status       string  // active, blocked
}

// domain/entity/subscription.go
type Subscription struct {
    ID        uuid.UUID
    TenantID  uuid.UUID
    ProductID uuid.UUID
    Status    string  // pending, active, suspended, canceled
}

// domain/entity/user_product_access.go
type UserProductAccess struct {
    UserID        uuid.UUID
    ProductID     uuid.UUID
    TenantID      uuid.UUID
    AccessProfile string  // admin, operator, view
}
```

### Observação 1.3

Sem tags GORM nas entidades; o domínio permanece independente de persistência. UUID usa `github.com/google/uuid`.

---

### Pensamento 4

As interfaces de repositório (portas) devem expor apenas os métodos necessários para o use case de login. Evitar over-engineering.

### Ação 1.4

Definir interfaces em `domain/repository/`:
```go
// tenant_repository.go
type TenantRepository interface {
    FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
}

// product_repository.go
type ProductRepository interface {
    FindBySlug(ctx context.Context, slug string) (*entity.Product, error)
}

// user_repository.go
type UserRepository interface {
    FindByEmailAndTenantID(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error)
}

// subscription_repository.go
type SubscriptionRepository interface {
    FindByTenantAndProduct(ctx context.Context, tenantID, productID uuid.UUID) (*entity.Subscription, error)
}

// user_product_access_repository.go
type UserProductAccessRepository interface {
    FindByUserAndProduct(ctx context.Context, userID, productID uuid.UUID) (*entity.UserProductAccess, error)
}
```

`FindByTenantAndProduct` retorna a subscription do tenant para o produto específico. Cada tenant pode ter assinaturas para múltiplos produtos; a validação deve considerar tenant_id + product_id. O use case valida o status (active, pending → ok; suspended, canceled → erro específico).

### Observação 1.4

O use case precisará da subscription do tenant para o produto informado no login. A assinatura é obtida considerando tenant_id e product_id (product_id vem do Product encontrado por product_slug).

---

### Pensamento 5

Os models GORM precisam refletir exatamente o schema do banco, incluindo schema `wisa_crm_db` e tipos ENUM. O GORM usa `TableName()` para especificar o schema. Verificar ADR-004 e estrutura atual do database.go.

### Ação 1.5

Criar models em `infrastructure/persistence/model/` (ou no mesmo package do repository):
- `tenant_model.go`: struct com tags GORM para wisa_crm_db.tenants
- `product_model.go`: struct para wisa_crm_db.products
- `user_model.go`: struct para wisa_crm_db.users
- `subscription_model.go`: struct para wisa_crm_db.subscriptions
- `user_product_access_model.go`: struct para wisa_crm_db.user_product_access

Usar `gorm:"column:...;type:..."` conforme necessário. Para ENUMs, usar string no Go. Configurar `TableName()` retornando `"wisa_crm_db.tenants"` etc. ou usar `Scopes` com search_path.

### Observação 1.5

O database.go já configura `search_path` para `wisa_crm_db` conforme migrations. Os models podem usar apenas o nome da tabela (ex: `tenants`) se o search_path estiver correto na conexão.

---

### Pensamento 6

Erros de domínio adicionais: para usuário com status `blocked`, precisamos de `ErrUserBlocked`. Para produto não encontrado (product_slug inválido), `ErrProductNotFound` — no contexto de login, mapear para ErrInvalidCredentials (ADR-010). Para produto inativo ou bloqueado (`inactive`, `blocked`), `ErrProductUnavailable` — mensagem explícita pois o produto existe mas não está disponível. Verificar se já existem em domain/errors.go.

### Ação 1.6

Adicionar em `domain/errors.go`:
```go
ErrUserBlocked        = errors.New("user blocked")
ErrProductNotFound    = errors.New("product not found")
ErrProductUnavailable = errors.New("product unavailable")
```

Adicionar em `pkg/errors/codes.go`:
- `CodeUserBlocked = "USER_BLOCKED"`
- Função `NewUserBlocked()` com message "Usuário sem permissão para acessar o sistema." e HTTP 403
- `CodeProductUnavailable = "PRODUCT_UNAVAILABLE"`
- Função `NewProductUnavailable()` com message "Produto indisponível para acesso." e HTTP 403

Adicionar no `mapper.go`: `ErrUserBlocked` → `NewUserBlocked()`; `ErrProductUnavailable` → `NewProductUnavailable()`.

Atualizar `NewSubscriptionSuspended()` e `NewSubscriptionCanceled()` para incluir os `details` conforme especificação do usuário:
- SUBSCRIPTION_SUSPENDED: details = "Sua assinatura não está ativa devido a pagamentos em aberto. Por favor, atualize sua forma de pagamento para acessar o software."
- SUBSCRIPTION_CANCELED: details = "Sua assinatura foi cancelada. Entre em contato com a equipe Wisa Labs para analisar o caso."

### Observação 1.6

A Fase 1 inclui esses ajustes no package de erro pois são pré-requisitos para o fluxo de login. Mantém a feature autocontida.

---

### Decisão final Fase 1

**Implementar:**

1. Criar entidades em `domain/entity/`: Tenant, Product, User, Subscription, UserProductAccess
2. Criar interfaces em `domain/repository/`: TenantRepository, ProductRepository, UserRepository, SubscriptionRepository, UserProductAccessRepository
3. Criar models GORM em `infrastructure/persistence/model/` (ou equivalente)
4. Adicionar `ErrUserBlocked`, `ErrProductNotFound` e `ErrProductUnavailable` no domain e no catálogo/codes; ajustar `NewSubscriptionSuspended` e `NewSubscriptionCanceled` com details; mapear no ErrorMapper (ErrProductNotFound → ErrInvalidCredentials; ErrProductUnavailable → NewProductUnavailable)

---

### Checklist de Implementação

1. [ ] Criar `domain/entity/tenant.go`, `product.go`, `user.go`, `subscription.go`, `user_product_access.go`
2. [ ] Criar `domain/repository/tenant_repository.go`, `product_repository.go`, `user_repository.go`, `subscription_repository.go`, `user_product_access_repository.go`
3. [ ] Criar models GORM para persistence (respeitando schema wisa_crm_db)
4. [ ] Adicionar ErrUserBlocked, ErrProductNotFound, ErrProductUnavailable e ajustes em pkg/errors e mapper

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Clean Architecture | Entidades puras, sem GORM no domínio |
| Dependency Rule | domain não importa infrastructure |
| ADR-005 | Interfaces no domínio, models na persistence |
| ADR-008 | Estrutura compatível com row-level tenancy |
| Schema wisa_crm_db | Models usando schema correto |

---

## Referências

- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../adrs/ADR-005-clean-architecture.md)
- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../adrs/ADR-008-arquitetura-multi-tenant.md)
- [backend/migrations/](../../../backend/migrations/)
