# Fase 5 — Documentação de Integração para Aplicações Clientes

## Objetivo

Criar documentação que oriente as aplicações clientes (ex.: gestao-pocket, sistemas em Angular + backend) sobre como integrar com o fluxo de Authorization Code do wisa-crm-service. Inclui: implementação do GET /callback, troca de code por token, validação de state (CSRF) e boas práticas de segurança.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O wisa-crm-service atua como Identity Provider (IdP). As aplicações clientes são os RP (Relying Parties). O fluxo completo envolve:
1. Cliente redireciona usuário para login no auth (com tenant_slug, product_slug, state)
2. Usuário autentica no auth
3. Auth responde 302 para callback do cliente com code e state
4. Cliente recebe GET /callback com code e state
5. Backend do cliente troca code por token via POST /auth/token
6. Cliente armazena token e libera acesso

A documentação deve cobrir os passos 4, 5 e 6 do ponto de vista do cliente. Os passos 1–3 já estão implícitos no fluxo do auth.

### Ação 1.1

Criar documento em `docs/integration/` ou `docs/client-integration/` — ex.: `auth-code-flow-integration.md` — que descreva o fluxo do lado do cliente.

### Observação 1.1

O diretório `docs/` pode ter seção específica para integrações. Verificar estrutura existente. Se não existir, criar `docs/integration/auth-code-flow.md`.

---

### Pensamento 2

O backend da aplicação cliente deve implementar um endpoint GET /callback (ou /{product_slug}/callback conforme URL do redirect). Esse endpoint:
- Recebe query params: `code` e `state`
- Valida o `state` contra o valor armazenado antes do redirect (CSRF)
- Se state inválido, rejeitar com 400 ou redirect para login com erro
- Se válido, chamar o auth server POST /api/v1/auth/token com o code no body
- Receber access_token
- Definir como o cliente vai entregar o token ao frontend: cookie HttpOnly, ou redirect para SPA com token em fragment (menos seguro), ou resposta HTML que injeta token em memória

A abordagem mais segura: o backend do cliente define um cookie HttpOnly com o access_token e redireciona para a aplicação (ex.: /dashboard). O frontend não precisa ver o token — as requisições ao backend do cliente incluem o cookie automaticamente. O backend do cliente valida o JWT em cada request.

Alternativa: o backend retorna o token em um redirect com o token em um cookie. O cliente (browser) armazena o cookie. O frontend faz requisições ao backend do cliente; o backend pode validar o JWT ou delegar ao frontend (que envia Authorization header). Para SPAs que chamam APIs diretamente, o cookie HttpOnly no domínio do cliente funciona se as APIs estiverem no mesmo domínio.

### Ação 1.2

Documentar no guia de integração:

**Backend do cliente — GET /callback (ou /{product_slug}/callback):**
1. Extrair `code` e `state` da query string
2. Validar `state`: comparar com o valor armazenado em sessão (ou cookie seguro) antes do redirect para login. Se não bater, retornar 400 Bad Request ou redirect para página de erro
3. Fazer POST para `https://auth.wisa.labs.com.br/api/v1/auth/token` com body JSON `{ "code": "<code>" }`
4. Se resposta 401, redirecionar para login novamente
5. Se 200, extrair `access_token` e `expires_in`
6. Definir cookie HttpOnly + Secure + SameSite=Strict com o access_token (ou armazenar em sessão server-side)
7. Redirecionar para a aplicação (ex.: / ou /dashboard)

### Observação 1.2

A URL do auth server deve ser configurável (variável de ambiente) no cliente. O documento deve usar placeholder.

---

### Pensamento 3

Se o cliente for uma SPA pura (frontend apenas) sem backend próprio, o callback seria uma rota do frontend. Nesse caso:
- O redirect 302 levaria o browser para `https://cliente.com/callback?code=X&state=Y`
- A SPA carregaria e extrairia code e state da URL
- A SPA faria POST para o auth server... **Problema:** cross-origin. O auth server precisaria de CORS para aceitar requisições do domínio do cliente. Ou a SPA chama o próprio backend do cliente, que então chama o auth server (BFF — Backend for Frontend). O cenário descrito pelo usuário é "backend da aplicação CLIENTE" — logo, há um backend. O fluxo com backend é o correto.

### Ação 1.3

Documentar que o callback deve ser um **endpoint de backend** do cliente. Se o cliente for SPA sem backend, precisará de um BFF ou de um proxy que realize a troca server-side. O token **nunca** deve ser exposto no frontend via URL ou em respostas que possam ser logadas.

### Observação 1.3

O ADR-010 e as code guidelines do frontend recomendam não armazenar JWT em localStorage (XSS). Cookie HttpOnly é preferível.

---

### Pensamento 4

Validação do state (CSRF):
- Antes de redirecionar para o login, o cliente gera um token aleatório (ex.: 32 bytes hex)
- Armazena em sessionStorage ou em cookie com SameSite
- Inclui no redirect: `?state=<token>`
- No callback, compara o state recebido com o armazenado
- Remove o state armazenado após uso (one-time)

O **backend** do cliente pode não ter acesso ao sessionStorage (isso é do frontend). O fluxo típico: o frontend inicia o redirect para login, armazenando state em sessionStorage. Após o redirect, o browser carrega a URL do **backend** /callback. O backend não tem sessionStorage — o state precisaria estar em um cookie que o backend consiga ler. Ou: o frontend recebe o redirect em uma rota como `/auth/callback` (rota da SPA), lê code e state, valida state (tem acesso ao sessionStorage), e então faz um POST para o backend do cliente com o code, para o backend trocar por token e setar cookie. Nesse cenário, o code trafega do frontend para o backend — ainda aceitável se via HTTPS e POST.

Simplificação: o state pode ser validado pelo frontend se o callback for uma rota do frontend que então chama o backend. O backend recebe o code via POST (não na URL) do frontend, chama auth/token, e retorna o cookie. O frontend faz: `POST /api/auth/exchange-code` com `{ code }`. O backend valida... mas o backend não tem o state. O state precisa ser validado por quem o armazenou. Se o frontend armazenou, o frontend valida. Fluxo:
- Frontend: gera state, guarda em sessionStorage, redirect para auth
- Auth: redirect 302 para `https://cliente.com/callback?code=X&state=Y`
- **Quem recebe /callback?** Se for rota do frontend (SPA): frontend carrega, valida state, extrai code, chama `POST /api/auth/exchange-code { code }`, backend troca e seta cookie
- Se for rota do backend: o backend recebe a requisição GET com code e state. O state foi armazenado... pelo frontend. O backend não o tem. Conclusão: quando o callback é no backend, o state deveria ter sido passado em um cookie pelo frontend antes do redirect. Assim o backend pode validar.

Documentar ambos os fluxos: (A) Callback no frontend, valida state, chama backend para troca; (B) Callback no backend, state em cookie, backend valida.

### Ação 1.4

No guia, incluir:

**Geração e validação do state:**
- Frontend gera `state = crypto.getRandomValues(...)` (ou equivalente)
- Armazena em cookie `oauth_state` (HttpOnly, SameSite=Strict, path=/callback) ou sessionStorage
- Inclui no redirect para login
- No callback: compara state da URL com armazenado; se diferente, abortar
- Backend que recebe GET /callback: ler cookie oauth_state, comparar com state da URL

### Observação 1.4

O cookie para state deve ser acessível ao backend se o callback for no backend. Path=/callback ou path=/ para que a requisição GET /callback?code=&state= inclua o cookie.

---

### Pensamento 5

Resposta do POST /auth/token: `{ "access_token": "JWT", "expires_in": 900 }`. O cliente deve:
- Armazenar o token de forma segura (cookie HttpOnly no backend)
- Validar o JWT antes de conceder acesso (sig, iss, aud, exp) — ADR-010
- O cliente deve ter a chave pública do auth para validar a assinatura
- O `aud` deve corresponder ao domínio do cliente
- O `exp` define a validade

### Ação 1.5

Documentar no guia:
- Como obter a chave pública (endpoint JWKS ou arquivo)
- Validar assinatura RS256
- Validar iss, aud, exp
- expires_in: usar para renovação via refresh token (quando implementado)

### Observação 1.5

O ADR-006 descreve o JWKS endpoint. Incluir referência.

---

### Pensamento 6

Fluxo de erro: se o code for inválido ou expirado, o auth retorna 401. O cliente deve redirecionar o usuário para o login novamente, sem expor mensagens técnicas.

### Ação 1.6

Documentar tratamento de erro: 401 → redirect para login; 503 → mensagem "Serviço temporariamente indisponível, tente novamente"; 500 → mensagem genérica.

### Observação 1.6

Não logar o code em produção. Não incluir o code em URLs de erro.

---

### Pensamento 7

A documentação pode incluir exemplos de código (pseudocódigo ou snippets) para:
- Handler GET /callback em Go (ou linguagem comum dos clientes)
- Chamada HTTP POST para /auth/token
- Configuração de cookie
- Exemplo de validação de state

### Ação 1.7

Incluir seção "Exemplo de implementação" com snippets adaptáveis.

### Observação 1.7

Os clientes podem usar várias linguagens. Fornecer exemplo em Go (aliado ao backend do projeto) e referências para outras.

---

### Pensamento 8

Atualizar o ADR-010 se necessário para refletir que a documentação de integração do cliente está em docs/integration/. O ADR já descreve o fluxo; a documentação de integração é um guia passo a passo para implementadores.

### Ação 1.8

Incluir referência ao ADR-010 e à ATA-2025-03-05 no documento de integração. Atualizar ADR-010 na seção de referências para apontar ao novo guia, se fizer sentido.

### Observação 1.8

O ADR-010 já foi atualizado para o fluxo com code + Redis. O guia de integração complementa com detalhes de implementação.

---

## Estrutura do Documento de Integração

1. **Visão geral do fluxo** — diagrama ou descrição dos passos
2. **Pré-requisitos** — URL do auth, variáveis de ambiente
3. **Endpoint GET /callback** — parâmetros, validação de state, fluxo
4. **Troca de code por token** — POST /api/v1/auth/token, request/response
5. **Armazenamento do token** — cookie HttpOnly, segurança
6. **Validação do JWT** — iss, aud, exp, assinatura
7. **Tratamento de erros** — 401, 503, 500
8. **Exemplo de implementação** — snippets
9. **Referências** — ADR-010, ADR-006, ATA

---

## Checklist de Implementação

- [ ] 1. Criar diretório docs/integration/ (se não existir)
- [ ] 2. Criar auth-code-flow-integration.md com a estrutura acima
- [ ] 3. Documentar GET /callback com validação de state
- [ ] 4. Documentar POST /auth/token e formato da resposta
- [ ] 5. Documentar armazenamento seguro do token (cookie HttpOnly)
- [ ] 6. Incluir exemplos de código (Go e referências)
- [ ] 7. Documentar tratamento de erros
- [ ] 8. Adicionar referência no README do backend ou no TRACKER central

---

## Conclusão

A Fase 5 entrega a documentação necessária para que as aplicações clientes integrem corretamente ao fluxo de Authorization Code. O guia cobre segurança (state, cookie, validação de JWT) e boas práticas. Nenhum código é implementado no wisa-crm-service — apenas documentação.
