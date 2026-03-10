# Fase 2 — Redirect, Callback e Token Exchange

## Objetivo

Implementar o fluxo completo de autenticação conforme docs/integration/auth-code-flow-integration.md: (1) redirect para login com state, (2) rota callback que recebe code e state, (3) troca de code por token via POST /auth/token, (4) armazenamento de tokens em cookie HttpOnly.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A ordem recomendada na doc (auth-code-flow-integration.md § "Ordem de Implementação Recomendada") é:
1. Redirect + state
2. Callback + token exchange
3. Armazenamento de token

O backend da test-app atua como **sistema cliente** do wisa-crm-service. O usuário não autenticado deve ser redirecionado para `https://auth.wisa.labs.com.br/login?tenant_slug=...&product_slug=...&state=...`.

### Ação 1.1

Implementar rota GET /login (ou /auth/redirect) no backend da test-app que:
1. Gera `state` — 32 bytes aleatórios em hex (crypto/rand)
2. Armazena `state` em cookie `oauth_state` (HttpOnly, SameSite=Strict, Path=/callback, duração 5 min)
3. Redireciona (HTTP 302) para:
   ```
   {AUTH_SERVER_URL}/login?tenant_slug={TENANT_SLUG}&product_slug={PRODUCT_SLUG}&state={state}
   ```
   Usar AUTH_SERVER_URL de variável de ambiente, valor padrão ou fixo: `https://auth.wisa.labs.com.br`

### Observação 1.1

Conforme doc §1.1: "Gerar state — token aleatório seguro (ex.: 32 bytes em hex)". O cookie deve ser HttpOnly e SameSite=Strict para prevenir CSRF (ADR-010). O Path=/callback garante que o cookie seja enviado apenas na rota de callback.

---

### Pensamento 2

Após login bem-sucedido, o auth server redireciona para o callback do cliente. A `redirect_url` é montada pelo auth a partir de tenant_slug e product_slug. O formato típico pode ser `{APP_URL}/callback` ou `{APP_URL}/{product_slug}/callback`. A doc menciona "GET /callback ou GET /{product_slug}/callback". Para a test-app, usar GET /callback é suficiente — o auth server precisa estar configurado para redirecionar para a URL da test-app.

### Ação 1.2

Garantir que APP_URL no .env permita que o auth server saiba o callback. O auth server monta a redirect_url internamente; a test-app deve usar um valor que o auth reconheça. Em ambiente de desenvolvimento, pode ser `http://localhost:8081/callback` — o auth pode ter whitelist por tenant. Documentar que o tenant de teste precisa ter a redirect_url configurada no auth.

### Observação 1.2

A doc (ADR-010) diz que o backend monta redirect_url a partir de tenant_slug e product_slug, possivelmente com whitelist. A test-app assume que existe configuração no auth para o tenant/produto de teste apontar para APP_URL/callback.

---

### Pensamento 3

Implementar GET /callback no backend da test-app. O handler deve:
1. Extrair `code` e `state` da query string
2. Se code ou state vazios → 400 ou redirect para /login?error=invalid_callback
3. Ler cookie `oauth_state`
4. Comparar state recebido com o armazenado — se diferente → 400 (CSRF)
5. Remover cookie oauth_state (one-time)
6. POST para {AUTH_SERVER_URL}/api/v1/auth/token com body `{"code": "<code>"}`
7. Se resposta 401 ou erro → redirect para /login?error=auth_failed
8. Se 200: parsear access_token, refresh_token, expires_in, refresh_expires_in
9. SetCookie para access_token (HttpOnly, Secure em prod, SameSite=Strict)
10. SetCookie para refresh_token (idem)
11. Redirect 302 para / (ou para a página principal do frontend)

### Ação 1.3

Implementar o handler conforme o exemplo em auth-code-flow-integration.md § "Handler GET /callback (Go)". Usar http.Client para o POST ao auth server. Cookies: nomes sugeridos `wisa_access_token` e `wisa_refresh_token`. Max-Age baseado em expires_in (em segundos).

### Observação 1.3

A doc recomenda "Cookie HttpOnly + Secure + SameSite=Strict". Em localhost, Secure=false pode ser necessário para HTTP. Usar Secure apenas quando APP_URL for HTTPS.

---

### Pensamento 4

O endpoint POST /api/v1/auth/token do auth server espera Content-Type: application/json. A URL base é AUTH_SERVER_URL; o path é /api/v1/auth/token (a doc usa /auth/token como abreviação). Verificar a URL exata no backend principal do wisa-crm-service.

### Ação 1.4

Consultar o backend principal: a rota é provavelmente `/api/v1/auth/token`. A URL completa: `{AUTH_SERVER_URL}/api/v1/auth/token`. Garantir que não haja double slash (ex.: se AUTH_SERVER_URL termina com /, concatenar corretamente).

### Observação 1.4

Conformidade com a documentação. O AUTH_SERVER_URL é `https://auth.wisa.labs.com.br` (sem barra final).

---

### Pensamento 5

Tratamento de erros no callback: a doc lista 400 (INVALID_REQUEST), 401 (CODE_INVALID_OR_EXPIRED), 500, 503. Em caso de erro, redirecionar para /login?error=auth_failed (ou error=code_expired) para o usuário tentar novamente.

### Ação 1.5

No handler do callback, ao receber status != 200 do POST /auth/token, fazer redirect para /login?error=auth_failed. Opcionalmente, preservar query param para exibir mensagem amigável no frontend ("Sessão expirada. Tente fazer login novamente.").

### Observação 1.5

Não expor detalhes do erro ao usuário (segurança). Mensagem genérica é suficiente.

---

## Decisão Final Fase 2

**Entregáveis:**

1. **Rota GET /login (ou /auth/redirect):**
   - Gera state (32 bytes hex)
   - SetCookie oauth_state (HttpOnly, SameSite=Strict, Path=/callback, MaxAge=300)
   - Redirect 302 para {AUTH_SERVER_URL}/login?tenant_slug=&product_slug=&state=

2. **Rota GET /callback:**
   - Extrai code e state da query
   - Valida state contra cookie
   - Remove cookie oauth_state
   - POST {AUTH_SERVER_URL}/api/v1/auth/token com {"code": "..."}
   - Em sucesso: SetCookie access_token e refresh_token; Redirect para /
   - Em erro: Redirect para /login?error=auth_failed

3. **Variáveis de ambiente:**
   - AUTH_SERVER_URL (default https://auth.wisa.labs.com.br)
   - TENANT_SLUG
   - PRODUCT_SLUG
   - APP_URL

4. **Integração com frontend:**
   - O frontend, ao detectar que não está autenticado, redireciona para GET /login do backend (ou exibe link "Fazer Login" que aponta para /login). O backend faz o redirect para o auth. Após callback, o usuário é redirecionado para / com os cookies setados.

---

## Checklist de Implementação

1. [ ] Implementar geração de state (crypto/rand)
2. [ ] Implementar rota GET /login com SetCookie e Redirect
3. [ ] Implementar rota GET /callback
4. [ ] Validação de state no callback
5. [ ] Chamada POST /api/v1/auth/token
6. [ ] SetCookie para access_token e refresh_token
7. [ ] Tratamento de erros (redirect para /login?error=)
8. [ ] Documentar fluxo no README da test-app

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| auth-code-flow-integration.md §1.1 | Redirect com state |
| auth-code-flow-integration.md §1.3 | Callback com validação state |
| auth-code-flow-integration.md §1.4 | Token exchange |
| auth-code-flow-integration.md §Armazenamento | Cookie HttpOnly |
| ADR-010 | State CSRF, fluxo Authorization Code |

---

## Referências

- [docs/integration/auth-code-flow-integration.md](../../../integration/auth-code-flow-integration.md)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
