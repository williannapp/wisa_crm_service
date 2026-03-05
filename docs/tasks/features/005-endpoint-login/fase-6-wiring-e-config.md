# Fase 6 — Wiring e Configuração Final

## Objetivo

Conectar todas as dependências no main.go, configurar variáveis de ambiente e garantir que o endpoint de login esteja funcional com as configurações necessárias para JWT, banco de dados e demais serviços.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O main.go atual (feature 001) configura: database, router, health handler. Precisamos adicionar:
- Repositórios (Tenant, User, Subscription, UserProductAccess)
- PasswordService (BcryptPasswordService)
- JWTService (RSAJWTService)
- AuthenticateUserUseCase
- AuthHandler

A ordem de construção segue a Dependency Rule: infraestrutura primeiro, depois use cases, depois handlers.

### Ação 1.1

Sequência no main.go:
```go
// 1. Database (já existe)
db := persistence.NewDatabase(...)

// 2. Repositories
tenantRepo := persistence.NewGormTenantRepository(db)
productRepo := persistence.NewGormProductRepository(db)
userRepo := persistence.NewGormUserRepository(db)
subscriptionRepo := persistence.NewGormSubscriptionRepository(db)
userProductAccessRepo := persistence.NewGormUserProductAccessRepository(db)

// 3. Crypto services
passwordSvc := crypto.NewBcryptPasswordService()
jwtSvc := crypto.NewRSAJWTService(cfg.JWTPrivateKeyPath, cfg.JWTIssuer, cfg.JWTExpiration, cfg.JWTKeyID)

// 4. Use case
authenticateUser := auth.NewAuthenticateUserUseCase(tenantRepo, productRepo, userRepo, subscriptionRepo, userProductAccessRepo, passwordSvc, jwtSvc, cfg.JWTAudBaseDomain)

// 5. Handler
authHandler := handler.NewAuthHandler(authenticateUser)

// 6. Router - adicionar rota
authGroup := router.Group("/api/v1/auth")
authGroup.POST("/login", authHandler.Login)
```

### Observação 1.1

O cfg (config) precisa carregar as variáveis de env. Verificar estrutura atual do main.go para ver como PORT, DATABASE_URL são lidos. Usar godotenv e variáveis de ambiente.

---

### Pensamento 2

Variáveis de ambiente necessárias:
- `JWT_PRIVATE_KEY_PATH` — caminho para o arquivo .pem da chave privada RSA
- `JWT_ISSUER` — valor para claim iss (default: "wisa-crm-service")
- `JWT_EXPIRATION_MINUTES` — duração do token em minutos (default: 15)
- `JWT_KEY_ID` — kid no header (default: "key-2026-v1")
- `JWT_AUD_BASE_DOMAIN` — domínio base para compor o aud com o slug (ex: "app.wisa-crm.com"). aud = slug + "." + base_domain (ex: cliente1.app.wisa-crm.com)

### Ação 1.2

Atualizar `.env.example`:
```
# JWT (Login / Auth)
JWT_PRIVATE_KEY_PATH=/path/to/private.pem
JWT_ISSUER=wisa-crm-service
JWT_EXPIRATION_MINUTES=15
JWT_KEY_ID=key-2026-v1
JWT_AUD_BASE_DOMAIN=app.wisa-crm.com
```

Documentar que JWT_PRIVATE_KEY_PATH deve apontar para arquivo fora do repositório. Gerar chave com: `openssl genrsa -out private.pem 4096`

### Observação 1.2

O .gitignore deve incluir `*.pem`, `private*.pem` para evitar commit acidental da chave.

---

### Pensamento 3

O BcryptPasswordService não precisa de config. O RSAJWTService precisa de: path da chave, issuer, expiration em minutos, kid. O use case precisa apenas de audienceBaseDomain para compor o aud com o slug do tenant. aud = slug + "." + base_domain.

### Ação 1.3

Não há fallback: aud é sempre construído a partir do slug + JWT_AUD_BASE_DOMAIN. O slug vem no redirect do sistema do cliente.

### Observação 1.3

Simples e previsível: o tenant é identificado pelo slug, que também é o subdomínio do sistema do cliente.

---

### Pensamento 4

Estrutura do package de config: o projeto pode ter um package `config` ou ler env diretamente no main. Se não existir, criar `internal/config` ou ler no main com `os.Getenv`. Para manter simples, ler no main e passar valores para os construtores. Quando o número de configs crescer, extrair para `cmd/api/config.go` ou package dedicado.

### Ação 1.4

No main.go, após godotenv.Load():
```go
jwtPrivateKeyPath := getEnv("JWT_PRIVATE_KEY_PATH", "")
jwtIssuer := getEnv("JWT_ISSUER", "wisa-crm-service")
jwtExpiration := getEnvInt("JWT_EXPIRATION_MINUTES", 15)
jwtKeyID := getEnv("JWT_KEY_ID", "key-2026-v1")
jwtAudBaseDomain := getEnv("JWT_AUD_BASE_DOMAIN", "app.wisa-crm.com")
```

Criar helper `getEnv(key, default string)` e `getEnvInt(key, default int)` se não existirem.

### Observação 1.4

Se JWT_PRIVATE_KEY_PATH vazio, o RSAJWTService deve falhar no startup (fail fast) ou o login retornará erro ao tentar assinar. Preferir fail fast: o NewRSAJWTService retorna erro se não conseguir carregar a chave; o main loga e sai.

---

### Pensamento 5

O database já está inicializado. Os repositórios precisam do `*gorm.DB`. Verificar a assinatura de `persistence.NewDatabase` ou equivalente. O database.go retorna `*gorm.DB` ou uma struct com DB. Os repositórios GORM recebem `*gorm.DB`.

### Ação 1.5

Verificar `internal/infrastructure/persistence/database.go` para ver a API. Ajustar os construtores dos repositórios para receber o tipo correto.

### Observação 1.5

A estrutura pode ter `Database` struct com método `DB() *gorm.DB`. Os repositórios recebem `db.DB()` ou o próprio `*gorm.DB` passado no construtor.

---

### Pensamento 6

Teste de integração manual: após wiring, executar o backend, criar um tenant de teste com slug, um user com senha hasheada bcrypt, uma subscription ativa, um user_product_access. Fazer POST para /api/v1/auth/login e verificar resposta com token. Decodificar o JWT em jwt.io para validar claims.

### Ação 1.6

Documentar no README ou em docs/backend o fluxo de teste: seed de dados de exemplo, comando curl para login, validação do token.

### Observação 1.6

A Fase 6 inclui o wiring. O teste manual e seeds podem ser tasks opcionais ou documentados para implementação futura.

---

### Decisão final Fase 6

**Implementar:**

1. Atualizar `cmd/api/main.go` com construção de todos os repositórios, serviços e use case
2. Registrar rota POST /api/v1/auth/login no router
3. Carregar variáveis de ambiente para JWT (JWT_PRIVATE_KEY_PATH, JWT_ISSUER, JWT_EXPIRATION_MINUTES, JWT_KEY_ID, JWT_AUD_FALLBACK)
4. Atualizar `.env.example` com as novas variáveis e documentação
5. Garantir que RSAJWTService falhe no startup se a chave não for carregável
6. Helpers getEnv/getEnvInt se não existirem
7. Documentar em docs/backend ou README como testar o endpoint de login

---

### Checklist de Implementação

1. [ ] Construir repositórios no main
2. [ ] Construir PasswordService e JWTService
3. [ ] Construir AuthenticateUserUseCase com todas as dependências (incluindo ProductRepo)
4. [ ] Construir AuthHandler e registrar rota
5. [ ] Adicionar variáveis JWT no .env.example
6. [ ] Implementar carregamento de config
7. [ ] Fail fast se chave privada não carregar
8. [ ] Rodar migração 000008
9. [ ] Teste manual de login

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Code guidelines §13 | main.go como ponto de wiring |
| ADR-006 | Chave privada em arquivo, fora do código |
| Configurável | iss, exp, aud fallback via env |
| Segurança | .gitignore para *.pem |
| Operacional | Documentação de setup |

---

## Referências

- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/adrs/ADR-006-jwt-com-assinatura-assimetrica.md](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [cmd/api/main.go](../../../backend/cmd/api/main.go)
