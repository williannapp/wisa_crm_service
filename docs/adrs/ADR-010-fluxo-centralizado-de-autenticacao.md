# ADR-010 — Fluxo Centralizado de Autenticação e Controle de Sessão

**Status:** Aceito  
**Data:** 2026-03-01  
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

**O fluxo de autenticação adotará o padrão de Authorization Code Flow simplificado** (inspirado em OAuth 2.0, mas sem o overhead de um servidor OAuth completo), com:

- Redirect-based login via parâmetro `redirect_uri`
- Tokens entregues via **cookie HTTP-only + Secure + SameSite=Strict**
- Refresh Token rotativo com validade de 7 dias
- Access Token de 15 minutos
- Proteção por rate limiting + account lockout progressivo
- Logout global via revogação do refresh token

---

## Justificativa

### 1. Fluxo completo de autenticação (passo a passo)

```
┌─────────┐     ┌───────────────────┐     ┌─────────────────────────────┐
│ Usuário │     │ Sistema do Cliente│     │   wisa-crm-service (IdP)    │
└────┬────┘     └────────┬──────────┘     └──────────────┬──────────────┘
     │                   │                               │
     │ GET /dashboard     │                               │
     ├──────────────────▶│                               │
     │                   │ JWT ausente ou expirado        │
     │                   │ HTTP 302 Redirect              │
     │◀──────────────────│ Location: https://auth.wisa-crm.com/login
     │                   │ ?redirect_uri=https://cliente1.seusistema.com/auth/callback
     │                   │ &tenant_slug=cliente1          │
     │                   │ &state=<random_state_token>    │
     │ GET /login?...    │                               │
     ├───────────────────┼──────────────────────────────▶│
     │                   │                               │ Exibe tela de login
     │ Preenche email+senha                              │
     ├───────────────────┼──────────────────────────────▶│
     │                   │                               │ 1. Valida tenant_slug
     │                   │                               │ 2. Valida credenciais
     │                   │                               │ 3. Valida assinatura
     │                   │                               │ 4. Gera access_token + refresh_token
     │                   │                               │ 5. Valida redirect_uri (whitelist)
     │                   │                               │ 6. HTTP 302 Redirect para redirect_uri
     │                   │                               │    + Set-Cookie: refresh_token (HTTP-only)
     │                   │                               │    + access_token no redirect hash ou query
     │                   │◀──────────────────────────────│ HTTP 302
     │                   │                               │
     │ GET /auth/callback│                               │
     ├──────────────────▶│                               │
     │                   │ Extrai access_token           │
     │                   │ Valida JWT (sig, iss, aud, exp)│
     │                   │ Armazena em memória (não localStorage)
     │ Acesso liberado   │                               │
     │◀──────────────────│                               │
```

### 2. Segurança do redirect_uri

O parâmetro `redirect_uri` é um vetor de ataque clássico em OAuth (Open Redirect). A proteção:

- **Whitelist estrita por tenant:** cada tenant cadastra previamente a lista de `redirect_uri` permitidas
- O `wisa-crm-service` **recusa qualquer `redirect_uri` não cadastrada** para o tenant:

```go
// No Use Case de autenticação
func (uc *AuthenticateUserUseCase) validateRedirectURI(tenantID uuid.UUID, redirectURI string) error {
    allowed, err := uc.tenantRepo.GetAllowedRedirectURIs(ctx, tenantID)
    if err != nil {
        return err
    }
    for _, uri := range allowed {
        if uri == redirectURI {
            return nil
        }
    }
    return domain.ErrInvalidRedirectURI
}
```

- O parâmetro `state` é um token aleatório gerado pelo sistema cliente que deve ser validado no callback para prevenir CSRF no fluxo de autenticação

### 3. Entrega do token e armazenamento

**Access Token:** entregue como parâmetro de query no redirect de callback (`?access_token=...`), capturado pelo Angular **apenas em memória** (variável de serviço) — nunca em localStorage nem sessionStorage.

**Refresh Token:** entregue via **Set-Cookie HTTP-only + Secure + SameSite=Strict**:
```
Set-Cookie: refresh_token=<value>; HttpOnly; Secure; SameSite=Strict; Path=/api/v1/auth/refresh; Max-Age=604800
```

**Por que esta separação?**
- O access token em memória desaparece ao fechar a aba — comportamento seguro
- O refresh token em cookie HTTP-only é invisível ao JavaScript — imune a XSS
- `SameSite=Strict` protege o refresh token contra CSRF
- `Path=/api/v1/auth/refresh` garante que o cookie só seja enviado para o endpoint de renovação

**Trade-off documentado:** o access token em memória é perdido ao recarregar a página. O Angular detecta a ausência do token e usa o refresh token (cookie) para obter novo access token automaticamente — transparente para o usuário.

### 4. Fluxo de renovação de sessão (Refresh Token Rotation)

A renovação usa **Refresh Token Rotation**: a cada renovação, um novo refresh token é emitido e o antigo é imediatamente revogado. Isso detecta roubo de refresh tokens:

```
1. Access token expira (15 min)
2. Angular detecta 401 e chama POST /api/v1/auth/refresh
   (refresh_token enviado automaticamente via cookie)
3. wisa-crm-service:
   a. Valida refresh token no banco (não revogado, não expirado)
   b. Verifica assinatura do tenant (se vencida → 402, não renova)
   c. Revoga o refresh token atual no banco
   d. Emite novo access token (JWT 15 min)
   e. Emite novo refresh token (7 dias)
   f. Retorna novo access token + Set-Cookie com novo refresh token
4. Angular atualiza o access token em memória
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
- Angular limpa o access token da memória
- Redireciona para tela de login

**Logout global (invalida todas as sessões):**
```
POST /api/v1/auth/logout
Cookie: refresh_token=<value>

→ wisa-crm-service revoga TODOS os refresh tokens do usuário no tenant
→ Limpa cookie: Set-Cookie: refresh_token=; Max-Age=0; ...
→ Registra evento no audit_log
```

Após logout global, todos os access tokens existentes expirarão naturalmente em até 15 minutos. Se necessário bloqueio imediato (ex: conta comprometida), o `jti` pode ser adicionado ao denylist.

### 7. Proteção contra CSRF no fluxo de autenticação

O parâmetro `state` no redirect de login é um token CSRF específico para o fluxo:

```javascript
// Sistema cliente — antes de redirecionar para login
const state = generateSecureRandomToken(); // ex: 32 bytes hex
sessionStorage.setItem('oauth_state', state);
window.location = `https://auth.wisa-crm.com/login?redirect_uri=...&state=${state}`;

// Callback — validar state antes de processar o token
const returnedState = new URLSearchParams(window.location.search).get('state');
const savedState = sessionStorage.getItem('oauth_state');
if (returnedState !== savedState) {
    // CSRF detectado — não processar o token
    throw new Error('State mismatch — potential CSRF attack');
}
```

---

## Consequências

### Positivas
- Fluxo seguro de autenticação com múltiplas camadas de proteção
- Refresh token em cookie HTTP-only imune a XSS
- Bloqueio por inadimplência funcional em no máximo 15 minutos
- Detecção de roubo de refresh token via rotation e reuse detection
- Logout global efetivo via revogação de todos os tokens
- Mensagem de erro genérica previne enumeração de usuários e tenants

### Negativas
- Access token em memória é perdido ao recarregar página — requer lógica de renovação automática no Angular
- Múltiplas abas do mesmo sistema cliente compartilham cookie mas precisam sincronizar access token (BroadcastChannel API ou shared worker)
- Logout local não invalida o token imediatamente — janela de 15 minutos aceita como trade-off
- Fluxo de redirect requer configuração prévia de `redirect_uri` por tenant

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| Open Redirect via redirect_uri manipulada | Baixa | Alto | Alta |
| Roubo de access token da memória via XSS | Baixa | Alto | Alta |
| Credential stuffing com lista de vazamentos | Alta | Médio | Alta |
| Bypass de account lockout via diferentes IPs | Alta | Médio | Média |
| Refresh token extraído de cookie via subdomain takeover | Muito Baixa | Crítico | Média |
| Race condition no refresh token rotation | Baixa | Médio | Média |

---

## Mitigações

### Open Redirect
- Whitelist estrita de redirect_uris por tenant (configuração no banco)
- Validação por comparação exata (não por prefix matching)
- Log e alerta para tentativas com redirect_uri não autorizada

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

- [OAuth 2.0 Security Best Current Practice (RFC 9700)](https://datatracker.ietf.org/doc/html/rfc9700)
- [OAuth 2.0 for Browser-Based Apps (RFC 9449)](https://datatracker.ietf.org/doc/html/rfc9449)
- [Refresh Token Rotation](https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)
