# Fase 2 — Documentação do Processo de Refresh

## Objetivo

Documentar de forma clara e completa o processo de **Refresh Token** no `docs/integration/auth-code-flow-integration.md`: quando chamar, parâmetros de entrada, resultados (códigos HTTP e respostas) e fluxo recomendado no interceptor.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O refresh token é obtido na troca do code (POST /auth/token) e é usado quando o access token expira (15 minutos). O cliente deve chamar POST /api/v1/auth/refresh com refresh_token, tenant_slug e product_slug. A documentação atual já descreve isso, mas precisa:
- Deixar explícitos os parâmetros de entrada e tipos
- Documentar todos os códigos de retorno possíveis
- Corrigir o formato JSON de erro (error_code)
- Incluir o fluxo no interceptor de forma passo a passo

### Ação 1.1

Consultar RefreshRequest e handler:
- RefreshRequest: `refresh_token`, `tenant_slug`, `product_slug` — todos required
- Resposta de sucesso: mesma estrutura do TokenResponse (access_token, expires_in, refresh_token, refresh_expires_in)
- Erros: 400 (INVALID_REQUEST), 401 (INVALID_CREDENTIALS para token inválido), 402 (SUBSCRIPTION_EXPIRED para assinatura vencida), 429, 500, 503

### Observação 1.1

O mapper converte ErrRefreshTokenInvalid em NewInvalidCredentials (401). ErrSubscriptionExpired em NewSubscriptionExpired (402). Documentar ambos.

---

### Pensamento 2

O fluxo no interceptor precisa ser descrito com clareza: o cliente faz requisição com Authorization: Bearer {access_token}; se receber 401, tenta refresh; se refresh 200, retenta a requisição original; se refresh 401 ou 402, redireciona para login ou renovação.

### Ação 1.2

Documentar o fluxo no interceptor em passos numerados, incluindo:
- Detecção de 401 na API protegida
- Chamada a POST /auth/refresh (não usar o access_token na chamada de refresh)
- Tratamento de 200: atualizar tokens; retentar requisição
- Tratamento de 401: limpar tokens; redirecionar para login
- Tratamento de 402: redirecionar para renovação de plano
- Tratamento de 429: aguardar Retry-After; exibir mensagem

### Observação 1.2

Evitar loop infinito: se a requisição original for a própria chamada de refresh ou de login, não retentar. Documentar essa precaução.

---

### Pensamento 3

Rotação do refresh token: o retorno do POST /auth/refresh inclui novo refresh_token. O antigo é invalidado. O cliente DEVE substituir o token antigo pelo novo. Nunca armazenar refresh em localStorage (XSS). Preferir cookie HttpOnly ou server-side.

### Ação 1.3

Reforçar na doc as boas práticas de segurança para refresh token: rotação, armazenamento, e o que fazer em caso de 401 no refresh (possível token revogado ou roubado).

### Observação 1.3

Alinhado com ADR-010 (Refresh Token Rotation, reuse detection).

---

### Pensamento 4

Tabela de erros do POST /auth/refresh:
- 400: INVALID_REQUEST — dados inválidos (campos faltando)
- 401: INVALID_CREDENTIALS — token inválido, expirado ou revogado (mensagem genérica)
- 402: SUBSCRIPTION_EXPIRED — assinatura vencida
- 429: RATE_LIMIT_EXCEEDED
- 500: INTERNAL_ERROR
- 503: SERVICE_UNAVAILABLE

### Ação 1.4

Criar tabela completa na documentação com HTTP status, error_code, message e ação recomendada para o cliente.

### Observação 1.4

O formato de erro deve ser consistente: `{ "status": "error", "error_code": "...", "message": "...", "details": "..." }`.

---

## Decisão Final Fase 2

**Entregáveis para a documentação revisada:**

1. **Seção "Processo de Refresh — Passo a Passo"** com:
   - **2.1 Quando chamar:** Access token expirado (401 em requisição protegida)
   - **2.2 Request:** `POST {AUTH_SERVER_URL}/api/v1/auth/refresh`
     - Body: `{ "refresh_token", "tenant_slug", "product_slug" }`
     - Content-Type: `application/json`
   - **2.3 Resposta 200:** `{ "access_token", "expires_in", "refresh_token", "refresh_expires_in" }`
     - O novo refresh_token substitui o antigo
   - **2.4 Respostas de erro** (tabela)
   - **2.5 Fluxo no Interceptor** (passos numerados, incluindo precaução contra loop)
   - **2.6 Rotação e boas práticas de segurança**

2. **Tabela de erros do POST /auth/refresh**
3. **Correção** do formato JSON de erro (error_code)

---

## Checklist de Implementação (Documentação)

1. [ ] Criar/ajustar seção "Processo de Refresh — Passo a Passo"
2. [ ] Documentar request e response 200
3. [ ] Adicionar tabela de erros do refresh
4. [ ] Documentar fluxo no interceptor (passos)
5. [ ] Documentar rotação e boas práticas
6. [ ] Corrigir formato JSON de erro se ainda não corrigido na Fase 1

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-010 | Refresh Token Rotation, reuse detection |
| ADR-006 | Renovação com verificação de assinatura |
| Backend real | RefreshRequest, RefreshTokenUseCase |

---

## Referências

- [auth-code-flow-integration.md](../../../integration/auth-code-flow-integration.md)
- [backend/internal/delivery/http/dto/refresh.go](../../../../backend/internal/delivery/http/dto/refresh.go)
- [backend/internal/usecase/auth/refresh_token.go](../../../../backend/internal/usecase/auth/refresh_token.go)
