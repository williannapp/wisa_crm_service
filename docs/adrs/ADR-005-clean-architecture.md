# ADR-005 — Clean Architecture como Padrão Arquitetural

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (backend)

---

## Contexto

O `wisa-crm-service` é um sistema de segurança crítica com expectativa de vida longa (sistema SaaS que evolui ao longo de anos). Suas regras de negócio — como validação de assinatura, emissão de JWT, bloqueio de tenant — são o núcleo do produto e devem ser:

- **Testáveis** de forma independente de frameworks, banco de dados e HTTP
- **Isoladas** das mudanças de tecnologia (trocar GORM por sqlx, ou REST por gRPC, não deve exigir reescrita das regras de negócio)
- **Compreensíveis** por novos desenvolvedores sem necessidade de entender toda a pilha tecnológica
- **Seguras** contra acoplamento acidental que exponha detalhes de implementação entre camadas

Sem uma arquitetura clara, sistemas de segurança tendem a crescer de forma desordenada, misturando lógica de negócio com lógica de infraestrutura — o que torna bugs de segurança mais difíceis de detectar e o código mais difícil de auditar.

A escolha do padrão arquitetural impacta diretamente:
- Testabilidade e cobertura de testes
- Capacidade de auditar a lógica de segurança de forma isolada
- Velocidade de onboarding de novos desenvolvedores
- Facilidade de troca de componentes de infraestrutura
- Manutenibilidade a longo prazo

---

## Decisão

**Clean Architecture (Robert C. Martin) é o padrão arquitetural adotado para o backend do `wisa-crm-service`.**

A implementação seguirá a **Dependency Rule**: dependências de código só podem apontar para dentro (em direção ao domínio), nunca para fora. O domínio não conhece o banco de dados, o framework HTTP ou qualquer detalhe de infraestrutura.

---

## Justificativa

### 1. Isolamento da lógica de segurança

Em um sistema de auth, as regras de negócio são o ativo mais crítico. A Clean Architecture garante que a lógica de validação de assinatura, emissão de JWT e controle de acesso por tenant viva em uma camada de domínio **completamente independente**:

- Não importa GORM, não importa `net/http`, não importa nenhuma biblioteca externa
- Pode ser auditada por um especialista de segurança sem conhecimento do stack tecnológico
- Pode ser testada com 100% de cobertura usando apenas mocks simples

### 2. Estrutura de camadas explícita

```
┌─────────────────────────────────────────────────────────┐
│                   Delivery Layer                        │
│  (HTTP Handlers, Middleware, Request/Response DTOs)     │
├─────────────────────────────────────────────────────────┤
│                   Use Case Layer                        │
│  (AuthenticateUser, ValidateSubscription, IssueJWT)     │
├─────────────────────────────────────────────────────────┤
│                   Domain Layer                          │
│  (Entities: User, Tenant, Subscription, Token)          │
│  (Interfaces: UserRepository, SubscriptionRepository)   │
├─────────────────────────────────────────────────────────┤
│                Infrastructure Layer                     │
│  (GORM Repositories, JWT Service impl, bcrypt, Email)   │
└─────────────────────────────────────────────────────────┘

Regra: as setas de dependência SEMPRE apontam para CIMA (em direção ao domínio)
```

### 3. Testabilidade por design

Cada caso de uso pode ser testado com mocks da interface de repositório:

```go
// Domínio define a interface — não conhece GORM
type UserRepository interface {
    FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
}

// Teste do Use Case com mock — sem banco real
func TestAuthenticateUser_InvalidPassword(t *testing.T) {
    mockRepo := &MockUserRepository{
        FindByEmailFn: func(ctx context.Context, tenantID uuid.UUID, email string) (*User, error) {
            return &User{PasswordHash: "$2a$12$hashedvalue"}, nil
        },
    }
    
    useCase := NewAuthenticateUserUseCase(mockRepo, ...)
    _, err := useCase.Execute(ctx, AuthInput{Email: "user@test.com", Password: "wrong"})
    
    assert.ErrorIs(t, err, ErrInvalidCredentials)
}
```

Testes unitários de lógica de segurança executam em milissegundos, sem banco de dados.

### 4. Estrutura de diretórios recomendada

```
wisa-crm-service/
├── cmd/
│   └── api/
│       └── main.go                    # Entrypoint — wires dependencies
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── user.go                # User entity + business rules
│   │   │   ├── tenant.go              # Tenant entity
│   │   │   └── subscription.go        # Subscription entity + validation logic
│   │   ├── repository/
│   │   │   ├── user_repository.go     # Interface
│   │   │   ├── tenant_repository.go   # Interface
│   │   │   └── subscription_repository.go
│   │   └── errors.go                  # Domain errors (ErrInvalidCredentials, etc.)
│   ├── usecase/
│   │   ├── auth/
│   │   │   ├── authenticate_user.go   # Use Case: login flow
│   │   │   └── authenticate_user_test.go
│   │   ├── subscription/
│   │   │   ├── validate_subscription.go
│   │   │   └── validate_subscription_test.go
│   │   └── token/
│   │       ├── issue_jwt.go           # Use Case: JWT emission
│   │       └── issue_jwt_test.go
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── gorm_user_repository.go      # GORM impl of UserRepository
│   │   │   └── gorm_subscription_repository.go
│   │   ├── crypto/
│   │   │   ├── bcrypt_password_service.go   # bcrypt impl
│   │   │   └── rsa_jwt_service.go           # JWT signing impl
│   │   └── http/
│   │       ├── middleware/
│   │       │   ├── rate_limiter.go
│   │       │   └── tenant_extractor.go
│   │       └── server.go
│   └── delivery/
│       └── http/
│           ├── handler/
│           │   ├── auth_handler.go    # HTTP handlers — call use cases
│           │   └── health_handler.go
│           └── dto/
│               ├── login_request.go   # Input DTOs com validação
│               └── token_response.go  # Output DTOs
└── pkg/
    └── logger/
        └── logger.go                  # Shared utilities
```

### 5. Dependency Injection no `main.go`

O `main.go` é o único lugar onde todas as camadas se conhecem — é onde as dependências são construídas e injetadas:

```go
// cmd/api/main.go — simplified
func main() {
    // Infrastructure
    db := infrastructure.NewDatabase(cfg.DatabaseURL)
    userRepo := persistence.NewGORMUserRepository(db)
    subscriptionRepo := persistence.NewGORMSubscriptionRepository(db)
    jwtService := crypto.NewRSAJWTService(cfg.PrivateKeyPath)
    passwordService := crypto.NewBcryptPasswordService(cfg.BcryptCost)

    // Use Cases
    authenticateUser := auth.NewAuthenticateUserUseCase(userRepo, subscriptionRepo, jwtService, passwordService)

    // Delivery
    authHandler := handler.NewAuthHandler(authenticateUser)
    
    // Server
    server := http.NewServer(authHandler)
    server.Start(cfg.Port)
}
```

Nenhum use case ou entidade de domínio precisa de um framework de DI — a injeção é manual e explícita.

### 6. Domain Errors para comunicação entre camadas

O domínio define seus próprios erros semânticos, que as camadas externas mapeiam para respostas HTTP:

```go
// domain/errors.go
var (
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrTenantNotFound     = errors.New("tenant not found")
    ErrSubscriptionExpired = errors.New("subscription expired")
    ErrAccountLocked      = errors.New("account locked")
)
```

O handler HTTP mapeia esses erros para status codes HTTP sem expor detalhes internos ao usuário.

---

## Consequências

### Positivas
- Lógica de segurança auditável de forma isolada, sem dependência de frameworks
- Testes unitários rápidos e completos da lógica de negócio
- Troca de componentes de infraestrutura sem impacto no domínio
- Código organizado e compreensível para novos membros da equipe
- Erros de domínio semânticos facilitam tratamento consistente de falhas de autenticação

### Negativas
- Mais arquivos e interfaces comparado a uma abordagem MVC simples
- Curva de aprendizado para desenvolvedores não familiarizados com Clean Architecture
- Risco de over-engineering para features simples — requer disciplina para não criar abstrações desnecessárias
- Injeção de dependência manual no `main.go` pode crescer em complexidade com muitos use cases

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| Violação da Dependency Rule (domínio importando infra) | Média | Alto | Alta |
| Over-engineering de entidades simples com interfaces desnecessárias | Alta | Baixo | Baixa |
| `main.go` crescendo de forma descontrolada (God Object) | Média | Médio | Média |
| Inconsistência no mapeamento de erros de domínio para HTTP | Média | Médio | Média |

---

## Mitigações

### Violação da Dependency Rule
- Configurar `depguard` ou `go-cleanarch` linter para detectar imports proibidos entre camadas
- Regra de code review: qualquer import de `infrastructure/` em `domain/` ou `usecase/` é bloqueante
- `go vet` e análise estática na CI

### Over-engineering
- Regra pragmática: só criar interface quando há pelo menos 2 implementações concretas **ou** necessidade de mock em testes
- Entidades simples podem ser structs sem comportamento — nem tudo precisa de interface

### `main.go` complexo
- Extrair funções de wiring para `internal/app/wire.go` se crescer muito
- Considerar `wire` (Google) como gerador de código de DI se o número de dependências justificar

### Mapeamento de erros
- Criar um `ErrorMapper` centralizado na camada de delivery que converta todos os domain errors em respostas HTTP padronizadas
- Documentar o mapeamento explicitamente (ex: `ErrInvalidCredentials` → HTTP 401 com corpo `{"code": "INVALID_CREDENTIALS"}`)

---

## Alternativas Consideradas

### MVC (Model-View-Controller)
- **Prós:** Simples, familiar, menos arquivos
- **Contras:** Tende a misturar lógica de negócio com lógica de infraestrutura (fat controllers, fat models), dificulta testes unitários da lógica de segurança, acoplamento com framework HTTP

### Hexagonal Architecture (Ports and Adapters)
- **Prós:** Filosoficamente semelhante à Clean Architecture, boa separação de concerns
- **Contras:** Nomenclatura diferente pode confundir (ports vs repositories, adapters vs infrastructure), mesmos benefícios da Clean Architecture com terminologia menos padronizada na comunidade Go

### Modular Monolith sem arquitetura explícita
- **Prós:** Desenvolvimento mais rápido no curto prazo
- **Contras:** Sem regras explícitas de dependência, a lógica de segurança inevitavelmente se mistura com infraestrutura ao longo do tempo, tornando auditorias e testes progressivamente mais difíceis

**Clean Architecture é a escolha ideal para um sistema de segurança crítica com expectativa de vida longa em um contexto SaaS em crescimento.**

---

## Referências

- [Clean Architecture — Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Practical Go Clean Architecture](https://github.com/evrone/go-clean-template)
- [go-cleanarch linter](https://github.com/roblaszczak/go-cleanarch)
- [Wire — Dependency Injection for Go](https://github.com/google/wire)
