# Fase 2 — Serviço de Autenticação, Configuração de API e HttpClient

## Objetivo

Criar o serviço de autenticação que fará a chamada HTTP `POST /api/v1/auth/login`, configurar a URL base da API e garantir que o `HttpClient` esteja disponível na aplicação. O serviço deve seguir as guidelines (core/, inject(), responsabilidade única) e a estrutura Clean Architecture do frontend.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A especificação define:
- Endpoint: `POST /api/v1/auth/login`
- Body: `{ tenant_slug, product_slug, user_email, password, state }`
- Resposta esperada: `{ redirect_url: string }` — o frontend executa `window.location.href = redirect_url`

**Nota sobre o backend:** O backend atual responde com HTTP 302 (redirect). Para um SPA que usa `HttpClient.post()`, o fluxo esperado é resposta **200 + JSON** com `redirect_url`. O ADR-010 menciona: "O backend retorna {"redirect_url": "..."} em JSON; o frontend do auth apenas executa window.location.href = response.redirect_url". Se o backend ainda retorna 302, será necessário ajuste no backend para suportar resposta JSON (ex.: quando `Accept: application/json`). O planejamento assume o contrato JSON.

### Ação 1.1

Documentar o contrato esperado no serviço:
- Request: `LoginRequest { tenant_slug, product_slug, user_email, password, state }`
- Response: `LoginResponse { redirect_url: string }`
- O serviço retorna `Observable<LoginResponse>` ou usa `firstValueFrom` para Promise.

### Observação 1.1

O backend DTO usa `tenant_slug` (snake_case) no JSON. Manter snake_case no body para compatibilidade com o backend Go.

---

### Pensamento 2

Onde colocar o serviço? As guidelines (frontend.md §1.1) indicam:
- `core/auth/` — AuthService, AuthGuard, AuthInterceptor
- `core/http/` — api-config

O `AuthService` (ou `LoginService`) deve ficar em `core/auth/`. A URL base da API pode vir de `core/http/api-config.ts` ou de um `InjectionToken` (APP_CONFIG, API_BASE_URL).

### Ação 1.2

Criar estrutura:
- `frontend/src/app/core/http/api-config.ts` — constante ou InjectionToken para URL base (ex.: `/api` ou `https://api.example.com`)
- `frontend/src/app/core/auth/auth.service.ts` — método `login(credentials: LoginRequest): Observable<LoginResponse>`

Em desenvolvimento, a API pode estar em `http://localhost:8080`; em produção, via proxy ou variável de ambiente. O Angular usa `environment` ou `providers` para configurar.

### Observação 1.2

Para desenvolvimento local com backend em outra porta, configurar proxy no `angular.json` ou usar URL absoluta. O `ng serve` com proxy permite `/api` → `http://localhost:8080`. Verificar se já existe configuração de proxy no projeto.

---

### Pensamento 3

O `HttpClient` precisa ser fornecido na aplicação. O Angular 18 usa `provideHttpClient()` em `app.config.ts`. O `package.json` não lista `@angular/common/http` separadamente — ele vem com `@angular/common`. Basta adicionar `provideHttpClient()` aos providers.

### Ação 1.3

Atualizar `frontend/src/app/app.config.ts`:
```typescript
import { provideHttpClient } from '@angular/common/http';

export const appConfig: ApplicationConfig = {
  providers: [
    provideZoneChangeDetection({ eventCoalescing: true }),
    provideRouter(routes),
    provideHttpClient(),
  ],
};
```

### Observação 1.3

Sem `provideHttpClient()`, a injeção de `HttpClient` no serviço falhará. A feature 010 pode não ter incluído isso se não havia chamadas HTTP ainda.

---

### Pensamento 4

Interface do AuthService:
```typescript
export interface LoginRequest {
  tenant_slug: string;
  product_slug: string;
  user_email: string;
  password: string;
  state: string;
}

export interface LoginResponse {
  redirect_url: string;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private http = inject(HttpClient);
  private apiBase = inject(API_BASE_URL); // ou constante

  login(credentials: LoginRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>(
      `${this.apiBase}/api/v1/auth/login`,
      credentials,
      { withCredentials: true } // se necessário para cookies
    );
  }
}
```

O `withCredentials: true` é relevante se o backend setar cookies (ex.: CSRF). Para o fluxo atual (redirect_url em JSON), pode não ser necessário. Incluir por precaução se a documentação indicar.

### Ação 1.4

Implementar `AuthService` com método `login`. Usar `inject(HttpClient)`. A URL base pode ser um `InjectionToken` com valor default `''` (para requisições relativas quando há proxy) ou `environment.apiUrl`.

### Observação 1.4

Se a aplicação for servida no mesmo domínio do backend (ex.: NGINX faz proxy `/api` → backend), a URL pode ser relativa: `/api/v1/auth/login`. Se em portas diferentes (dev), usar proxy ou URL absoluta.

---

### Pensamento 5

Tratamento de erros: o serviço não deve tratar erros de forma específica — deixa para o componente ou um interceptor. O componente que chama `login()` vai usar `subscribe` ou `async` pipe e tratar `error` no callback. O AuthService pode apenas repassar o Observable.

Para erros HTTP (401, 400, 500), o backend retorna JSON com estrutura padronizada (pkg/errors). O frontend pode mapear para mensagem genérica: "Credenciais inválidas" para 401 (conforme ADR-002 e guidelines §8.2).

### Ação 1.5

O AuthService retorna o Observable puro. O tratamento de erro fica na Fase 3 (integração no componente). Nesta fase, apenas garantir que o serviço seja testável e que a assinatura esteja correta.

### Observação 1.5

Conformidade com guidelines §8.2: mensagem genérica para erros de autenticação.

---

### Pensamento 6

Testes: criar `auth.service.spec.ts` com `HttpClientTestingModule` ou `provideHttpClientTesting()`. Mock da requisição POST e verificação de que o serviço chama a URL correta com o body correto.

### Ação 1.6

Incluir no checklist a criação de teste unitário para o AuthService. O foco do planejamento é a implementação; os testes são parte da fase.

### Observação 1.6

Guidelines §7: TestBed com provideHttpClient e provideHttpClientTesting.

---

## Decisão Final Fase 2

**Entregáveis:**
1. `provideHttpClient()` adicionado em `app.config.ts`
2. `API_BASE_URL` ou equivalente (InjectionToken ou constante) para URL base da API
3. Interfaces `LoginRequest` e `LoginResponse` (em `core/auth/` ou `domain/models/`)
4. `AuthService` em `core/auth/auth.service.ts` com método `login(credentials: LoginRequest): Observable<LoginResponse>`
5. Chamada `POST {apiBase}/api/v1/auth/login` com body em snake_case
6. Serviço `providedIn: 'root'`
7. Uso de `inject(HttpClient)` e `inject(API_BASE_URL)` conforme guidelines
8. (Opcional) Teste unitário do AuthService

---

## Checklist de Implementação

1. [ ] Adicionar `provideHttpClient()` em `app.config.ts`
2. [ ] Criar `API_BASE_URL` InjectionToken ou configurar em environment
3. [ ] Criar interfaces `LoginRequest` e `LoginResponse`
4. [ ] Criar `AuthService` em `core/auth/auth.service.ts`
5. [ ] Implementar `login()` com `HttpClient.post<LoginResponse>()`
6. [ ] Garantir que o body usa snake_case (tenant_slug, product_slug, user_email, password, state)
7. [ ] Verificar proxy ou CORS para desenvolvimento local
8. [ ] (Opcional) Criar `auth.service.spec.ts` com teste de chamada POST

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docs/code_guidelines/frontend.md §1.1 | AuthService em core/auth/ |
| docs/code_guidelines/frontend.md §3.1 | Responsabilidade única, providedIn: 'root' |
| docs/code_guidelines/frontend.md §3.3 | inject() para DI |
| docs/code_guidelines/frontend.md §6 | Nomenclatura AuthService, LoginRequest |
| ADR-005 (Clean Architecture) | Serviço na camada application/infrastructure |

---

## Pré-requisito Backend (se aplicável)

Se o backend retornar HTTP 302 em vez de JSON, será necessário:
- Adicionar suporte a resposta JSON no handler de login (ex.: quando `Accept: application/json`)
- Ou alterar o handler para sempre retornar 200 + `{"redirect_url": "..."}` para requisições XHR

O frontend planejado assume resposta 200 + JSON.

---

## Referências

- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [backend/internal/delivery/http/dto/login.go](../../../../backend/internal/delivery/http/dto/login.go)
- [docs/backend/README.md](../../../backend/README.md) — contrato do endpoint
- [ADR-010 — Fluxo Centralizado de Autenticação](../../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
