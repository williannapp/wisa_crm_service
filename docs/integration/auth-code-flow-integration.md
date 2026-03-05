# Integração com o Fluxo de Authorization Code — wisa-crm-service

Este documento orienta as aplicações clientes (ex.: gestao-pocket, sistemas em Angular + backend) sobre como integrar com o fluxo de Authorization Code do **wisa-crm-service** (Identity Provider centralizado).

---

## Visão Geral do Fluxo

```
┌─────────┐     ┌───────────────────┐     ┌─────────────────────────────────┐
│ Usuário │     │ Sistema Cliente   │     │   wisa-crm-service (IdP)          │
└────┬────┘     └────────┬──────────┘     └──────────────────┬────────────────┘
     │                   │                                  │
     │ 1. Acessa app     │                                  │
     ├──────────────────▶│                                  │
     │                   │ 2. Redirect 302                  │
     │◀──────────────────│ Location: auth/login?tenant=...&product=...&state=...
     │                   │                                  │
     │ 3. GET auth/login  │                                  │
     ├───────────────────┼─────────────────────────────────▶│
     │                   │                                  │ Exibe tela de login
     │ 4. Preenche email+senha                              │
     ├───────────────────┼─────────────────────────────────▶│
     │                   │                                  │ Valida credenciais
     │                   │                                  │ Gera code, armazena Redis
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
     │                   │ 8. { access_token, expires_in } │
     │                   │ Set-Cookie ou armazena em sessão  │
     │ 9. Acesso liberado │                                  │
     │◀──────────────────│                                  │
```

---

## Pré-requisitos

- **URL do auth server:** configurável via variável de ambiente (ex.: `AUTH_SERVER_URL=https://auth.wisa.labs.com.br`)
- **Chave pública do auth:** para validar a assinatura RS256 do JWT (endpoint JWKS ou arquivo)

---

## Implementação do Callback (Backend do Cliente)

O **backend** da aplicação cliente deve implementar um endpoint que recebe o redirect do auth:

- **Rota:** `GET /callback` ou `GET /{product_slug}/callback` (conforme URL montada pelo auth)
- **Query params:** `code`, `state`

### Passos no Handler GET /callback

1. **Extrair** `code` e `state` da query string.
2. **Validar `state`:** comparar com o valor armazenado em sessão (ou cookie seguro) antes do redirect para o login. Se não bater → retornar 400 Bad Request ou redirect para página de erro (possível CSRF).
3. **Trocar code por token:**
   - `POST {AUTH_SERVER_URL}/api/v1/auth/token`
   - Body JSON: `{ "code": "<code>" }`
   - Content-Type: `application/json`
4. **Se resposta 401** (code inválido/expirado): redirecionar usuário para login novamente.
5. **Se 200:** extrair `access_token` e `expires_in`.
6. **Armazenar token de forma segura:**
   - Cookie HttpOnly + Secure + SameSite=Strict com o access_token
   - Ou armazenar em sessão server-side
7. **Redirecionar** para a aplicação (ex.: `/` ou `/dashboard`).

### Geração e Validação do State (CSRF)

**Antes do redirect para o login:**

- Frontend gera `state = crypto.getRandomValues(...)` (32 bytes em hex)
- Armazena em cookie `oauth_state` (HttpOnly, SameSite=Strict, path adequado) ou sessionStorage
- Inclui no redirect: `?state=<token>`

**No callback:**

- Backend lê cookie `oauth_state` (se callback for no backend)
- Compara com `state` da URL
- Se diferente → abortar e retornar erro
- Remover o state após uso (one-time)

Se o callback for uma rota do **frontend** (SPA):

- Frontend valida o state (tem acesso ao sessionStorage)
- Frontend chama `POST /api/auth/exchange-code` com `{ code }`
- Backend do cliente troca code por token, seta cookie e redireciona

---

## Troca de Code por Token — POST /api/v1/auth/token

### Request

```http
POST /api/v1/auth/token
Content-Type: application/json

{
  "code": "64_hex_chars_received_in_callback_url"
}
```

### Resposta 200

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

### Resposta 401 (code inválido ou expirado)

O auth **não diferencia** "não encontrado", "expirado" ou "já usado" — resposta genérica:

```json
{
  "code": "CODE_INVALID_OR_EXPIRED",
  "message": "Código de autorização inválido ou expirado."
}
```

---

## Fluxo de Renovação (Refresh Token)

### Quando o Access Token Expira

O access token tem validade de 15 minutos. Quando expira, requisições à API retornam 401.

### Chamada ao Refresh

```http
POST {AUTH_SERVER_URL}/api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "<token recebido na troca de code>",
  "tenant_slug": "cliente1",
  "product_slug": "gestao-pocket"
}
```

### Respostas do Refresh

**200 OK**

```json
{
  "access_token": "...",
  "expires_in": 900,
  "refresh_token": "...",
  "refresh_expires_in": 604800
}
```

**401 Unauthorized**

Token inválido (expirado, revogado ou inexistente). Mensagem genérica. Redirecionar para login.

**402 Payment Required**

Assinatura do tenant vencida. Redirecionar para página de renovação de plano.

### Fluxo no Interceptor

1. Cliente faz requisição com `access_token` no header
2. API retorna 401
3. Cliente chama `POST /auth/refresh` com `refresh_token`, `tenant_slug`, `product_slug`
4. Se 200: atualiza tokens em memória; retenta requisição original
5. Se 401 ou 402: limpa tokens; redireciona para login

### Rotação do Refresh Token

O `refresh_token` retornado substitui o antigo; o antigo é invalidado. Sempre armazene apenas o refresh token mais recente.

### Boas Práticas de Segurança

- Sempre substituir o `refresh_token` antigo pelo novo retornado
- Nunca armazenar refresh_token em localStorage (vulnerável a XSS)
- Preferir cookie HttpOnly ou armazenamento server-side

### Rate Limiting

O endpoint `POST /auth/refresh` está sujeito a rate limiting por IP. Evite chamadas em excesso; em caso de 429, aguardar Retry-After antes de retentar.

---

## Armazenamento do Token

- **Cookie HttpOnly:** recomendado para SPAs que fazem requisições ao backend do cliente no mesmo domínio
- **Session server-side:** alternativa válida se o cliente for uma aplicação tradicional (servidor renderiza HTML)
- **Nunca:** localStorage, sessionStorage (vulnerável a XSS) — exceto se o token for apenas para memória de curta duração e CSP rigoroso

---

## Validação do JWT no Cliente

Antes de conceder acesso, a aplicação cliente **deve**:

1. **Validar assinatura RS256** usando a chave pública do wisa-crm-service
2. **Verificar `iss`** (issuer) — deve corresponder ao auth configurado
3. **Verificar `aud`** (audiência) — deve corresponder ao domínio do cliente
4. **Verificar `exp`** (expiração)
5. **Verificar `nbf`** (not before), se presente

Referência: [ADR-006 — JWT com Assinatura Assimétrica](../adrs/ADR-006-jwt-com-assinatura-assimetrica.md).

---

## Tratamento de Erros

| HTTP | Situação                          | Ação do Cliente                                                |
|------|-----------------------------------|----------------------------------------------------------------|
| 401  | Code inválido/expirado; refresh token inválido | Redirecionar usuário para login novamente                      |
| 402  | Assinatura vencida (no refresh)   | Redirecionar para renovação de plano                           |
| 503  | Serviço temporariamente indisponível | Exibir "Serviço temporariamente indisponível. Tente novamente." |
| 500  | Erro interno                      | Mensagem genérica; não expor detalhes                         |

- **Não logar** o `code` em produção
- **Não incluir** o code em URLs de erro

---

## Exemplo de Implementação (Go)

### Handler GET /callback

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

    resp, err := h.httpClient.Post(authServerURL+"/api/v1/auth/token",
        "application/json",
        strings.NewReader(`{"code":"`+code+`"}`),
    )
    if err != nil || resp.StatusCode != 200 {
        c.Redirect(302, "/login?error=auth_failed")
        return
    }

    var out struct {
        AccessToken string `json:"access_token"`
        ExpiresIn   int    `json:"expires_in"`
    }
    json.NewDecoder(resp.Body).Decode(&out)

    c.SetCookie("wisa_access_token", out.AccessToken, out.ExpiresIn, "/", cookieDomain, true, true, true)
    c.Redirect(302, "/dashboard")
}
```

### Iniciar redirect para login

```go
state := hex.EncodeToString(cryptoRand(32))
c.SetCookie("oauth_state", state, 300, "/callback", "", true, true, true)
redirectURL := fmt.Sprintf("%s/login?tenant_slug=%s&product_slug=%s&state=%s",
    authServerURL, tenantSlug, productSlug, state)
c.Redirect(302, redirectURL)
```

---

## Rate Limiting

O endpoint `POST /api/v1/auth/token` deve estar sob **rate limiting** no NGINX (ex.: 10 req/min por IP) para mitigar tentativas de adivinhação de codes. Ver [ADR-007 — NGINX](../adrs/ADR-007-nginx-como-reverse-proxy.md).

---

## Referências

- [ADR-010 — Fluxo Centralizado de Autenticação](../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [ADR-006 — JWT com Assinatura Assimétrica](../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [OAuth 2.0 Security BCP (RFC 9700)](https://datatracker.ietf.org/doc/html/rfc9700)
