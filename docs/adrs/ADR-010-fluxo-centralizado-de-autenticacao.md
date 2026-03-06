# ADR-010 — Fluxo Centralizado de Autenticação e Controle de Sessão

**Status:** Aceito  
**Data:** 2026-03-01  
**Alterado em:** 2025-03-05 (entrega do Access Token via code temporário + Redis — ver [ATA-2025-03-05](../DON'T%20READ/ideas-new-features/ATA-2025-03-05-redirect-seguro-e-passagem-jwt.md))  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (fluxo de autenticação end-to-end)

---

## Contexto

Este ADR documenta a decisão sobre o **fluxo completo de autenticação** — como o usuário passa da tela em branco (sistema do cliente) até ter acesso liberado, e como as sessões são gerenciadas ao longo do tempo.

O sistema tem uma característica central e inegociável: **o `wisa-crm-service` é o único ponto de autenticação para todos os sistemas clientes**. Sistemas clientes não implementam login próprio — eles delegam completamente ao Identity Provider centralizado.

Isso cria desafios específicos que precisam ser explicitamente endereçados:

1. **Como o sistema cliente redireciona o usuário** para o auth central de forma segura?
2. **Como o auth central retorna o token** ao sistema cliente sem interceptação?
3. **Como renovar a sessão** sem exigir login repetitivo a cada 15 minutos?
4. **Como garantir que o bloqueio por inadimplência seja efetivo** sem comprometer a experiência do usuário?
5. **Como proteger o endpoint de login** contra força bruta, credential stuffing e bots?
6. **Como gerenciar logout seguro** e invalidação de sessão?

---

## Decisão

**O fluxo de autenticação adotará o padrão de Authorization Code Flow** (OAuth 2.0), com:

- **Redirect seguro:** Backend monta internamente a `redirect_url` a partir de `tenant_slug` e `product_slug` — nenhuma URL externa é confiada (evita Open Redirect)
- **Code temporário:** Após login, backend gera um code opaco, armazena no **Redis** (TTL ~2 min), retorna `redirect_url` com `?code=...&state=...`
- **Troca code por token:** Cliente chama `POST /api/v1/auth/token` com `{"code": "..."}` no body; backend retorna `{"token": "JWT..."}` e invalida o code (uso único)
- Access Token de 15 minutos; Refresh Token rotativo com validade de 7 dias
- **Token nunca na URL** — apenas o code (opaco, descartável) aparece na query
- Proteção por rate limiting + account lockout progressivo
- Logout global via revogação do refresh token

**Infraestrutura:** Requer Redis para armazenamento dos authorization codes.

**Escopo:** Funciona para qualquer domínio — clientes em `*.wisa.labs.com.br` ou com domínio próprio.

---

## Justificativa

### 1. Fluxo completo de autenticação (passo a passo)

```
┌─────────┐     ┌───────────────────┐     ┌─────────────────────────────────┐
│ Usuário │     │ Sistema do Cliente│     │   wisa-crm-service (IdP)          │
└────┬────┘     └────────┬──────────┘     └──────────────────┬────────────────┘
     │                   │                                  │
     │ GET /dashboard     │                                  │
     ├──────────────────▶│                                  │
     │                   │ JWT ausente ou expirado           │
     │                   │ HTTP 302 Redirect                 │
     │◀──────────────────│ Location: https://auth.wisa.labs.com.br/login
     │                   │ ?tenant_slug=cliente1              │
     │                   │ &product_slug=gestao-pocket        │
     │                   │ &state=<random_state_token>       │
     │ GET /login?...     │                                  │
     ├───────────────────┼─────────────────────────────────▶│
     │                   │                                  │ Exibe tela de login
     │ Preenche email+senha                                 │
     ├───────────────────┼─────────────────────────────────▶│
     │                   │                                  │ 1. Valida tenant_slug
     │                   │                                  │ 2. Valida credenciais
     │                   │                                  │ 3. Valida assinatura
     │                   │                                  │ 4. Gera JWT + code temporário
     │                   │                                  │ 5. Armazena code no Redis (TTL ~2min)
     │                   │                                  │ 6. Monta redirect_url com ?code=&state=
     │                   │                                  │ 7. Resposta JSON: { redirect_url }
     │                   │◀─────────────────────────────────│
     │                   │                                  │
     │ Frontend auth: window.location = redirect_url         │
     │ GET https://cliente1.wisa.labs.com.br/gestao-pocket/callback?code=abc&state=...
     ├──────────────────▶│                                  │
     │                   │ Extrai code da URL                │
     │                   │ POST /api/v1/auth/token           │
     │                   │ Body: { "code": "abc" }            │
     │                   ├─────────────────────────────────▶│
     │                   │                                  │ 8. Busca code no Redis
     │                   │                                  │ 9. Retorna { "token": "JWT..." }
     │                   │                                  │ 10. Remove code (uso único)
     │                   │◀─────────────────────────────────│
     │                   │ Armazena token em memória         │
     │                   │ Valida JWT (sig, iss, aud, exp)   │
     │ Acesso liberado   │                                  │
     │◀──────────────────│                                  │
```

### 2. Segurança do redirect (evita Open Redirect)

O backend **monta a `redirect_url` internamente** a partir de `tenant_slug` e `product_slug` validados. Nenhuma URL externa é confiada como parâmetro de entrada:

- **Montagem interna:** A URL é construída com base em regra fixa ou whitelist por tenant:
  ```go
  // Exemplo: https://{tenant_slug}.wisa.labs.com.br/{product_slug}
  redirectURL := fmt.Sprintf("https://%s.wisa.labs.com.br/%s", tenantSlug, productSlug)
  ```
- **Alternativa:** Whitelist de `redirect_uri` por tenant no banco — a combinação `tenant_slug` + `product_slug` deve gerar uma URL permitida para aquele tenant
- **Resposta do login:** O backend retorna `{"redirect_url": "..."}` em JSON; o frontend do auth apenas executa `window.location.href = response.redirect_url`
- O parâmetro `state` é um token aleatório gerado pelo sistema cliente; o backend deve incluí-lo na `redirect_url` (ex: `?state=...`) para o cliente validar no retorno e prevenir CSRF

### 3. Entrega do token via Authorization Code (Redis)

**Resposta do login:** O backend retorna `{"redirect_url": "https://...?code=abc&state=xyz"}`. O token JWT **nunca** aparece na URL.

**Armazenamento do code no Redis:**
- Chave: `auth_code:{code}` — código opaco (ex: 32 bytes em hex)
- Valor: JSON com token(s) ou referência para emissão
- TTL: 120 segundos
- Uso único: removido imediatamente após troca por token

**Endpoint `POST /api/v1/auth/token`:**
```json
// Request
{ "code": "abc123xyz" }

// Response 200
{ "token": "eyJhbGciOiJSUzI1NiIs...", "refresh_token": "..." }
```

O endpoint busca o code no Redis, retorna o access token (e opcionalmente refresh token), e **remove** o code (uso único).

**Armazenamento no cliente:** O cliente armazena o access token **apenas em memória** (variável de serviço) — nunca localStorage nem sessionStorage. O refresh token pode ser retornado junto na troca e armazenado da mesma forma.

**Validação obrigatória:** Aplicação cliente valida JWT (sig, iss, aud, exp) antes de conceder acesso.

### 4. Fluxo de renovação de sessão (Refresh Token Rotation)

A renovação usa **Refresh Token Rotation**: a cada renovação, um novo refresh token é emitido e o antigo é imediatamente revogado. O refresh token é obtido na troca do code (`/api/v1/auth/token`) e armazenado em memória pelo cliente.

```
1. Access token expira (15 min)
2. Cliente recebe 401; chama POST /api/v1/auth/refresh com body: { "refresh_token": "..." }
3. wisa-crm-service:
   a. Valida refresh token no banco (não revogado, não expirado)
   b. Verifica assinatura do tenant (se vencida → 402, não renova)
   c. Revoga o refresh token atual no banco
   d. Emite novo access token (JWT 15 min)
   e. Emite novo refresh token (7 dias)
   f. Retorna 200 + { "token": "...", "refresh_token": "..." }
4. Cliente atualiza tokens em memória; pode retentar a requisição original
```

**Detecção de roubo de refresh token:**
- Se um refresh token já revogado é apresentado, o `wisa-crm-service` **invalida toda a família de tokens** (todos os refresh tokens do usuário naquele tenant) e exige novo login — isso detecta reuse attack

### 5. Proteção do endpoint de login

**Camadas de proteção (em ordem de aplicação):**

#### Camada 1: NGINX Rate Limiting (ADR-007)
- 5 requisições/minuto por IP
- Retorna HTTP 429 com cabeçalho `Retry-After`

#### Camada 2: Account Lockout progressivo (backend Go)
```go
// Lógica de lockout no Use Case
const (
    MaxFailedAttempts = 5
    LockoutDuration   = 15 * time.Minute
    ExtendedLockout   = 1 * time.Hour  // após 10 tentativas
)

func (uc *AuthenticateUserUseCase) checkAccountLockout(user *domain.User) error {
    if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
        return domain.ErrAccountLocked
    }
    return nil
}

func (uc *AuthenticateUserUseCase) recordFailedAttempt(ctx context.Context, user *domain.User) error {
    user.FailedAttempts++
    if user.FailedAttempts >= MaxFailedAttempts {
        lockUntil := time.Now().Add(LockoutDuration)
        if user.FailedAttempts >= 10 {
            lockUntil = time.Now().Add(ExtendedLockout)
        }
        user.LockedUntil = &lockUntil
        // Registrar evento de lockout no audit_log
    }
    return uc.userRepo.Update(ctx, user)
}
```

#### Camada 3: Resposta constante em tempo (anti-timing)
```go
// Sempre executar bcrypt, mesmo quando usuário não existe
// Isso normaliza o tempo de resposta e previne user enumeration via timing
func (uc *AuthenticateUserUseCase) Execute(ctx context.Context, input AuthInput) (*AuthOutput, error) {
    user, err := uc.userRepo.FindByEmail(ctx, input.TenantID, input.Email)
    
    dummyHash := "$2a$12$dummy_hash_for_timing_normalization_padding"
    hashToCompare := dummyHash
    if err == nil {
        hashToCompare = user.PasswordHash
    }
    
    // bcrypt sempre executa — tempo constante independente de usuário existir
    passwordValid := uc.passwordSvc.Compare(input.Password, hashToCompare)
    
    if err != nil || !passwordValid {
        return nil, domain.ErrInvalidCredentials  // mensagem genérica sempre
    }
    
    // ...
}
```

#### Camada 4: Mensagem de erro genérica
- Sempre retornar `{"error": "invalid_credentials"}` para qualquer falha de autenticação
- Nunca diferenciar "email não encontrado" de "senha incorreta"
- Nunca diferenciar "tenant não encontrado" de "credenciais inválidas"

### 6. Logout e invalidação de sessão

**Logout local (apenas na aba atual):**
- Cliente limpa access token e refresh token da memória
- Redireciona para tela de login (ou auth central)
- Opcional: chamar `POST /api/v1/auth/logout` com `refresh_token` no body para revogar a sessão atual

**Logout global (invalida todas as sessões):**
```
POST /api/v1/auth/logout
Body: { "refresh_token": "..." }

→ wisa-crm-service revoga TODOS os refresh tokens do usuário no tenant
→ Registra evento no audit_log
→ Retorna 200
```

Após logout global, todos os access tokens existentes expirarão naturalmente em até 15 minutos. Se necessário bloqueio imediato (ex: conta comprometida), o `jti` pode ser adicionado ao denylist.

### 7. Proteção contra CSRF no fluxo de autenticação

O parâmetro `state` no redirect de login é um token CSRF específico para o fluxo:

```javascript
// Sistema cliente — antes de redirecionar para login
const state = generateSecureRandomToken(); // ex: 32 bytes hex
sessionStorage.setItem('oauth_state', state);
window.location = `https://auth.wisa.labs.com.br/login?tenant_slug=cliente1&product_slug=gestao-pocket&state=${state}`;

// Após redirect de volta — validar state e trocar code por token
const params = new URLSearchParams(window.location.search);
const returnedState = params.get('state');
const code = params.get('code');
const savedState = sessionStorage.getItem('oauth_state');
if (returnedState !== savedState) {
    throw new Error('State mismatch — potential CSRF attack');
}
// Trocar code por token: POST /api/v1/auth/token { "code": code }
// Armazenar token em memória; opcional: history.replaceState para remover code da URL
```

---

## Consequências

### Positivas
- Fluxo seguro alinhado ao padrão OAuth 2.0 Authorization Code
- Token JWT **nunca** exposto na URL — apenas code opaco e descartável
- Funciona para **qualquer domínio** (subdomínios ou domínio próprio do cliente)
- Bloqueio por inadimplência funcional em no máximo 15 minutos
- Detecção de roubo de refresh token via rotation e reuse detection
- Logout global efetivo via revogação de todos os tokens
- Mensagem de erro genérica previne enumeração de usuários e tenants

### Negativas
- **Redis obrigatório** — nova dependência de infraestrutura
- **Latência extra** — requisição adicional (troca code por token) no primeiro acesso
- Token em memória é perdido ao recarregar página — requer lógica de renovação automática no cliente
- Múltiplas abas do mesmo sistema cliente precisam sincronizar token (BroadcastChannel ou shared worker)

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| Open Redirect | Baixa | Alto | Alta |
| Roubo de access token via XSS (token em memória) | Baixa | Alto | Alta |
| Credential stuffing com lista de vazamentos | Alta | Médio | Alta |
| Bypass de account lockout via diferentes IPs | Alta | Médio | Média |
| Indisponibilidade do Redis | Média | Alto | Alta |
| Code interceptado e trocado antes do cliente (janela ~2 min) | Baixa | Médio | Média |
| Refresh token extraído da memória via XSS | Baixa | Alto | Alta |
| Race condition no refresh token rotation | Baixa | Médio | Média |

---

## Mitigações

### Open Redirect
- Backend monta `redirect_url` internamente a partir de `tenant_slug` e `product_slug` — nenhuma URL externa confiada
- Alternativa: whitelist estrita por tenant no banco
- Log e alerta para tentativas com parâmetros suspeitos

### Indisponibilidade do Redis
- Configurar Redis com persistência e alta disponibilidade (ex: Redis Sentinel, cluster)
- Monitoramento de saúde do Redis; fallback: falha de login com mensagem adequada
- TTL curto no code reduz impacto de perda de dados no Redis

### Roubo de access token via XSS
- CSP rigoroso no Angular (ADR-002) previne XSS
- Access token de vida curta (15 min) limita a janela de exploração
- Monitorar uso de tokens por IP inesperado no audit_log

### Credential stuffing
- Rate limiting por IP (NGINX) + Account lockout (backend)
- Após N bloqueios, exigir verificação adicional (CAPTCHA, email de confirmação)
- Monitorar padrões de login de múltiplos IPs para a mesma conta
- Integrar com serviços de threat intelligence (ex: HaveIBeenPwned API) para senhas vazadas

### Bypass de lockout por IP rotation
- Lockout baseado em conta (email), não apenas por IP
- Mesmo que o atacante rotacione IPs, a conta é bloqueada após N tentativas
- Rate limiting por IP como proteção complementar, não como proteção primária

### Race condition no refresh
- Usar `SELECT FOR UPDATE` ao buscar e revogar o refresh token (garantia de atomicidade)
- Transação explícita que revoga o token antigo e insere o novo de forma atômica

---

## Alternativas Consideradas

### OAuth 2.0 completo (Authorization Code Flow)
- **Prós:** Padrão da indústria, bibliotecas de cliente disponíveis para qualquer linguagem, suporta terceiros
- **Contras:** Complexidade significativamente maior (authorization server, consent screen, scopes, client credentials); overhead desnecessário para um sistema fechado onde todos os clientes são confiáveis e controlados pelo mesmo operador

### SAML 2.0
- **Prós:** Padrão enterprise maduro, suporte a SSO com provedores externos
- **Contras:** Complexidade de implementação extremamente alta (XML, metadata, assertions), overkill para o caso de uso atual

### Session-based authentication (sem JWT)
- **Prós:** Revogação imediata de sessão, sem problema de token expirado
- **Contras:** Requer que sistemas clientes consultem o `wisa-crm-service` em cada requisição para validar a sessão — viola o requisito de validação local nos sistemas clientes; cria acoplamento e ponto único de falha

**O fluxo de Authorization Code simplificado com JWT assimétrico é a escolha mais adequada para este sistema fechado e controlado, mantendo segurança sem complexidade desnecessária.**

---

## Referências

- [ATA-2025-03-05 — Redirect Seguro e Passagem de JWT](../DON'T READ/ideas-new-features/ATA-2025-03-05-redirect-seguro-e-passagem-jwt.md)
- [OAuth 2.0 Security Best Current Practice (RFC 9700)](https://datatracker.ietf.org/doc/html/rfc9700)
- [OAuth 2.0 for Browser-Based Apps (RFC 9449)](https://datatracker.ietf.org/doc/html/rfc9449)
- [Refresh Token Rotation](https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
