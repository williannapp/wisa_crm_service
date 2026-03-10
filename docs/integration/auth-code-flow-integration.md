# Integração com o Fluxo de Authorization Code — wisa-crm-service

Este documento orienta as aplicações clientes (ex.: gestao-pocket, sistemas em Angular + backend) sobre como integrar com o fluxo de Authorization Code do **wisa-crm-service** (Identity Provider centralizado).

---

## Pré-requisitos

- **URL do auth server:** configurável via variável de ambiente (ex.: `AUTH_SERVER_URL=https://auth.wisa.labs.com.br`)
- **Chave pública do auth:** obter via endpoint `GET /.well-known/jwks.json`
- **TLS em produção:** certificado válido para o domínio do auth e do cliente

---

## Ordem de Implementação Recomendada

Para integradores que começam do zero, a ordem sugerida é:

1. **Redirect + state** — Gerar state, armazenar em cookie/session, redirecionar para `{AUTH_SERVER_URL}/login?tenant_slug=...&product_slug=...&state=...`
2. **Callback + token exchange** — Implementar rota GET que recebe `code` e `state`, valida state, chama `POST /auth/token`, armazena tokens
3. **Armazenamento de token** — Cookie HttpOnly ou session server-side (nunca localStorage)
4. **Validação JWT no backend** — Assinatura, iss, aud, exp, nbf, kid
5. **Interceptor + refresh** — Ao receber 401 na API protegida, chamar `POST /auth/refresh`, retentar requisição ou redirecionar para login

---

## Visão Geral do Fluxo

```
┌─────────┐     ┌───────────────────┐     ┌─────────────────────────────────┐
│ Usuário │     │ Sistema Cliente  │     │   wisa-crm-service (IdP)         │
└────┬────┘     └────────┬──────────┘     └──────────────────┬─────────────┘
     │                   │                                  │
     │ 1. Acessa app     │                                  │
     ├──────────────────▶│                                  │
     │                   │ 2. Redirect 302                  │
     │◀──────────────────│ Location: auth/login?tenant_slug=...&product_slug=...&state=...
     │                   │                                  │
     │ 3. GET auth/login  │                                  │
     ├───────────────────┼─────────────────────────────────▶│
     │                   │                                  │ Exibe tela de login
     │ 4. Preenche email+senha                              │
     ├───────────────────┼─────────────────────────────────▶│
     │                   │ POST /auth/login                  │
     │                   │                                  │ Valida credenciais
     │                   │                                  │ Gera code, armazena Redis (TTL 40s)
     │ 5. Redirect 302   │                                  │
     │◀──────────────────│◀─────────────────────────────────│
     │ Location: cliente.com/product/callback?code=...&state=...
     │                   │                                  │
     │ 6. GET /callback?code=&state=                         │
     ├──────────────────▶│                                  │
     │                   │ 7. POST /auth/token { code }      │
     │                   ├─────────────────────────────────▶│
     │                   │                                  │ Valida code, emite JWT
     │                   │◀─────────────────────────────────│
     │                   │ 8. { access_token, refresh_token, expires_in } │
     │                   │ Armazena em cookie/sessão        │
     │ 9. Acesso liberado │                                  │
     │◀──────────────────│                                  │
```

---

## Processo de Autenticação — Passo a Passo

### 1.1 Iniciar login (redirect)

O sistema cliente, ao detectar usuário não autenticado, deve:

1. **Gerar `state`** — token aleatório seguro (ex.: 32 bytes em hex)
2. **Armazenar `state`** em cookie (ex.: `oauth_state`, HttpOnly, SameSite=Strict) ou sessão
3. **Redirecionar** o usuário (HTTP 302) para a página de login do auth:

```
GET {AUTH_SERVER_URL}/login?tenant_slug={tenant}&product_slug={product}&state={state}
```

**Parâmetros obrigatórios:**

| Parâmetro    | Descrição                          |
|---------------|------------------------------------|
| `tenant_slug` | Identificador do tenant (cliente) |
| `product_slug`| Identificador do produto           |
| `state`      | Token CSRF para validação no callback |

Exemplo: `https://auth.wisa.labs.com.br/login?tenant_slug=cliente1&product_slug=gestao-pocket&state=a1b2c3d4...`

---

### 1.2 Envio de credenciais (POST /api/v1/auth/login)

O frontend do auth (ou o cliente, se hospedar a tela) envia as credenciais via API.

**Request:**
```http
POST {AUTH_SERVER_URL}/api/v1/auth/login
Content-Type: application/json
Accept: application/json    ← obrigatório para resposta JSON (fluxo SPA)
```

**Body:**
```json
{
  "tenant_slug": "cliente1",
  "product_slug": "gestao-pocket",
  "user_email": "usuario@example.com",
  "password": "senha123",
  "state": "a1b2c3d4..."
}
```

| Campo        | Obrigatório | Descrição                              |
|--------------|-------------|----------------------------------------|
| `tenant_slug`| Sim         | Identificador do tenant                |
| `product_slug`| Sim        | Identificador do produto               |
| `user_email` | Sim         | Email do usuário                       |
| `password`   | Sim         | Senha do usuário                       |
| `state`      | Recomendado | Token CSRF — incluído na redirect_url  |

**Respostas de sucesso:**

- **200 OK** (quando `Accept: application/json`):
  ```json
  {
    "redirect_url": "https://cliente.wisa.labs.com.br/gestao-pocket/callback?code=abc123...&state=a1b2c3d4..."
  }
  ```
  O frontend deve executar `window.location.href = response.redirect_url`.

- **302 Found** (fluxo form tradicional, sem Accept: application/json):
  O backend redireciona diretamente para a `redirect_url`. O navegador segue o redirect automaticamente.

**Respostas de erro:**

Todas as respostas de erro usam o formato padronizado:

```json
{
  "status": "error",
  "error_code": "INVALID_CREDENTIALS",
  "message": "Credenciais inválidas.",
  "details": ""
}
```

| HTTP | error_code              | Descrição                                          |
|------|-------------------------|----------------------------------------------------|
| 400  | INVALID_REQUEST         | Dados inválidos (campos faltando ou formato)       |
| 401  | INVALID_CREDENTIALS     | Credenciais inválidas (mensagem genérica)          |
| 403  | ACCOUNT_LOCKED          | Conta bloqueada por tentativas de login            |
| 403  | USER_BLOCKED            | Usuário sem permissão                              |
| 403  | PRODUCT_UNAVAILABLE     | Produto indisponível                               |
| 403  | SUBSCRIPTION_SUSPENDED  | Assinatura suspensa por pendência financeira       |
| 403  | SUBSCRIPTION_CANCELED   | Assinatura cancelada                               |
| 404  | TENANT_NOT_FOUND        | Tenant não encontrado                              |
| 429  | RATE_LIMIT_EXCEEDED     | Muitas tentativas; aguardar Retry-After            |
| 500  | INTERNAL_ERROR          | Erro interno                                       |
| 503  | SERVICE_UNAVAILABLE     | Serviço temporariamente indisponível (ex.: Redis)  |

---

### 1.3 Implementação do callback

O **backend** da aplicação cliente deve implementar um endpoint que recebe o redirect do auth após login bem-sucedido.

**Rota:** `GET /callback` ou `GET /{product_slug}/callback` (conforme URL montada pelo auth)

**Query params recebidos:** `code`, `state`

**Passos no handler:**

1. **Extrair** `code` e `state` da query string.
2. **Validar `state`** — comparar com o valor armazenado (cookie/sessão) antes do redirect. Se diferente → 400 ou redirect para página de erro (possível CSRF).
3. **Remover state** após uso (one-time).
4. **Trocar code por token:**
   - `POST {AUTH_SERVER_URL}/api/v1/auth/token`
   - Body: `{ "code": "<code>" }`
   - Content-Type: `application/json`
5. **Se 401** (code inválido/expirado): redirecionar para login novamente.
6. **Se 200:** extrair `access_token`, `refresh_token`, `expires_in`, `refresh_expires_in`.
7. **Armazenar tokens** de forma segura (cookie HttpOnly ou sessão).
8. **Redirecionar** para a aplicação (ex.: `/dashboard`).

**Segurança:**

- **Code:** TTL de 40 segundos no Redis; uso único (removido após troca).
- **State:** Obrigatório para prevenção de CSRF; sempre validar.
- **Open Redirect:** O backend monta `redirect_url` internamente — nunca aceitar URL externa como parâmetro.
- **Não logar** o `code` em produção; não incluí-lo em URLs de erro.

---

### 1.4 Troca de code por token (POST /api/v1/auth/token)

**Request:**
```http
POST {AUTH_SERVER_URL}/api/v1/auth/token
Content-Type: application/json

{
  "code": "64_hex_chars_received_in_callback_url"
}
```

**Resposta 200:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_in": 900,
  "refresh_token": "abc123...",
  "refresh_expires_in": 604800
}
```

- `access_token`: JWT RS256, expira em 15 minutos (900 segundos)
- `expires_in`: segundos até a expiração do access token
- `refresh_token`: token opaco para renovação de sessão (7 dias)
- `refresh_expires_in`: segundos até a expiração do refresh token (604800 = 7 dias)

**Respostas de erro:**

Formato: `{ "status": "error", "error_code": "...", "message": "...", "details": "..." }`

| HTTP | error_code               | Descrição                                      |
|------|--------------------------|------------------------------------------------|
| 400  | INVALID_REQUEST          | Campo code ausente ou inválido                 |
| 401  | CODE_INVALID_OR_EXPIRED   | Código inválido, expirado ou já usado         |
| 500  | INTERNAL_ERROR            | Erro interno                                   |
| 503  | SERVICE_UNAVAILABLE       | Redis indisponível                             |

---

## Processo de Refresh — Passo a Passo

### 2.1 Quando chamar

O access token expira em 15 minutos. Quando requisições à API protegida retornam **401**, o cliente deve chamar o endpoint de refresh antes de redirecionar para login.

### 2.2 Request

```http
POST {AUTH_SERVER_URL}/api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "<token recebido na troca de code>",
  "tenant_slug": "cliente1",
  "product_slug": "gestao-pocket"
}
```

Todos os campos são obrigatórios.

### 2.3 Resposta 200

```json
{
  "access_token": "...",
  "expires_in": 900,
  "refresh_token": "...",
  "refresh_expires_in": 604800
}
```

O novo `refresh_token` **substitui** o antigo; o antigo é invalidado (rotação). Sempre armazene apenas o refresh token mais recente.

### 2.4 Respostas de erro (POST /auth/refresh)

| HTTP | error_code             | Descrição                                      | Ação recomendada          |
|------|------------------------|------------------------------------------------|---------------------------|
| 400  | INVALID_REQUEST        | Dados inválidos (campos faltando)              | Corrigir request          |
| 401  | INVALID_CREDENTIALS    | Token inválido, expirado ou revogado           | Redirecionar para login   |
| 402  | SUBSCRIPTION_EXPIRED   | Assinatura vencida                             | Redirecionar para renovação |
| 429  | RATE_LIMIT_EXCEEDED    | Muitas tentativas                              | Aguardar Retry-After      |
| 500  | INTERNAL_ERROR         | Erro interno                                   | Mensagem genérica         |
| 503  | SERVICE_UNAVAILABLE     | Serviço indisponível                           | Tentar novamente mais tarde |

### 2.5 Fluxo no Interceptor

1. Cliente faz requisição com `Authorization: Bearer {access_token}`.
2. API retorna **401**.
3. **Importante:** Se a requisição original for a própria chamada de refresh ou de login, **não** retentar — evitar loop infinito.
4. Cliente chama `POST /auth/refresh` com `refresh_token`, `tenant_slug`, `product_slug` (**não** use o access_token na chamada de refresh).
5. **Se 200:** atualizar access_token e refresh_token em memória; **retentar** a requisição original.
6. **Se 401:** limpar tokens; redirecionar para login.
7. **Se 402:** redirecionar para página de renovação de plano.
8. **Se 429:** aguardar cabeçalho `Retry-After`; exibir mensagem ao usuário.

### 2.6 Rotação e boas práticas

- **Rotação:** O refresh_token retornado substitui o antigo; o antigo é invalidado imediatamente.
- **Armazenamento:** Nunca localStorage (vulnerável a XSS). Preferir cookie HttpOnly ou armazenamento server-side.
- **401 no refresh:** Possível token revogado ou roubado — limpar sessão e redirecionar para login.

---

## Obtenção da Chave Pública (JWKS)

A aplicação cliente deve obter a chave pública RSA para validar a assinatura dos JWTs. O wisa-crm-service expõe um endpoint padrão em formato JWKS (RFC 7517).

### URL do endpoint

```
GET {AUTH_SERVER_URL}/.well-known/jwks.json
```

Exemplo: `https://auth.wisa-crm.com/.well-known/jwks.json`

### Formato da resposta

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-2026-v1",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

### Uso no cliente

1. **Selecionar pelo `kid`:** O header do JWT contém o claim `kid`. Use-o para localizar a chave correspondente no array `keys`.
2. **Cache:** O endpoint envia `Cache-Control: public, max-age=86400` (24 horas). O cliente pode cachear a resposta localmente.
3. **Rotação:** Durante rotação, o JWKS pode conter múltiplas chaves. Tokens antigos continuam válidos até expirar; novos tokens usam o novo `kid`.

---

## Validação do JWT no Cliente

Antes de conceder acesso, a aplicação cliente **deve**:

1. **Validar assinatura RS256** usando a chave pública do wisa-crm-service
2. **Verificar `iss`** (issuer) — deve corresponder ao auth configurado
3. **Verificar `aud`** (audiência) — deve corresponder ao domínio do cliente
4. **Verificar `exp`** (expiração)
5. **Verificar `nbf`** (not before), se presente
6. **Usar `kid`** do header para selecionar a chave correta no JWKS

Referência: [ADR-006 — JWT com Assinatura Assimétrica](../adrs/ADR-006-jwt-com-assinatura-assimetrica.md).

---

## Logout

### Logout local (implementado pelo cliente)

- Limpar `access_token` e `refresh_token` da memória/cookie/sessão
- Redirecionar o usuário para a tela de login (ou auth central)

### Logout global (futuro)

O ADR-010 prevê o endpoint `POST /api/v1/auth/logout` com `refresh_token` no body para revogar a sessão atual e, em cenário avançado, invalidar todas as sessões do usuário no tenant. **Este endpoint ainda não está implementado.** Quando disponível, será documentado em atualização futura.

---

## Armazenamento do Token

- **Cookie HttpOnly + Secure + SameSite=Strict:** recomendado para SPAs que fazem requisições ao backend do cliente no mesmo domínio
- **Session server-side:** alternativa válida para aplicações tradicionais
- **Nunca:** localStorage, sessionStorage (vulnerável a XSS) — exceto token em memória de curta duração em SPA com CSP rigoroso

---

## Tabela Consolidada de Erros

### Por HTTP status — Ação geral

| HTTP | Situação                                | Ação do Cliente                              |
|------|----------------------------------------|----------------------------------------------|
| 400  | Dados inválidos                         | Corrigir campos; exibir mensagem de validação|
| 401  | Code/token inválido ou expirado         | Redirecionar para login                      |
| 402  | Assinatura vencida (no refresh)         | Redirecionar para renovação de plano         |
| 403  | Conta bloqueada, produto indisponível  | Exibir mensagem específica do error_code     |
| 404  | Tenant não encontrado (login)           | Verificar tenant_slug                         |
| 429  | Rate limit excedido                     | Aguardar Retry-After; exibir mensagem        |
| 500  | Erro interno                            | Mensagem genérica; não expor detalhes        |
| 503  | Serviço indisponível                    | Exibir "Tente novamente mais tarde"          |

### Por endpoint — error_code e message

| Endpoint        | error_code               | HTTP | Quando ocorre                          |
|-----------------|--------------------------|------|----------------------------------------|
| POST /auth/login| INVALID_REQUEST          | 400  | Campos inválidos ou faltando           |
| POST /auth/login| INVALID_CREDENTIALS      | 401  | Credenciais inválidas                   |
| POST /auth/login| ACCOUNT_LOCKED           | 403  | Conta bloqueada por tentativas          |
| POST /auth/login| USER_BLOCKED             | 403  | Usuário bloqueado                      |
| POST /auth/login| PRODUCT_UNAVAILABLE      | 403  | Produto indisponível                   |
| POST /auth/login| SUBSCRIPTION_SUSPENDED   | 403  | Assinatura suspensa                    |
| POST /auth/login| SUBSCRIPTION_CANCELED    | 403  | Assinatura cancelada                   |
| POST /auth/login| TENANT_NOT_FOUND         | 404  | Tenant não encontrado                  |
| POST /auth/login| RATE_LIMIT_EXCEEDED      | 429  | Muitas tentativas                      |
| POST /auth/login| INTERNAL_ERROR           | 500  | Erro interno                           |
| POST /auth/login| SERVICE_UNAVAILABLE       | 503  | Redis ou serviço indisponível          |
| POST /auth/token| INVALID_REQUEST          | 400  | Campo code ausente                     |
| POST /auth/token| CODE_INVALID_OR_EXPIRED  | 401  | Code inválido, expirado ou usado       |
| POST /auth/token| INTERNAL_ERROR           | 500  | Erro interno                           |
| POST /auth/token| SERVICE_UNAVAILABLE      | 503  | Redis indisponível                     |
| POST /auth/refresh| INVALID_REQUEST        | 400  | Campos inválidos ou faltando           |
| POST /auth/refresh| INVALID_CREDENTIALS     | 401  | Refresh token inválido ou revogado     |
| POST /auth/refresh| SUBSCRIPTION_EXPIRED    | 402  | Assinatura vencida                     |
| POST /auth/refresh| RATE_LIMIT_EXCEEDED     | 429  | Muitas tentativas                      |
| POST /auth/refresh| INTERNAL_ERROR          | 500  | Erro interno                           |
| POST /auth/refresh| SERVICE_UNAVAILABLE      | 503  | Serviço indisponível                  |

---

## Exemplos de Implementação

### Handler GET /callback (Go)

```go
func (h *AuthCallbackHandler) Callback(c *gin.Context) {
    code := c.Query("code")
    state := c.Query("state")
    if code == "" || state == "" {
        c.AbortWithStatus(400)
        return
    }

    savedState, _ := c.Cookie("oauth_state")
    if savedState != state {
        c.AbortWithStatus(400) // CSRF
        return
    }
    c.SetCookie("oauth_state", "", -1, "/callback", "", true, true)

    body, _ := json.Marshal(map[string]string{"code": code})
    resp, err := h.httpClient.Post(authServerURL+"/api/v1/auth/token",
        "application/json",
        bytes.NewReader(body),
    )
    if err != nil || resp.StatusCode != 200 {
        c.Redirect(302, "/login?error=auth_failed")
        return
    }

    var out struct {
        AccessToken      string `json:"access_token"`
        ExpiresIn        int    `json:"expires_in"`
        RefreshToken     string `json:"refresh_token"`
        RefreshExpiresIn int    `json:"refresh_expires_in"`
    }
    json.NewDecoder(resp.Body).Decode(&out)

    c.SetCookie("wisa_access_token", out.AccessToken, out.ExpiresIn, "/", cookieDomain, true, true, true)
    c.Redirect(302, "/dashboard")
}
```

### Iniciar redirect para login (Go)

```go
state := hex.EncodeToString(cryptoRand(32))
c.SetCookie("oauth_state", state, 300, "/callback", "", true, true, true)
redirectURL := fmt.Sprintf("%s/login?tenant_slug=%s&product_slug=%s&state=%s",
    authServerURL, tenantSlug, productSlug, state)
c.Redirect(302, redirectURL)
```

### Fluxo SPA — Login (TypeScript)

```typescript
async function login(tenantSlug: string, productSlug: string, email: string, password: string): Promise<void> {
  const state = crypto.getRandomValues(new Uint8Array(32));
  const stateHex = Array.from(state).map(b => b.toString(16).padStart(2, '0')).join('');
  sessionStorage.setItem('oauth_state', stateHex);

  const res = await fetch(`${AUTH_SERVER_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
    body: JSON.stringify({
      tenant_slug: tenantSlug,
      product_slug: productSlug,
      user_email: email,
      password,
      state: stateHex,
    }),
  });

  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.message || 'Falha no login');
  }

  const { redirect_url } = await res.json();
  window.location.href = redirect_url;
}
```

### Fluxo SPA — Callback e troca de code

No callback (rota do frontend que recebe `?code=...&state=...`):

```typescript
const params = new URLSearchParams(window.location.search);
const code = params.get('code');
const state = params.get('state');
const savedState = sessionStorage.getItem('oauth_state');

if (!code || !state || state !== savedState) {
  // CSRF ou parâmetros inválidos
  window.location.href = '/login?error=invalid_callback';
  return;
}
sessionStorage.removeItem('oauth_state');

const res = await fetch(`${AUTH_SERVER_URL}/api/v1/auth/token`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ code }),
});

if (!res.ok) {
  window.location.href = '/login?error=auth_failed';
  return;
}

const { access_token, refresh_token, expires_in } = await res.json();
// Armazenar em memória ou enviar para backend setar cookie
// Redirecionar para app
```

### Interceptor com refresh (conceitual)

```typescript
// Pseudocódigo para interceptor HTTP (Angular/fetch)
async function handleRequest(url: string, options: RequestInit): Promise<Response> {
  let res = await fetch(url, { ...options, headers: { ...options.headers, 'Authorization': `Bearer ${accessToken}` } });

  if (res.status === 401 && !isRefreshOrLoginUrl(url)) {
    const refreshRes = await fetch(`${AUTH_SERVER_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token, tenant_slug, product_slug }),
    });

    if (refreshRes.ok) {
      const data = await refreshRes.json();
      accessToken = data.access_token;
      refreshToken = data.refresh_token;
      return handleRequest(url, options); // Retentar
    }
    if (refreshRes.status === 401 || refreshRes.status === 402) {
      redirectToLogin();
      return refreshRes;
    }
  }
  return res;
}
```

---

## Rate Limiting

Os endpoints `POST /api/v1/auth/login`, `POST /api/v1/auth/token` e `POST /api/v1/auth/refresh` estão sujeitos a **rate limiting** (ex.: por IP). Em caso de 429, o cabeçalho `Retry-After` indica segundos para nova tentativa.

Ver [ADR-007 — NGINX](../adrs/ADR-007-nginx-como-reverse-proxy.md) para configuração recomendada.

---

## Referências

- [ADR-010 — Fluxo Centralizado de Autenticação](../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [ADR-006 — JWT com Assinatura Assimétrica](../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [OAuth 2.0 Security BCP (RFC 9700)](https://datatracker.ietf.org/doc/html/rfc9700)
