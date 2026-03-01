# Code Guidelines — Backend Go (Clean Architecture + GORM)

## 1. Estrutura de Projeto em Camadas

### 1.1 Layout Recomendado (Clean Architecture)

```
wisa-crm-service/
├── cmd/
│   └── api/
│       └── main.go                 # Entrypoint — wiring de dependências
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── user.go
│   │   │   ├── tenant.go
│   │   │   └── subscription.go
│   │   ├── repository/
│   │   │   ├── user_repository.go      # Interface
│   │   │   └── subscription_repository.go
│   │   └── errors.go                   # Erros de domínio
│   ├── usecase/
│   │   ├── auth/
│   │   │   ├── authenticate_user.go
│   │   │   └── authenticate_user_test.go
│   │   └── subscription/
│   │       └── validate_subscription.go
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── gorm_user_repository.go
│   │   │   └── gorm_subscription_repository.go
│   │   ├── crypto/
│   │   │   ├── bcrypt_password_service.go
│   │   │   └── rsa_jwt_service.go
│   │   └── http/
│   │       └── middleware/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/
│   │       └── dto/
│   └── app/
│       └── wire.go                  # Wiring de dependências (se necessário)
├── pkg/
│   └── logger/
│       └── logger.go
├── migrations/
│   ├── 000001_create_tenants.up.sql
│   ├── 000001_create_tenants.down.sql
│   └── ...
├── go.mod
└── go.sum
```

### 1.2 Regra de Dependência (Dependency Rule)

- Dependências só podem apontar **para dentro** (em direção ao domínio).
- O domínio **não importa** GORM, `net/http` ou qualquer biblioteca de infraestrutura.
- Use `depguard` ou `go-cleanarch` na CI para validar imports entre camadas.

---

## 2. Organização de Pacotes

| Pacote | Responsabilidade |
|--------|------------------|
| `domain/entity` | Entidades puras e regras de negócio |
| `domain/repository` | Interfaces de repositório (portas) |
| `usecase` | Casos de uso — orquestram entidades e repositórios |
| `infrastructure/persistence` | Implementações GORM dos repositórios |
| `delivery/http` | Handlers HTTP, middlewares, DTOs |
| `pkg` | Utilitários compartilhados (logger, etc.) |

---

## 3. Entidades, Use Cases e Repositories

### 3.1 Entidades de Domínio

- Structs puras em `domain/entity/`.
- Sem tags GORM — domínio não conhece persistência.
- Regras de negócio encapsuladas em métodos das entidades quando fizer sentido.

### 3.2 Interfaces de Repository

```go
// domain/repository/user_repository.go
package repository

type UserRepository interface {
    FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error)
    Create(ctx context.Context, user *entity.User) error
}
```

- Interface definida no domínio.
- Use Case depende da interface, não da implementação.

### 3.3 Implementação com GORM

```go
// infrastructure/persistence/gorm_user_repository.go
package persistence

type gormUserRepository struct {
    db *gorm.DB
}

func (r *gormUserRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error) {
    var user entity.User
    result := r.db.WithContext(ctx).Where("tenant_id = ? AND email = ?", tenantID, email).First(&user)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, domain.ErrUserNotFound
    }
    return &user, result.Error
}
```

---

## 4. Boas Práticas com GORM

### 4.1 Sempre Usar Contexto

```go
// CORRETO
r.db.WithContext(ctx).Where(...).Find(...)

// INCORRETO — conexão pode ficar pendurada sem timeout
r.db.Where(...).Find(...)
```

### 4.2 Nunca Concatenar Strings em Queries

```go
// CORRETO — parametrizado
db.Where("email = ?", email)

// INCORRETO — SQL Injection
db.Where(fmt.Sprintf("email = '%s'", email))
```

### 4.3 Convenções de Modelos GORM

- Campo `ID` como chave primária por padrão.
- `CreatedAt` e `UpdatedAt` para timestamps automáticos.
- Nomes de struct em snake_case plural para tabelas (ex.: `User` → `users`).
- Tags `gorm:"primaryKey"` para chaves customizadas.

```go
type User struct {
    gorm.Model
    TenantID  uuid.UUID `gorm:"type:uuid;index"`
    Email     string    `gorm:"uniqueIndex:idx_tenant_email"`
}
```

### 4.4 Associações e Preload

- Use `Preload()` explicitamente para evitar N+1.
- Prefira `Joins()` para joins de alta frequência com múltiplos registros.

### 4.5 Transações

```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&Animal{Name: "Giraffe"}).Error; err != nil {
        return err // rollback automático
    }
    if err := tx.Create(&Animal{Name: "Lion"}).Error; err != nil {
        return err
    }
    return nil // commit
})
```

### 4.6 Hooks com Assinatura Correta (GORM v2)

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    tx.Statement.Select("Name", "Age")
    return nil
}
```

- Assinatura obrigatória: `func(tx *gorm.DB) error`.

---

## 5. Padrões de Migração

### 5.1 Migrations Versionadas

- **Não** use `AutoMigrate` em produção.
- Use `golang-migrate` com arquivos `.up.sql` e `.down.sql`.
- Nomenclatura: `000001_descricao.up.sql`, `000001_descricao.down.sql`.

### 5.2 AutoMigrate Apenas em Desenvolvimento

```go
// Apenas para dev local
db.AutoMigrate(&User{}, &Tenant{}, &Subscription{})
```

---

## 6. Tratamento de Erros Idiomático em Go

### 6.1 Erros de Domínio

```go
// domain/errors.go
var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrTenantNotFound     = errors.New("tenant not found")
    ErrSubscriptionExpired = errors.New("subscription expired")
)
```

### 6.2 Wrapping e Verificação

```go
if err != nil {
    return fmt.Errorf("find user: %w", err)
}

if errors.Is(err, domain.ErrUserNotFound) {
    return nil, ErrInvalidCredentials
}
```

### 6.3 Mapeamento de Erros para HTTP

- Centralizar mapeamento em um `ErrorMapper` na camada delivery.
- Documentar: `ErrInvalidCredentials` → HTTP 401, `ErrTenantNotFound` → HTTP 404.

---

## 7. Estratégia de Logs

| Nível | Uso |
|-------|-----|
| Debug | Apenas em desenvolvimento |
| Info | Operações significativas (login, criação de tenant) |
| Warn | Situações recuperáveis (tentativa de login inválida) |
| Error | Falhas que requerem atenção |
| Fatal | Apenas no startup (ex.: banco inacessível) |

- Não logar credenciais, tokens ou dados sensíveis.
- Incluir `request_id` ou `trace_id` para correlação.

---

## 8. Padrão para Middlewares

- Middlewares na camada `infrastructure/http/middleware/`.
- Assinatura: `func(next http.Handler) http.Handler` ou equivalente para o router usado.
- Responsabilidades típicas: rate limiting, tenant extraction, logging, recovery.

```go
func TenantExtractor(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := extractTenantFromHeader(r)
        ctx := context.WithValue(r.Context(), tenantKey, tenantID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## 9. Estrutura para Testes Unitários

### 9.1 Testes de Use Case com Mocks

```go
func TestAuthenticateUser_InvalidPassword(t *testing.T) {
    mockRepo := &MockUserRepository{
        FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email string) (*entity.User, error) {
            return &entity.User{PasswordHash: "$2a$12$hashedvalue"}, nil
        },
    }
    useCase := NewAuthenticateUserUseCase(mockRepo, ...)
    _, err := useCase.Execute(ctx, AuthInput{Email: "user@test.com", Password: "wrong"})
    assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}
```

### 9.2 Organização de Testes

- Arquivo `*_test.go` no mesmo pacote para testes internos.
- Pacote `_test` (ex.: `usecase_test`) para testes que dependem apenas da API pública.

---

## 10. Convenções de Nomenclatura

| Elemento | Convenção | Exemplo |
|----------|-----------|---------|
| Pacotes | lowercase, singular | `auth`, `entity`, `repository` |
| Interfaces | PascalCase, sufixo opcional | `UserRepository`, `TokenService` |
| Structs | PascalCase | `User`, `AuthenticateUserUseCase` |
| Funções exportadas | PascalCase | `NewUserRepository`, `FindByEmail` |
| Variáveis não exportadas | camelCase | `userRepo`, `tenantID` |
| Constantes | PascalCase ou SCREAMING | `MaxRetries`, `API_VERSION` |

---

## 11. Princípios de Clean Code Aplicados

- **Single Responsibility:** Uma responsabilidade por struct/função.
- **Interface segregation:** Interfaces pequenas e focadas.
- **Dependency Inversion:** Use cases dependem de interfaces, não de implementações.
- **Pragmatismo:** Criar interface apenas quando houver ≥2 implementações ou necessidade de mock.
- **Explicitude:** Preferir código explícito a "mágica" ou convenções obscuras.

---

## 12. Diretrizes de Segurança

| Diretriz | Descrição |
|----------|-----------|
| SQL Injection | Nunca concatenar strings em `db.Raw()` ou `db.Exec()`; usar placeholders `?` |
| Tenant isolation | Global Scope GORM para injetar `tenant_id` em todas as queries; testes de isolamento |
| Credenciais | Nunca diferenciar "usuário não encontrado" vs "senha incorreta" na resposta |
| Logging | Não logar passwords, tokens ou dados sensíveis |
| Context timeouts | Sempre propagar context com timeout em operações de I/O |
| Pool de conexões | Configurar `ConnMaxLifetime`, `ConnMaxIdleTime`, `MaxOpenConns` |

### Pool de Conexões Recomendado

```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
sqlDB.SetConnMaxIdleTime(2 * time.Minute)
```

---

## 13. Injeção de Dependência no main.go

- `main.go` é o único ponto onde todas as camadas se conectam.
- Construção manual e explícita de dependências.
- Se crescer demais, extrair para `internal/app/wire.go` ou considerar `wire` (Google).

```go
func main() {
    db := infrastructure.NewDatabase(cfg.DatabaseURL)
    userRepo := persistence.NewGORMUserRepository(db)
    authenticateUser := auth.NewAuthenticateUserUseCase(userRepo, ...)
    authHandler := handler.NewAuthHandler(authenticateUser)
    server := http.NewServer(authHandler)
    server.Start(cfg.Port)
}
```
