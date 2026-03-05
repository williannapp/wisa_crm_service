# Fase 4 — Documentação de Integração: Fluxo de Refresh Token

## Objetivo

Documentar para as aplicações clientes como integrar com o endpoint `POST /api/v1/auth/refresh`, quando chamar, formato da requisição/resposta e tratamento de erros. Atualizar o documento de integração existente (`docs/integration/auth-code-flow-integration.md`) para incluir o fluxo de renovação de sessão.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O fluxo de refresh ocorre quando o access token expira (15 minutos). O cliente deve interceptar respostas 401 em requisições à API protegida, chamar o endpoint de refresh e, se bem-sucedido, retentar a requisição original com o novo token.

### Ação 4.1

Criar seção "Fluxo de Renovação (Refresh Token)" no documento de integração, ou criar documento complementar `docs/integration/refresh-token-integration.md` e referenciar no principal.

**Decisão:** Adicionar seção ao documento existente para manter coesão, seguindo o padrão da Feature 007.

### Observação 4.1

O documento auth-code-flow-integration já cobre o fluxo inicial. O refresh é extensão natural.

---

### Pensamento 2

Conteúdo a documentar:

1. **Quando chamar:** Ao receber 401 em requisição autenticada; ou proativamente antes do `expires_in` (ex.: renovar 1 min antes de expirar).
2. **Request:** `POST /api/v1/auth/refresh` com body JSON `{ "refresh_token", "product_slug", "tenant_slug" }`
3. **Response 200:** `{ "access_token", "expires_in", "refresh_token", "refresh_expires_in" }`
4. **Response 401:** Token inválido (expirado, revogado ou inexistente) — mensagem genérica, sem motivo específico
5. **Response 402:** Assinatura vencida (redirecionar para renovação de plano)
6. **Rotação:** O refresh_token retornado substitui o antigo; o antigo é invalidado
7. **Armazenamento:** Refresh token em memória ou cookie HttpOnly (nunca localStorage)

### Ação 4.2

Estrutura da seção no documento:

```markdown
## Fluxo de Renovação (Refresh Token)

### Quando o Access Token Expira

O access token tem validade de 15 minutos. Quando expira, requisições à API retornam 401.

### Resposta do POST /auth/token

A partir desta versão, o endpoint retorna também refresh_token:

{
  "access_token": "eyJ...",
  "expires_in": 900,
  "refresh_token": "abc123...",
  "refresh_expires_in": 604800
}

- access_token: JWT para Authorization header
- expires_in: segundos até expiração (900 = 15 min)
- refresh_token: token opaco para renovação
- refresh_expires_in: segundos até expiração do refresh (604800 = 7 dias)

### Chamada ao Refresh

POST {AUTH_SERVER_URL}/api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "<token recebido na troca de code>",
  "tenant_slug": "cliente1",
  "product_slug": "gestao-pocket"
}

### Respostas

**200 OK**
{ "access_token": "...", "expires_in": 900, "refresh_token": "...", "refresh_expires_in": 604800 }

**401 Unauthorized**
Token inválido (expirado, revogado ou inexistente). Mensagem genérica. Redirecionar para login.

**402 Payment Required**
Assinatura do tenant vencida. Redirecionar para página de renovação de plano.

### Fluxo no Interceptor (Backend/Frontend)

1. Cliente faz requisição com access_token no header
2. API retorna 401
3. Cliente chama POST /auth/refresh com refresh_token, tenant_slug, product_slug
4. Se 200: atualiza tokens em memória; retenta requisição original
5. Se 401 ou 402: limpa tokens; redireciona para login
```

### Observação 4.2

O tenant_slug e product_slug devem ser conhecidos pelo cliente — os mesmos usados no login. Podem vir de configuração ou da URL (subdomínio/rota).

---

### Pensamento 3

Rate limiting: o endpoint /refresh deve estar sob rate limiting no NGINX (similar ao /login). Documentar em docs/backend ou em vps-configurations.

### Ação 4.3

Adicionar nota no documento de integração:

> **Rate limiting:** O endpoint /auth/refresh está sujeito a rate limiting por IP. Evite chamadas em excesso; em caso de 429, aguardar Retry-After antes de retentar.

### Observação 4.3

Conformidade com ADR-007 (NGINX rate limiting).

---

### Pensamento 4

Segurança: o refresh token é rotativo. Se o mesmo refresh token for usado duas vezes (possível roubo), o wisa-crm-service pode invalida toda a família de tokens (ADR-010). Documentar que o cliente deve armazenar apenas o refresh token mais recente e usar apenas uma vez por renovação.

### Ação 4.4

Adicionar subseção "Boas práticas de segurança":

- Sempre substituir o refresh_token antigo pelo novo retornado
- Nunca armazenar refresh_token em localStorage (XSS)
- Preferir cookie HttpOnly ou armazenamento server-side

### Observação 4.4

Alinhado com OWASP e ADR-010.

---

## Checklist de Implementação

- [ ] 1. Adicionar seção "Fluxo de Renovação (Refresh Token)" em docs/integration/auth-code-flow-integration.md
- [ ] 2. Documentar request/response do POST /auth/refresh
- [ ] 3. Documentar códigos de status 200, 401, 402
- [ ] 4. Documentar que POST /auth/token agora retorna refresh_token e refresh_expires_in
- [ ] 5. Incluir nota sobre rate limiting
- [ ] 6. Incluir boas práticas de armazenamento do refresh token
- [ ] 7. Atualizar diagrama de sequência (opcional) com fluxo de refresh

---

## Referências

- docs/integration/auth-code-flow-integration.md
- ADR-010 — Fluxo de renovação, refresh token rotation
- ADR-007 — Rate limiting
