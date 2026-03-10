# Fase 1 — Revisão e Documentação do Processo de Autenticação

## Objetivo

Revisar a documentação em `docs/integration/auth-code-flow-integration.md` e documentar de forma clara e completa o processo de **Autenticação**, incluindo: redirect para a página de login, envio de credenciais via POST /api/v1/auth/login, implementação do callback e troca de code por token via POST /api/v1/auth/token. Para cada passo, descrever parâmetros de entrada, resultados (códigos HTTP e respostas) e processos.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O fluxo de autenticação tem múltiplas etapas que precisam ser entendidas em sequência. A documentação atual menciona "redirect para auth/login" mas não diferencia claramente:
- **Etapa A:** O sistema cliente redireciona o usuário para a **página de login** do auth (GET para a URL do frontend)
- **Etapa B:** O usuário preenche credenciais e o **frontend** envia POST /api/v1/auth/login
- **Etapa C:** O auth responde com redirect_url (ou 302); o usuário é redirecionado para o **callback** do cliente
- **Etapa D:** O backend do cliente recebe code e state, troca o code por token via POST /api/v1/auth/token

### Ação 1.1

Mapear o fluxo exato a partir do código real:
- O redirect inicial vai para `{AUTH_SERVER_URL}/login?tenant_slug=...&product_slug=...&state=...` (página Angular)
- O frontend exibe o formulário; ao submeter, chama POST /api/v1/auth/login com body JSON
- O backend retorna 200+JSON `{ redirect_url }` (se Accept: application/json) ou 302 (se form tradicional)
- O callback é uma rota do **cliente** (não do auth) — ex.: `https://cliente.dominio.com/product/callback?code=...&state=...`

### Observação 1.1

A documentação atual diz "auth/login" no diagrama — está correto como referência à página. O parâmetro correto na URL é `tenant_slug` e `product_slug` (não apenas tenant e product). Verificar consistência em todo o documento.

---

### Pensamento 2

O sistema cliente precisa saber exatamente:
1. **Como iniciar o login** — montar a URL de redirect com os parâmetros corretos
2. **Quais parâmetros** são obrigatórios na URL de redirect
3. **O que o POST /auth/login espera** — campos do body, Content-Type
4. **O que o POST /auth/login retorna** — sucesso (200/302), erros (401, 403, 404, 429, 500, 503)
5. **Como implementar o callback** — rota, query params, validação de state, chamada ao /auth/token
6. **O que o POST /auth/token espera e retorna**

### Ação 1.2

Consultar os DTOs e handler do backend para garantir precisão:
- LoginRequest: `tenant_slug`, `product_slug`, `user_email`, `password`, `state` (todos required exceto state?)
- Do login.go: `State` tem tag `json:"state"` sem binding required — o state é enviado pelo frontend e repassado na redirect_url
- Resposta de sucesso login: 200 + `{ "redirect_url": "..." }` ou 302
- TokenRequest: `code` (required)
- TokenResponse: `access_token`, `expires_in`, `refresh_token`, `refresh_expires_in`

### Observação 1.2

O `state` é obrigatório no fluxo OAuth para CSRF — deve ser enviado no POST /auth/login para que o backend inclua na redirect_url. O cliente gera o state antes do redirect e valida no callback. Documentar isso explicitamente.

---

### Pensamento 3

O formato de erro retornado pelo backend usa `status`, `error_code`, `message`, `details` (opcional). A doc atual na seção "Resposta 401" do /auth/token mostra `"code": "CODE_INVALID_OR_EXPIRED"` — o campo correto no JSON é `error_code`, não `code`. Preciso corrigir para evitar confusão.

### Ação 1.3

Verificar o AppError: o JSON serializado usa `error_code` e `status`. A documentação deve refletir o formato real:
```json
{
  "status": "error",
  "error_code": "CODE_INVALID_OR_EXPIRED",
  "message": "Código de autorização inválido ou expirado.",
  "details": ""
}
```

### Observação 1.3

Conformidade com o pkg/errors do projeto — a doc deve ser a fonte de verdade para desenvolvedores que integram.

---

### Pensamento 4

Riscos de segurança a documentar:
- **Open Redirect:** O backend monta redirect_url internamente — nunca aceitar URL externa. Já documentado, mas reforçar.
- **State CSRF:** O cliente DEVE validar o state no callback.
- **Code single-use:** O code expira em 40s e é removido após uso. Documentar o TTL.
- **Não logar o code** em produção.

### Ação 1.4

Incluir na documentação uma seção de "Segurança" ou reforçar nos passos:
- Nunca confiar em redirect_url vinda de parâmetro externo
- Sempre validar state no callback
- Code: TTL 40 segundos, uso único
- Não incluir code em logs ou URLs de erro

### Observação 1.4

Alinhado com ADR-010 e OWASP.

---

### Pensamento 5

O POST /auth/login pode retornar diversos erros mapeados no MapToAppError. A doc deve listar todos os códigos HTTP e error_codes possíveis para o login:
- 400: INVALID_REQUEST (dados inválidos)
- 401: INVALID_CREDENTIALS (credenciais inválidas — mensagem genérica)
- 403: ACCOUNT_LOCKED, USER_BLOCKED, PRODUCT_UNAVAILABLE, SUBSCRIPTION_SUSPENDED, SUBSCRIPTION_CANCELED
- 404: TENANT_NOT_FOUND, USER_NOT_FOUND (mas no mapper, TenantNotFound e UserNotFound em contexto auth podem mapear para INVALID_CREDENTIALS — verificar)
- 429: RATE_LIMIT_EXCEEDED
- 500: INTERNAL_ERROR
- 503: SERVICE_UNAVAILABLE (ex.: Redis indisponível)

### Ação 1.5

Consultar o mapper: ErrProductNotFound e ErrTenantNotFound em auth — o mapper retorna NewInvalidCredentials para ErrProductNotFound e NewTenantNotFound para ErrTenantNotFound. Ou seja, 404 para tenant, 401 genérico para product. Documentar a tabela completa de erros do login.

### Observação 1.5

Algumas respostas (ex.: TENANT_NOT_FOUND com 404) poderiam ajudar em debug, mas ADR-010 recomenda mensagem genérica para auth. O mapper atual mistura — documentar o que o backend realmente retorna.

---

## Decisão Final Fase 1

**Entregáveis para a documentação revisada:**

1. **Seção "Processo de Autenticação — Passo a Passo"** com:
   - **1.1 Iniciar login (redirect):** O sistema cliente gera `state`, armazena em cookie/session, redireciona o usuário para:
     ```
     GET {AUTH_SERVER_URL}/login?tenant_slug={tenant}&product_slug={product}&state={state}
     ```
     Parâmetros obrigatórios: `tenant_slug`, `product_slug`, `state`.

   - **1.2 Envio de credenciais (POST /api/v1/auth/login):** O frontend do auth (ou o cliente, se hospedar a tela) envia:
     - Request: `POST {AUTH_SERVER_URL}/api/v1/auth/login`
     - Body: `{ "tenant_slug", "product_slug", "user_email", "password", "state" }`
     - Content-Type: `application/json`
     - Se SPA: `Accept: application/json`
     - Respostas:
       - **200:** `{ "redirect_url": "https://..." }` — frontend executa `window.location.href = redirect_url`
       - **302:** Redirect direto (fluxo form tradicional) — Location: callback com code e state
       - **400, 401, 403, 404, 429, 500, 503:** JSON com `status`, `error_code`, `message` (ver tabela de erros)

   - **1.3 Implementação do callback:** O backend do cliente implementa GET `/callback` ou `/{product_slug}/callback`:
     - Query params recebidos: `code`, `state`
     - Validar state (comparar com valor armazenado)
     - Chamar POST {AUTH_SERVER_URL}/api/v1/auth/token com `{ "code": "..." }`
     - Se 200: extrair access_token, refresh_token; armazenar; redirecionar para app
     - Se 401: redirecionar para login

   - **1.4 Troca de code por token (POST /api/v1/auth/token):**
     - Request: `{ "code": "..." }`
     - Resposta 200: `{ "access_token", "expires_in", "refresh_token", "refresh_expires_in" }`
     - Respostas de erro: 400 (INVALID_REQUEST), 401 (CODE_INVALID_OR_EXPIRED), 503 (SERVICE_UNAVAILABLE), 500 (INTERNAL_ERROR)
     - Formato de erro: `{ "status": "error", "error_code": "...", "message": "...", "details": "..." }`

2. **Tabela de códigos de erro do login** (POST /auth/login)
3. **Tabela de códigos de erro do token** (POST /auth/token)
4. **Correção** do formato JSON de erro (usar `error_code` em vez de `code`)

---

## Checklist de Implementação (Documentação)

1. [ ] Criar/ajustar seção "Processo de Autenticação — Passo a Passo"
2. [ ] Documentar 1.1 — Redirect para login (URL, parâmetros)
3. [ ] Documentar 1.2 — POST /auth/login (request, response 200/302, erros)
4. [ ] Documentar 1.3 — Callback (handler, validação state, chamada /auth/token)
5. [ ] Documentar 1.4 — POST /auth/token (request, response, erros)
6. [ ] Adicionar tabela de erros do login
7. [ ] Adicionar tabela de erros do token
8. [ ] Corrigir formato JSON de erro (error_code)

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docs/code_guidelines | N/A (doc apenas) |
| ADR-010 | Fluxo Authorization Code, state CSRF, mensagem genérica |
| ADR-006 | JWT RS256, validação iss/aud/exp |
| Backend real | DTOs, mapper, handler consultados |

---

## Referências

- [auth-code-flow-integration.md](../../../integration/auth-code-flow-integration.md)
- [backend/internal/delivery/http/dto/login.go](../../../../backend/internal/delivery/http/dto/login.go)
- [backend/internal/delivery/http/dto/token.go](../../../../backend/internal/delivery/http/dto/token.go)
- [backend/internal/delivery/http/errors/mapper.go](../../../../backend/internal/delivery/http/errors/mapper.go)
- [backend/pkg/errors/codes.go](../../../../backend/pkg/errors/codes.go)
