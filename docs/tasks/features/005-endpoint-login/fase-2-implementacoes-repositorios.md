# Fase 2 — Implementações dos Repositórios GORM

## Objetivo

Implementar as interfaces de repositório definidas na Fase 1 usando GORM, garantindo queries parametrizadas, uso de contexto e mapeamento correto entre models e entidades de domínio.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O code guidelines §4 exige:
- Sempre usar `WithContext(ctx)` nas operações
- Nunca concatenar strings em queries (prevenir SQL injection)
- Usar placeholders `?` em Where

Os repositórios devem retornar `domain.ErrTenantNotFound`, `domain.ErrUserNotFound`, `domain.ErrProductNotFound` quando não encontrarem registros. O GORM retorna `gorm.ErrRecordNotFound` nesses casos.

### Ação 1.1

Padrão para tratamento de "not found":
```go
if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, domain.ErrTenantNotFound
}
return &entity, result.Error
```

### Observação 1.1

Usar `errors.Is` para compatibilidade com erros wrapped. O use case de auth converterá ErrTenantNotFound, ErrUserNotFound e ErrProductNotFound em ErrInvalidCredentials antes de retornar (ADR-010).

---

### Pensamento 2

O schema usa `wisa_crm_db`. O database.go configura `search_path` na conexão? Verificar como as queries são executadas. Se o search_path já inclui `wisa_crm_db`, os models podem referenciar tabelas sem prefixo de schema. Caso contrário, usar tabelas qualificadas.

### Ação 1.2

Verificar `internal/infrastructure/persistence/database.go`. Se usar `SET search_path = wisa_crm_db` na conexão ou no Open, as queries `db.Table("tenants")` ou `db.Model(&TenantModel{})` funcionarão. Os models devem ter `TableName()` retornando `"tenants"` (sem schema) se o search_path estiver configurado.

### Observação 1.2

O docs/backend e migrations indicam uso de search_path. A implementação deve seguir o padrão já existente no projeto.

---

### Pensamento 3

`FindByTenantAndProduct`: retornar a subscription do tenant para o produto específico. Query: `WHERE tenant_id = ? AND product_id = ?`. A assinatura é por tenant + produto; não há "subscription mais recente" sem especificar o produto. O use case verifica o status (active, pending → ok; suspended, canceled → erro específico).

### Ação 1.3

Assinatura da implementação:
```go
func (r *gormSubscriptionRepository) FindByTenantAndProduct(ctx context.Context, tenantID, productID uuid.UUID) (*entity.Subscription, error)
```

Query: `WHERE tenant_id = ? AND product_id = ? LIMIT 1`

Se não encontrar, retornar `nil, domain.ErrSubscriptionExpired` (ou similar). O use case verifica status quando encontrada e retorna ErrSubscriptionSuspended ou ErrSubscriptionCanceled conforme o caso.

### Observação 1.3

A assinatura é específica por tenant + produto. O product_id vem do Product encontrado por product_slug. Não há ambiguidade: uma subscription por (tenant_id, product_id) na tabela.

---

### Pensamento 4

UserProductAccess: FindByUserAndProduct. Se o usuário não tiver registro em user_product_access para o produto, o que fazer? O schema indica UNIQUE(user_id, product_id). Um usuário pode não ter acesso explícito a um produto. Para login, se não houver registro, usar perfil padrão `view` ou retornar erro? A especificação não cobre esse caso. Adotar: se não houver registro, retornar erro `ErrUserNoProductAccess` ou usar default `view`. Para MVP, assumir que usuários com acesso ao tenant têm pelo menos um user_product_access para o produto da assinatura. Se não tiver, usar `view` como default.

### Ação 1.5

`FindByUserAndProduct(ctx, userID, productID) (*entity.UserProductAccess, error)`. Se não encontrar, retornar `nil, domain.ErrUserNotFound` ou criar `ErrUserNoProductAccess`. Para simplificar: retornar (nil, ErrUserNoProductAccess). O use case pode tratar com default "view" ou retornar erro. Documentar na Fase 4 a decisão: usar "view" como default quando não houver registro.

### Observação 1.5

Decisão: quando não houver user_product_access, o use case usará "view" como access_profile default. O repositório retorna `nil, domain.ErrUserNoProductAccess` (opcional) ou podemos ter `FindByUserAndProduct` retornar (nil, nil) e o use case trata. Melhor: retornar (nil, ErrRecordNotFound) e no use case, se err != nil, usar "view" como default. Ou o repositório retorna um struct com AccessProfile="view" quando não encontrar? Isso mistura lógica. Manter: repositório retorna (nil, err) quando não encontrar. Use case: se err != nil, usar "view" como default. Assim não precisamos de ErrUserNoProductAccess.

---

### Pensamento 5

Conversão model → entity: cada repositório deve mapear o model GORM para a entidade de domínio. Campos como CreatedAt, UpdatedAt podem ser ignorados na entity se não forem necessários no use case.

### Ação 1.6

Criar funções de conversão privadas no package persistence, ex: `toEntity(m *TenantModel) *entity.Tenant`. Manter mapeamento explícito e simples.

---

### Decisão final Fase 2

**Implementar:**

1. `infrastructure/persistence/gorm_tenant_repository.go` — FindBySlug
2. `infrastructure/persistence/gorm_product_repository.go` — FindBySlug
3. `infrastructure/persistence/gorm_user_repository.go` — FindByEmailAndTenantID
4. `infrastructure/persistence/gorm_subscription_repository.go` — FindByTenantAndProduct
5. `infrastructure/persistence/gorm_user_product_access_repository.go` — FindByUserAndProduct (retorna nil, err quando não encontrar; use case usa "view" como default)
6. Todos usando `db.WithContext(ctx)`, queries parametrizadas, mapeamento para entidades

---

### Checklist de Implementação

1. [ ] Implementar GormTenantRepository
2. [ ] Implementar GormProductRepository
3. [ ] Implementar GormUserRepository
4. [ ] Implementar GormSubscriptionRepository (FindByTenantAndProduct)
5. [ ] Implementar GormUserProductAccessRepository
6. [ ] Garantir uso de WithContext e queries parametrizadas
7. [ ] Mapear gorm.ErrRecordNotFound para domain errors apropriados

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Code guidelines §4.1 | WithContext em todas as queries |
| Code guidelines §4.2 | Queries parametrizadas |
| ADR-003/004 | Uso de GORM com schema wisa_crm_db |
| Clean Architecture | Persistence implementa interfaces do domain |
| Tenant isolation | Filtro por tenant_id em todas as queries |

---

## Referências

- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/adrs/ADR-004-gorm-como-orm.md](../../adrs/ADR-004-gorm-como-orm.md)
- [fase-1-entidades-repositorios-models.md](./fase-1-entidades-repositorios-models.md)
