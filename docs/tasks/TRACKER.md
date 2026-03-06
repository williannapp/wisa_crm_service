# Task Tracker

> **Este arquivo é o ponto central de controle do projeto.**  
> Consulte-o ao início de cada sessão de desenvolvimento para saber exatamente onde parou.

---

## Estrutura de Diretórios

O diretório `docs/tasks/` está organizado por **features** e **fixes**:

```
docs/tasks/
├── TRACKER.md              ← Você está aqui (visão geral)
├── README.md               ← Guia para criar novas features/fixes
├── features/               ← Features em desenvolvimento
│   ├── 001-estrutura-inicial-backend/
│   ├── 002-configuracao-banco-dados/
│   ├── 003-estrutura-tabelas-banco-dados/
│   ├── 004-package-erro-padronizado/
│   ├── 005-endpoint-login/
│   ├── 006-jwt-cookie-redirect-url/
│   ├── 007-auth-code-flow-redis/
│   ├── 008-refresh-token-endpoint/
│   ├── 009-public-key-endpoint/
│   ├── 010-estrutura-inicial-frontend/
│   ├── 011-tela-login/
│   └── 012-frontend-login-implementation/
└── fixes/                  ← Correções e bugs
    └── (vazio — para futuras correções)
```

---

## Legenda

- `[ ]` Pendente
- `[~]` Em andamento
- `[x]` Concluída
- `[-]` Cancelada

---

## Status Geral

| Feature/Fix | Descrição | Progresso | Status |
|-------------|-----------|-----------|--------|
| [001-estrutura-inicial-backend](features/001-estrutura-inicial-backend/TRACKER.md) | Estrutura inicial do backend: diretórios, libs, .gitignore, env, Dockerfile, health | 6/6 | Concluída |
| [002-configuracao-banco-dados](features/002-configuracao-banco-dados/TRACKER.md) | Configuração do banco: estrutura base, env, containers, documentação, ORM/migrations | 5/5 | Concluída |
| [003-estrutura-tabelas-banco-dados](features/003-estrutura-tabelas-banco-dados/TRACKER.md) | Estrutura de tabelas: schema wisa_crm_db, tenants, products, subscriptions, payments, users, user_product_access, refresh_tokens, audit_logs | 6/6 | Concluída |
| [004-package-erro-padronizado](features/004-package-erro-padronizado/TRACKER.md) | Package de erro padronizado: estrutura pkg/errors, AppError, catálogo de códigos, ErrorMapper, integração na delivery | 3/3 | Concluída |
| [005-endpoint-login](features/005-endpoint-login/TRACKER.md) | Endpoint POST /api/v1/auth/login: validações (tenant, product, email, senha, status, assinatura), emissão JWT RS256. URL: slug.domain.com.br/product_slug | 6/6 | Concluída |
| [006-jwt-cookie-redirect-url](features/006-jwt-cookie-redirect-url/TRACKER.md) | JWT como cookie HttpOnly + redirect_url. Supersedido pela 007 (ATA escolheu Authorization Code + Redis) | — | Cancelada |
| [007-auth-code-flow-redis](features/007-auth-code-flow-redis/TRACKER.md) | Authorization Code Flow: Redis para codes, login retorna 302, POST /auth/token troca code por JWT. TTL 40s. Cliente implementa GET /callback | 5/5 | Concluída |
| [008-refresh-token-endpoint](features/008-refresh-token-endpoint/TRACKER.md) | Refresh Token: migration product_id, refresh no token exchange, POST /auth/refresh, validação hash+tenant+product, rotação atômica, 7 dias | 4/4 | Concluída |
| [009-public-key-endpoint](features/009-public-key-endpoint/TRACKER.md) | Endpoint público GET /.well-known/jwks.json: JWKS Provider, chave pública RSA, sem autenticação, Cache-Control 24h, suporte a múltiplas chaves | 3/3 | Concluída |
| [010-estrutura-inicial-frontend](features/010-estrutura-inicial-frontend/TRACKER.md) | Estrutura inicial do frontend: diretórios Angular, bibliotecas, .gitignore, Dockerfile, serviço no docker-compose | 5/5 | Concluída |
| [011-tela-login](features/011-tela-login/TRACKER.md) | Tela de login: design baseado no protótipo Login-Wisa. Design apenas, sem lógica de auth | 5/5 | Concluída |
| [012-frontend-login-implementation](features/012-frontend-login-implementation/TRACKER.md) | Implementação do login no frontend: parâmetros de query (tenant_slug, product_slug, state), validação, POST /auth/login, redirect | 3/3 | Concluída |

---

## Diário de Sessões

*(Registre aqui as atividades significativas de cada sessão de desenvolvimento)*

### Sessão 1 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 001 — Estrutura inicial do backend
  - Criação de documentos de planejamento para as 6 fases
- **Features/fixes criados:** 001-estrutura-inicial-backend (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-estrutura-diretorios.md](features/001-estrutura-inicial-backend/fase-1-estrutura-diretorios.md)

### Sessão 2 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 001 — Estrutura inicial do backend
  - Fase 1: Estrutura de diretórios em `backend/` (Clean Architecture)
  - Fase 2: go.mod com Gin v1.9.1 e godotenv v1.5.1
  - Fase 3: .gitignore na raiz e em backend/
  - Fase 4: .env.example com PORT e APP_ENV
  - Fase 5: Dockerfile multi-stage (golang:1.25-alpine, alpine:3.19)
  - Fase 6: Endpoint GET /health com handler Gin
- **Features/fixes concluídos:** 001-estrutura-inicial-backend
- **Tasks concluídas:** 6/6 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

### Sessão 3 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 002 — Configuração do banco de dados
  - Criação de documentos de planejamento para as 5 fases
- **Features/fixes criados:** 002-configuracao-banco-dados (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-estrutura-base-banco.md](features/002-configuracao-banco-dados/fase-1-estrutura-base-banco.md)

### Sessão 4 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 002 — Configuração do banco de dados
  - Fase 1: persistence/database.go com GORM, pool configurado; main.go com DATABASE_URL
  - Fase 2: .env.example com DATABASE_URL documentada
  - Fase 3: backend/docker/database/Dockerfile, backend/docker/backend/Dockerfile, docker-compose.yml (postgres + backend)
  - Fase 5: migrations 000001_init_schema, cmd/migrate, Makefile
  - Fase 4: docs/backend/README.md, docs/backend/vps-configurations.md
- **Features/fixes concluídos:** 002-configuracao-banco-dados
- **Tasks concluídas:** 5/5 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

### Sessão 5 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 003 — Estrutura de tabelas do banco de dados
  - Criação de documentos de planejamento para as 6 fases (schema, tabelas, índices, RLS, triggers)
- **Features/fixes criados:** 003-estrutura-tabelas-banco-dados (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-schema-e-enums.md](features/003-estrutura-tabelas-banco-dados/fase-1-schema-e-enums.md)

### Sessão 6 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 003 — Estrutura de tabelas do banco de dados
  - Fase 1: Migration 000002 — schema wisa_crm_db e 8 tipos ENUM
  - Fase 2: Migration 000003 — tabelas tenants e products
  - Fase 3: Migration 000004 — tabelas subscriptions e payments
  - Fase 4: Migration 000005 — tabelas users e user_product_access
  - Fase 5: Migration 000006 — tabelas refresh_tokens e audit_logs (particionamento)
  - Fase 6: Migration 000007 — índices, RLS e triggers set_updated_at
  - Atualização de backend/.env.example e docs/backend/vps-configurations.md com search_path
- **Features/fixes concluídos:** 003-estrutura-tabelas-banco-dados
- **Tasks concluídas:** 6/6 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

### Sessão 7 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 004 — Package de Erro Padronizado
  - Criação de documentos de planejamento para as 3 fases (estrutura + AppError, catálogo + mapper, integração delivery)
- **Features/fixes criados:** 004-package-erro-padronizado (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-estrutura-diretorios-tipo-erro.md](features/004-package-erro-padronizado/fase-1-estrutura-diretorios-tipo-erro.md)

### Sessão 8 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 004 — Package de Erro Padronizado
  - Fase 1: pkg/errors com AppError, NewAppError, codes.go, MarshalJSON
  - Fase 2: Catálogo completo (INVALID_CREDENTIALS, ACCOUNT_LOCKED, etc.), domain/errors.go, MapToAppError em delivery/http/errors
  - Fase 3: RespondWithError, Recovery middleware em infrastructure/http/middleware, registro no router
  - Documentação em docs/backend/README.md
- **Features/fixes concluídos:** 004-package-erro-padronizado
- **Tasks concluídas:** 3/3 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

### Sessão 9 — 2026-03-04
- **Atividades realizadas:**
  - Planejamento da Feature 005 — Endpoint de Login
  - Criação de documentos de planejamento para as 6 fases (entidades/repos, implementações GORM, serviços crypto, use case, handler, wiring)
- **Features/fixes criados:** 005-endpoint-login (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-entidades-repositorios-models.md](features/005-endpoint-login/fase-1-entidades-repositorios-models.md)

### Sessão 10 — 2026-03-04
- **Atividades realizadas:**
  - Implementação completa da Feature 005 — Endpoint de Login
  - Fase 1: Entidades (Tenant, Product, User, Subscription, UserProductAccess), interfaces de repositório, models GORM, erros de domínio (ErrUserBlocked, ErrProductNotFound, ErrProductUnavailable)
  - Fase 2: Repositórios GORM (Tenant, Product, User, Subscription, UserProductAccess)
  - Fase 3: BcryptPasswordService e RSAJWTService (RS256, 15 min)
  - Fase 4: Use Case AuthenticateUser com validações completas, timing constante (dummy hash), aud = slug + base domain
  - Fase 5: AuthHandler, DTOs LoginRequest/LoginResponse, rota POST /api/v1/auth/login
  - Fase 6: Wiring no main.go, variáveis JWT em .env.example, documentação em docs/backend/README.md
- **Features/fixes concluídos:** 005-endpoint-login
- **Tasks concluídas:** 6/6 fases
- **Próximas atividades:** Implementar próxima feature conforme TRACKER

### Sessão 11 — 2026-03-05
- **Atividades realizadas:**
  - Planejamento da Feature 006 — JWT como Cookie e URL de Redirect
  - Criação dos documentos de planejamento para as 3 fases (variável de ambiente, use case redirect_url, handler Set-Cookie)
  - Baseado na ATA-2025-03-05 (redirect seguro e passagem JWT via cookie)
- **Features/fixes criados:** 006-jwt-cookie-redirect-url (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-variavel-ambiente-dominio.md](features/006-jwt-cookie-redirect-url/fase-1-variavel-ambiente-dominio.md)

### Sessão 12 — 2026-03-05
- **Atividades realizadas:**
  - Planejamento da Feature 007 — Authorization Code Flow com Redis
  - Criação dos documentos de planejamento para as 5 fases: (1) Infraestrutura Redis, (2) AuthCodeStore no Redis, (3) Alteração do login para redirect 302, (4) Endpoint POST /auth/token, (5) Documentação de integração do cliente
  - Baseado na ATA-2025-03-05 e especificação do usuário: code TTL 40s, redirect HTTP 302, resposta { access_token, expires_in }
- **Features/fixes criados:** 007-auth-code-flow-redis (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-redis-infraestrutura.md](features/007-auth-code-flow-redis/fase-1-redis-infraestrutura.md)

### Sessão 13 — 2026-03-05
- **Atividades realizadas:**
  - Implementação completa da Feature 007 — Authorization Code Flow com Redis
  - Fase 1: Redis no docker-compose, REDIS_URL em .env.example, cache.NewRedisClient em infrastructure/cache
  - Fase 2: AuthCodeData, AuthCodeStore, RedisAuthCodeStore com GetDel (single-use), ErrCodeInvalidOrExpired, ErrAuthCodeStorageUnavailable
  - Fase 3: LoginInput.State, LoginOutput.RedirectURL, generateAuthCode (crypto/rand), authCodeStore no Use Case, c.Redirect(302)
  - Fase 4: ExchangeCodeForTokenUseCase, AuthHandler.Token, POST /api/v1/auth/token, TokenRequest/TokenResponse
  - Fase 5: docs/integration/auth-code-flow-integration.md com guia para clientes (callback, state, troca code, validação JWT)
- **Features/fixes concluídos:** 007-auth-code-flow-redis
- **Tasks concluídas:** 5/5 fases
- **Próximas atividades:** Próxima feature conforme TRACKER

### Sessão 14 — 2026-03-05
- **Atividades realizadas:**
  - Planejamento da Feature 008 — Refresh Token Endpoint
  - Criação dos documentos de planejamento para as 4 fases: (1) Migration product_id em refresh_tokens, (2) Refresh token no fluxo token exchange, (3) Endpoint POST /auth/refresh, (4) Documentação de integração
  - Análise: tabela refresh_tokens precisa de product_id para validação por tenant_slug e product_slug
- **Features/fixes criados:** 008-refresh-token-endpoint (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-migration-product-id-refresh-tokens.md](features/008-refresh-token-endpoint/fase-1-migration-product-id-refresh-tokens.md)

### Sessão 15 — 2026-03-05
- **Atividades realizadas:**
  - Implementação completa da Feature 008 — Refresh Token Endpoint
  - Fase 1: Migration 000008 — product_id em refresh_tokens, índice idx_refresh_tokens_lookup
  - Fase 2: AuthCodeData.ProductID, RefreshToken entity/repository/generator, refresh no POST /auth/token
  - Fase 3: RefreshTokenUseCase, POST /api/v1/auth/refresh, validação hash+tenant+product, rotação atômica, verificação assinatura
  - Fase 4: Documentação em docs/integration e docs/backend
- **Features/fixes concluídos:** 008-refresh-token-endpoint
- **Tasks concluídas:** 4/4 fases
- **Próximas atividades:** Próxima feature conforme TRACKER

### Sessão 16 — 2026-03-05
- **Atividades realizadas:**
  - Planejamento da Feature 009 — Public Key Endpoint (JWKS)
  - Criação dos documentos de planejamento para as 3 fases: (1) JWKS Provider extração chave pública, (2) Endpoint GET /.well-known/jwks.json, (3) Documentação integração e VPS/NGINX
  - Criação de docs/vps/configurations.md com config NGINX para endpoint JWKS
- **Features/fixes criados:** 009-public-key-endpoint (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-jwks-provider-extrair-chave-publica.md](features/009-public-key-endpoint/fase-1-jwks-provider-extrair-chave-publica.md)

### Sessão 17 — 2026-03-05
- **Atividades realizadas:**
  - Implementação completa da Feature 009 — Public Key Endpoint (JWKS)
  - Fase 1: domain/service/jwks_provider.go (JWK, JWKS, JWKSProvider); infrastructure/crypto/rsa_jwks_provider.go (RSAJWKSProvider, conversão rsa.PublicKey → JWK base64url)
  - Fase 2: delivery/http/handler/jwks_handler.go (JWKSHandler, GetJWKS); rota GET /.well-known/jwks.json no main.go; Cache-Control 24h
  - Fase 3: Seção "Obtenção da Chave Pública (JWKS)" em docs/integration/auth-code-flow-integration.md; GET /.well-known/jwks.json em docs/backend/README.md; docs/vps/configurations.md já existente
- **Features/fixes concluídos:** 009-public-key-endpoint
- **Tasks concluídas:** 3/3 fases
- **Próximas atividades:** Próxima feature conforme TRACKER

### Sessão 18 — 2026-03-05
- **Atividades realizadas:**
  - Planejamento da Feature 010 — Estrutura Inicial do Frontend
  - Criação dos documentos de planejamento para as 5 fases: (1) Estrutura de diretórios Angular, (2) Importar bibliotecas, (3) .gitignore, (4) Dockerfile multi-stage, (5) Docker Compose
- **Features/fixes criados:** 010-estrutura-inicial-frontend (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-estrutura-diretorios.md](features/010-estrutura-inicial-frontend/fase-1-estrutura-diretorios.md)

### Sessão 19 — 2026-03-06
- **Atividades realizadas:**
  - Implementação completa da Feature 010 — Estrutura Inicial do Frontend
  - Fase 1: `ng new frontend` (Angular 18) + estrutura Clean Architecture (core/auth, core/http, domain/models, domain/ports, application/use-cases, infrastructure/http, infrastructure/storage, features/auth, features/home, shared/components, shared/pipes)
  - Fase 2: package.json com engines Node >=18.19.1, validação build production
  - Fase 3: frontend/.gitignore com regras Angular/Node e .env (sem .env.example)
  - Fase 4: frontend/Dockerfile multi-stage (node:20-alpine builder, nginx:alpine runtime), frontend/nginx.conf com try_files para SPA, .dockerignore
  - Fase 5: serviço frontend no docker-compose.yml (porta 4200:80, depends_on backend)
  - frontend/README.md com estrutura, Docker e requisitos Node
- **Features/fixes concluídos:** 010-estrutura-inicial-frontend
- **Tasks concluídas:** 5/5 fases
- **Próximas atividades:** Próxima feature conforme TRACKER

### Sessão 20 — 2026-03-05
- **Atividades realizadas:**
  - Planejamento da Feature 011 — Tela de Login
  - Criação dos documentos de planejamento para as 5 fases: (1) Design tokens e estilos base, (2) Estrutura e rota da página de login, (3) Layout da página (gradient, overlay), (4) Card e formulário, (5) Componentes visuais, responsividade e acessibilidade
  - Design baseado no protótipo Login-Wisa (React/Tailwind); adaptação para Angular/SCSS
  - Escopo: design apenas, sem lógica de autenticação (a ser implementada em features futuras)
- **Features/fixes criados:** 011-tela-login (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-design-tokens-estilos.md](features/011-tela-login/fase-1-design-tokens-estilos.md)

### Sessão 21 — 2026-03-05
- **Atividades realizadas:**
  - Implementação completa da Feature 011 — Tela de Login
  - Fase 1: Design tokens em `frontend/src/styles/_login-tokens.scss`, fonte Inter em index.html, variáveis CSS para gradiente, card, tipografia, inputs e botão
  - Fase 2: LoginPageComponent em `features/auth/login/`, rota lazy-loaded `/login`, redirecionamento `''` → `'login'`
  - Fase 3: Layout full-screen com gradient (slate→blue), overlay de textura noise.svg, centralização flex
  - Fase 4: Card com barra gradient, formulário usuário/senha, ícones SVG inline (User, Lock, Eye/EyeOff), toggle senha, Signals (username, password, showPassword)
  - Fase 5: data-testid em elementos interativos, labels for/id, aria-label no toggle, responsividade (padding, min-height 44px touch targets), required e autocomplete nos inputs, preventDefault nos links placeholder
- **Features/fixes concluídos:** 011-tela-login
- **Tasks concluídas:** 5/5 fases
- **Próximas atividades:** Próxima feature conforme TRACKER

### Sessão 22 — 2026-03-06
- **Atividades realizadas:**
  - Planejamento da Feature 012 — Frontend Login Implementation
  - Criação dos documentos de planejamento para as 3 fases: (1) Parâmetros de query e validação (tenant_slug, product_slug, state), (2) Serviço de autenticação e configuração HTTP, (3) Integração do formulário, submit e redirect
  - Análise: frontend espera resposta JSON com redirect_url; backend atual retorna HTTP 302 — documentado como pré-requisito para ajuste no backend se necessário
- **Features/fixes criados:** 012-frontend-login-implementation (apenas planejamento)
- **Tasks concluídas:** —
- **Próximas atividades:** Implementar Fase 1 conforme [fase-1-parametros-query-validacao.md](features/012-frontend-login-implementation/fase-1-parametros-query-validacao.md)

### Sessão 23 — 2026-03-06
- **Atividades realizadas:**
  - Implementação completa da Feature 012 — Frontend Login Implementation
  - Fase 1: LoginPageComponent lê tenant_slug, product_slug, state via ActivatedRoute; validação no ngOnInit; bloco de erro com role="alert" e data-testid="login-params-error" quando params inválidos
  - Fase 2: core/http/api-config.ts (API_BASE_URL), core/auth/auth.service.ts (LoginRequest, LoginResponse, login()), provideHttpClient() e API_BASE_URL em app.config; proxy.conf.json para dev (/api → backend:8080)
  - Backend: auth_handler.go retorna 200 + JSON { redirect_url } quando Accept contém application/json (SPA); mantém 302 para fluxo tradicional
  - Fase 3: onSubmit com AuthService.login(), window.location.href no sucesso; loginError e isSubmitting; label Email, input type="email"; takeUntilDestroyed para unsubscribe; bloco login-error com data-testid
- **Features/fixes concluídos:** 012-frontend-login-implementation
- **Tasks concluídas:** 3/3 fases
- **Próximas atividades:** Próxima feature conforme TRACKER

---

## Como Usar Este Tracker

1. **Ao iniciar uma sessão:** Leia este arquivo para saber o status atual
2. **Para detalhes de uma feature:** Consulte o `TRACKER.md` dentro da pasta da feature e os arquivos de tasks
3. **Durante o trabalho:** Atualize o checkbox da task (`[ ]` → `[~]` → `[x]`) no arquivo correspondente
4. **Ao finalizar a sessão:** Atualize a tabela "Status Geral" e adicione uma entrada no "Diário de Sessões"
5. **Para criar nova feature ou fix:** Consulte o [README.md](README.md) com o guia de criação