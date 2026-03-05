# ATA — Reunião de Brainstorm: Redirect Seguro e Passagem de JWT

**Data:** 05 de março de 2025  
**Tema:** Redirect seguro do sistema de login para o software do cliente e passagem segura do token JWT  
**Participantes:** Equipe de arquitetura (brainstorm)

---

## Objetivo

Definir a solução para:

1. **Redirect seguro** — O backend do sistema de login deve montar e retornar a URL de redirect internamente (baseado em `tenant_slug` e `product_slug`), sem confiar em URLs externas para evitar Open Redirect.
2. **Passagem segura do JWT** — Evitar exposição do token na URL (query parameters), que traz riscos de vazamento por histórico do navegador e compartilhamento acidental.

**Contexto de domínios:**
- Sistema de login: `https://auth.wisa.labs.com.br`
- Aplicações clientes: `https://{tenant_slug}.wisa.labs.com.br` (ex: `https://lingerie-maria.wisa.labs.com.br`)

---

## Soluções Propostas

| # | Solução | Descrição |
|---|---------|-----------|
| 1 | Code temporário + PostgreSQL | Fluxo tipo OAuth: backend gera code, armazena em tabela PostgreSQL, retorna `redirect_url` com code. Cliente troca code por token via endpoint. Sem Redis. |
| 2 | **Code temporário + Redis** | Mesmo fluxo do item 1, porém usando Redis para armazenar os codes (mais alinhado ao padrão OAuth). |
| 3 | Cookie com Domain compartilhado | JWT (access token) entregue via cookie HttpOnly, com `Domain=.wisa.labs.com.br`. Validação obrigatória de `aud` no cliente. Somente para `*.wisa.labs.com.br`. |
| 4 | Redirect com Form POST | Página HTML com formulário oculto que faz POST do token para o callback do cliente (token no body, não na URL). |

**Solução escolhida:** Opção 2 — Code temporário + Redis.

---

## Conclusão: Solução Final a Implementar

### Visão geral

O fluxo segue o padrão **Authorization Code** do OAuth 2.0. Após autenticação bem-sucedida, o backend gera um **code temporário**, armazena-o no **Redis** (com TTL curto), retorna `redirect_url` com o code na query. O token JWT **nunca** aparece na URL. A aplicação cliente troca o code pelo token chamando o endpoint `POST /api/v1/auth/token`. O code é de uso único e expira rapidamente.

**Vantagens:** Funciona para qualquer domínio (inclusive clientes com domínio próprio); padronizado (OAuth); token nunca exposto na URL. **Infraestrutura:** Requer Redis.

### Fluxo detalhado

```
┌─────────┐     ┌───────────────────────┐     ┌─────────────────────────────────┐
│ Usuário │     │ Sistema do Cliente    │     │   wisa-crm-service (IdP)         │
│         │     │ (lingerie-maria...)   │     │   auth.wisa.labs.com.br          │
└────┬────┘     └───────────┬───────────┘     └─────────────────┬───────────────┘
     │                      │                                   │
     │ 1. GET /app          │                                   │
     ├─────────────────────▶│                                   │
     │                      │ 2. Sem token válido                │
     │ 3. Redirect 302       │    → Redirect para login           │
     │    auth...?tenant=X  │                                    │
     │    &product=Y&state=Z │                                    │
     │◀─────────────────────│                                    │
     │                      │                                    │
     │ 4. GET /login        │                                    │
     ├──────────────────────┼───────────────────────────────────▶│
     │                      │                                    │
     │ 5. Form login        │                                    │
     │    (email, senha)    │                                    │
     ├──────────────────────┼───────────────────────────────────▶│
     │                      │                                    │
     │                      │  6. Valida credenciais, assinatura  │
     │                      │  7. Gera JWT + code temporário     │
     │                      │  8. Armazena code no Redis (TTL ~2min)│
     │                      │  9. Monta redirect_url com ?code=X  │
     │                      │                                    │
     │ 10. JSON: {redirect_url: "https://...?code=abc&state=Z"}  │
     │◀─────────────────────┼────────────────────────────────────│
     │                      │                                    │
     │ 11. Frontend auth redireciona: window.location = redirect_url
     │     GET https://lingerie-maria.wisa.labs.com.br/gestao-pocket?code=abc&state=Z
     ├─────────────────────▶│                                    │
     │                      │ 12. Cliente extrai code da URL      │
     │                      │ 13. POST /api/v1/auth/token         │
     │                      │     Body: { "code": "abc" }          │
     │                      ├───────────────────────────────────▶│
     │                      │                                    │ 14. Busca code no Redis
     │                      │                                    │ 15. Retorna { "token": "JWT..." }
     │                      │                                    │ 16. Remove code do Redis (uso único)
     │                      │◀───────────────────────────────────│
     │                      │ 17. Cliente armazena token em memória│
     │ Acesso liberado      │     Valida JWT (sig, iss, aud, exp) │
     │◀─────────────────────│                                    │
```

### Especificação técnica da solução

#### 1. Resposta do endpoint de login em caso de sucesso

```json
{
  "redirect_url": "https://lingerie.wisa-labs.com/callback?code=XYZ&state=123"
}
```

- `code`: Token opaco, single-use, armazenado no Redis. TTL sugerido: 120 segundos (2 minutos).
- `state`: Repassado do parâmetro recebido no login; cliente valida no retorno (CSRF).

#### 2. Armazenamento do code no Redis

```text
Chave: auth_code:{code}
Valor: JSON com { "access_token": "...", "tenant_id": "...", "user_id": "..." }
  ou apenas { "token": "..." } se preferir armazenar o JWT diretamente
TTL: 120 segundos
```

O code é **removido** imediatamente após troca por token (uso único).

#### 3. Endpoint POST /api/v1/auth/token

**Request:**
```json
{
  "code": "abc123xyz"
}
```

**Response 200:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIs..."
}
```

**Erros:**
- `400` — code ausente ou inválido
- `401` — code expirado ou já utilizado

#### 4. Construção segura da redirect_url

A URL é montada internamente pelo backend:

```go
redirectURL := fmt.Sprintf("https://%s.wisa.labs.com.br/%s?code=%s&state=%s",
    tenantSlug, productSlug, code, state)
```

Validar que `tenant_slug` e `product_slug` geram uma URL permitida (whitelist ou regra de domínio).

#### 5. Integração do cliente

1. **Callback:** Página/rota que recebe o redirect (ex: `/gestao-pocket`, `/auth/callback`).
2. **Extrair code:** `const code = new URLSearchParams(window.location.search).get('code')`
3. **Trocar por token:** `POST https://auth.wisa.labs.com.br/api/v1/auth/token` com `{ "code": code }`
4. **Armazenar:** Token em memória (variável de serviço); nunca localStorage/sessionStorage.
5. **Limpar URL:** `history.replaceState` para remover `code` da barra de endereço (opcional, UX).
6. **Validação:** Validar JWT (sig, iss, aud, exp).

#### 6. Refresh Token

O refresh token pode ser retornado junto no endpoint `/api/v1/auth/token` (no corpo da resposta) ou em fluxo separado. O cliente usa o refresh token para renovar o access token quando expirado (conforme ADR-010).

#### 7. Escopo

Esta solução funciona para **qualquer domínio** — clientes em `*.wisa.labs.com.br` ou com domínio próprio (ex: `app.clienteexterno.com.br`). O code na URL é trocado por token via chamada server-side ou pelo frontend do cliente ao auth.

---

## Tarefas de implementação (resumo)

1. **Infraestrutura:** Adicionar Redis ao projeto (Docker, variáveis de ambiente).
2. **Backend wisa-crm-service:** Gerar code temporário após login; armazenar no Redis com TTL.
3. **Backend wisa-crm-service:** Retornar `redirect_url` com `?code=...&state=...` no JSON.
4. **Backend wisa-crm-service:** Implementar endpoint `POST /api/v1/auth/token` que troca code por token.
5. **Frontend auth (Angular):** Após login, redirect para `response.redirect_url`.
6. **Documentação para clientes:** Guia de integração — extrair code, chamar `/api/v1/auth/token`, armazenar token.
7. **ADR-010:** Atualizar para refletir o fluxo com code + Redis.

---

## Referências

- [docs/context.md](../../context.md)
- [ADR-006 — JWT com Assinatura Assimétrica](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [ADR-010 — Fluxo Centralizado de Autenticação](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
