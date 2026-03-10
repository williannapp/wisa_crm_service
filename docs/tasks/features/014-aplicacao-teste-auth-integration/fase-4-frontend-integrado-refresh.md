# Fase 4 — Frontend Integrado e Interceptor com Refresh

## Objetivo

Integrar o frontend Angular da test-app com o backend: consumir a rota GET /api/hello para exibir "Hello World" quando autenticado; implementar interceptor HTTP que, ao receber 401, chama o endpoint de refresh; evitar loop infinito; redirecionar para login quando refresh falhar.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O frontend da test-app precisa:
1. Ao carregar, verificar se o usuário está autenticado — chamar GET /api/hello (com credentials: include para enviar cookies)
2. Se 200: exibir "Hello World" (ou mensagem da API)
3. Se 401: redirecionar para GET /login do backend (que fará redirect para o auth)
4. O fluxo de login já está no backend (Fase 2); o frontend apenas precisa detectar 401 e redirecionar

### Ação 1.1

No componente principal (ex.: HomePage ou AppComponent):
- ngOnInit: chamar GET /api/hello (URL base configurável, ex.: /api/hello se o frontend é servido pelo backend ou usa proxy)
- Se 200: exibir "Hello World" (ou message do response)
- Se 401: window.location.href = '/login' (para que o backend processe o redirect ao auth)
- Tratar erros de rede (exibir mensagem ou redirect para login)

### Observação 1.1

O cookie é HttpOnly, então o JavaScript não tem acesso. O fetch com credentials: 'include' envia o cookie automaticamente. O backend valida e retorna 200 ou 401.

---

### Pensamento 2

O access token expira em 15 minutos. Quando o frontend chama GET /api/hello e o token está expirado, o backend retorna 401. Nesse momento, o cliente deve tentar refresh antes de redirecionar para login (doc § "Processo de Refresh — Passo a Passo").

O refresh é feito via POST {AUTH_SERVER_URL}/api/v1/auth/refresh com body:
```json
{
  "refresh_token": "<token>",
  "tenant_slug": "cliente1",
  "product_slug": "gestao-pocket"
}
```

O refresh_token está em cookie HttpOnly — o frontend não consegue lê-lo. Duas opções:
- (A) O backend da test-app expõe POST /api/auth/refresh: lê refresh_token do cookie, chama o auth server, atualiza os cookies, retorna 200
- (B) O frontend não faz refresh — apenas redireciona para login em 401

A opção (A) é mais completa e segue a doc. O backend da test-app atua como proxy: o frontend chama POST /api/auth/refresh (que envia o cookie), o backend troca com o auth, atualiza cookies, retorna sucesso. Então o frontend retenta a requisição original.

### Ação 1.2

Implementar no backend da test-app o endpoint POST /api/auth/refresh que:
1. Lê refresh_token do cookie wisa_refresh_token
2. Se ausente → 401
3. POST {AUTH_SERVER_URL}/api/v1/auth/refresh com body { refresh_token, tenant_slug, product_slug }
4. Se 200: atualiza cookies access_token e refresh_token; retorna 200
5. Se 401/402: remove cookies; retorna 401 (para o frontend redirecionar para login)
6. Se 429: repassar Retry-After; retornar 429

### Observação 1.2

Conforme doc §2.2: "Todos os campos são obrigatórios". O tenant_slug e product_slug vêm das variáveis de ambiente da test-app.

---

### Pensamento 3

Interceptor HTTP no Angular: quando qualquer requisição retorna 401, o interceptor deve:
1. Verificar se a requisição é para /api/auth/refresh ou /login — NÃO retentar (evitar loop)
2. Chamar POST /api/auth/refresh (com credentials: include)
3. Se 200: retentar a requisição original
4. Se 401 ou 402: redirecionar para /login
5. Se 429: exibir mensagem ou aguardar Retry-After

### Ação 1.3

Implementar AuthInterceptor no Angular:
- Usar HttpInterceptorFn (functional interceptor) ou classe HttpInterceptor
- Detectar 401 na resposta
- Verificar se a URL da requisição original é de refresh ou login → não interceptar
- Chamar refresh, se sucesso retentar com concatMap/switchMap
- Usar provideHttpClient(withInterceptors([authInterceptor])) no app.config

### Observação 1.3

Doc §2.5: "Se a requisição original for a própria chamada de refresh ou de login, não retentar — evitar loop infinito."

---

### Pensamento 4

Configuração de URL base da API: o frontend precisa saber para onde enviar as requisições. Em desenvolvimento com ng serve, o frontend roda em localhost:4201 e o backend em localhost:8081. Duas abordagens:
- (A) Proxy: configurar proxy.conf.json para /api → http://localhost:8081
- (B) URL absoluta: configurar API_BASE_URL = http://localhost:8081 e usar em todas as requisições
- (C) Se o backend serve o frontend (produção): relative URLs /api/hello funcionam

Para desenvolvimento, proxy é mais simples. Para produção (backend servindo frontend), as URLs relativas funcionam. Documentar ambos os cenários.

### Ação 1.4

Configurar:
- Em app.config: injection token para API_BASE_URL (ou usar relative /api quando servido pelo backend)
- proxy.conf.json para desenvolvimento: /api → http://localhost:8081
- HttpClient com withCredentials: true (ou fetch com credentials: 'include') para enviar cookies em cross-origin se necessário

### Observação 1.4

Quando frontend e backend estão no mesmo domínio (ex.: backend serve frontend em / e API em /api), cookies são enviados automaticamente. Em dev com portas diferentes, CORS deve permitir credentials e o frontend usa a URL do backend.

---

### Pensamento 5

Fluxo completo do usuário:
1. Usuário acessa http://localhost:4201 (ou 8081 se backend serve frontend)
2. Frontend carrega, chama GET /api/hello
3. Sem cookie → 401 → interceptor tenta refresh → 401 (sem refresh_token) → redirect /login
4. Backend /login redireciona para auth.wisa.labs.com.br
5. Usuário faz login no auth
6. Auth redireciona para /callback?code=&state=
7. Backend troca code por token, seta cookies, redirect /
8. Frontend carrega, chama GET /api/hello com cookie → 200 → exibe "Hello World"
9. Após 15 min, token expira; próxima requisição → 401 → interceptor chama refresh → 200 → retenta → sucesso

### Ação 1.5

Garantir que a rota raiz (/) do backend sirva o index.html do Angular (quando em modo integrado) ou que o frontend seja acessível. Se separados, o usuário acessa o frontend; o frontend chama a API no backend. O redirect para /login deve ir ao backend (onde está o handler). Se frontend em :4201 e backend em :8081, o link "Fazer Login" ou redirect em 401 deve apontar para http://localhost:8081/login.

### Observação 1.5

Em cenário com frontend e backend separados, o 401 pode incluir header ou body indicando a URL de login (ex.: Location: /login ou X-Login-URL: http://localhost:8081/login). O interceptor usa window.location.href = loginUrl para redirecionar. Ou assumir que a API está no mesmo origin (via proxy) e usar /login.

---

## Decisão Final Fase 4

**Entregáveis:**

1. **Backend: POST /api/auth/refresh**
   - Lê refresh_token do cookie
   - Chama auth server POST /auth/refresh
   - Atualiza cookies em caso de sucesso
   - Retorna 401 em falha (para redirect)

2. **Frontend: Componente principal**
   - Chama GET /api/hello ao carregar
   - Exibe "Hello World" quando 200
   - Usa interceptor para tratar 401

3. **Frontend: AuthInterceptor**
   - Intercepta 401
   - Exclui requisições para /api/auth/refresh e /login
   - Chama POST /api/auth/refresh
   - Se 200: retenta requisição original
   - Se 401/402: redirect para /login
   - withCredentials ou equivalente para cookies

4. **Configuração**
   - Proxy ou API_BASE_URL
   - Documentar fluxo completo no README

---

## Checklist de Implementação

1. [ ] Implementar POST /api/auth/refresh no backend
2. [ ] Implementar chamada GET /api/hello no frontend
3. [ ] Implementar AuthInterceptor
4. [ ] Configurar proxy ou API URL
5. [ ] Tratar redirect para /login em 401 após refresh falhar
6. [ ] Evitar loop infinito (não retentar refresh/login)
7. [ ] Testar fluxo completo: login → hello → esperar expiração → refresh → hello
8. [ ] Atualizar README com instruções

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| auth-code-flow-integration.md §2 | Processo de Refresh |
| auth-code-flow-integration.md §2.5 | Fluxo no Interceptor |
| auth-code-flow-integration.md §2.6 | Rotação, não localStorage |
| frontend.md §8 | Interceptors para 401 |
| ADR-002 | Não armazenar JWT em localStorage |

---

## Referências

- [docs/integration/auth-code-flow-integration.md](../../../integration/auth-code-flow-integration.md) § "Processo de Refresh"
- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
