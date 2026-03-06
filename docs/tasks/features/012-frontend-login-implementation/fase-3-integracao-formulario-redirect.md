# Fase 3 — Integração do Formulário, Submit e Redirect

## Objetivo

Integrar o `AuthService` no `LoginPageComponent`, conectar o formulário ao método de login, executar o redirect via `window.location.href` quando o backend retornar sucesso e tratar erros (ex.: credenciais inválidas) exibindo mensagem genérica ao usuário.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O fluxo no submit:
1. Usuário preenche email e senha (tenant_slug, product_slug, state já vêm da URL — Fase 1)
2. Clica em "Entrar"
3. Frontend monta o body: `{ tenant_slug, product_slug, user_email, password, state }`
4. POST para `/api/v1/auth/login`
5. Sucesso (200): `window.location.href = response.redirect_url`
6. Erro (401, 400, 500): exibir mensagem "Credenciais inválidas" ou mensagem genérica

O formulário atual usa `username` e `password` como signals. O backend espera `user_email`. A tela de login deve ser alterada para considerar **Email** em vez de **Usuário**: renomear o signal para `userEmail`, alterar o label do campo para "Email" e garantir que o valor seja válido (formato email).

### Ação 1.1

No `LoginPageComponent`:
- Renomear o signal `username` para `userEmail` em todo o componente (template e classe)
- Alterar o label do campo de "Usuário" para "Email" na tela de login
- Validar formato email antes de enviar (opcional no frontend — o backend já valida com `binding:"required,email"`)

### Observação 1.1

O backend retorna 400 se `user_email` não for email válido. O frontend pode adicionar `type="email"` no input e `required` para validação nativa, ou usar `Validators.email` do Reactive Forms. Para simplicidade, a validação nativa (HTML5) pode ser suficiente nesta fase.

---

### Pensamento 2

Estado de loading: durante a requisição, desabilitar o botão "Entrar" e exibir indicador de carregamento (opcional). Isso evita múltiplos submits e melhora a UX.

### Ação 1.2

Adicionar signal `isSubmitting = signal(false)`. No início do submit: `isSubmitting.set(true)`. No final (sucesso ou erro): `isSubmitting.set(false)`. O botão deve ter `[disabled]="isSubmitting()"` e opcionalmente texto "Entrando..." quando loading.

### Observação 1.2

Em caso de redirect (sucesso), o `isSubmitting.set(false)` pode ser omitido pois a página será descarregada. Mas em caso de erro, é essencial liberar o botão para nova tentativa.

---

### Pensamento 3

Tratamento de erro: o `AuthService.login()` retorna `Observable<LoginResponse>`. O componente deve:
```typescript
this.authService.login(body).subscribe({
  next: (res) => {
    window.location.href = res.redirect_url;
  },
  error: (err) => {
    this.isSubmitting.set(false);
    this.loginError.set('Credenciais inválidas. Tente novamente.');
  }
});
```

Conforme ADR-002 e guidelines §8.2: **nunca** diferenciar "usuário não encontrado" de "senha incorreta". Sempre mensagem genérica: "Credenciais inválidas. Tente novamente."

### Ação 1.3

Criar signal `loginError = signal<string | null>(null)`. No `error` do subscribe, setar a mensagem. No template, exibir `loginError()` acima do botão ou dentro do card, com `role="alert"`. Limpar `loginError` ao iniciar novo submit (opcional, para não manter erro de tentativa anterior).

### Observação 1.3

O backend pode retornar diferentes códigos (INVALID_CREDENTIALS, ACCOUNT_LOCKED, SUBSCRIPTION_EXPIRED, etc.). Para o usuário, a mensagem deve ser genérica em casos de falha de autenticação. Para 500/503, pode usar "Serviço temporariamente indisponível. Tente novamente."

### Pensamento 4

O formulário só deve ser submetido se os query params forem válidos (Fase 1). O `onSubmit` deve checar `queryParamsError()` e, se houver erro, não prosseguir (o formulário já estará oculto, mas por segurança).

### Ação 1.4

No `onSubmit()`:
```typescript
if (this.queryParamsError()) return;
if (this.isSubmitting()) return;

const body: LoginRequest = {
  tenant_slug: this.tenantSlug(),
  product_slug: this.productSlug(),
  user_email: this.userEmail(), // ou username() se mantiver o nome
  password: this.password(),
  state: this.state(),
};

this.loginError.set(null);
this.isSubmitting.set(true);

this.authService.login(body).subscribe({
  next: (res) => { window.location.href = res.redirect_url; },
  error: (err) => {
    this.isSubmitting.set(false);
    this.loginError.set('Credenciais inválidas. Tente novamente.');
  }
});
```

### Observação 1.4

Usar `takeUntilDestroyed()` ou gerenciar subscription para evitar memory leaks. Em Angular 18, o `DestroyRef` pode ser usado. Ou usar `Subscription` e fazer `unsubscribe` no `ngOnDestroy`. Como o redirect descarrega a página no sucesso, o vazamento só ocorreria em caso de erro — ainda assim, é boa prática limpar.

Alternativa: usar `firstValueFrom` ou `lastValueFrom` com `async/await` e try/catch, dentro de um `effect` ou método. Ou `toSignal` com `result` — mas para operação com side effect (redirect), subscribe é mais direto.

### Pensamento 5

Alteração no template: o input de usuário deve ter `type="email"` para validação nativa e `autocomplete="username"` (ou `email`). O binding atual usa `username` — verificar se o componente usa `userEmail` ou `username` e alinhar.

### Ação 1.5

Atualizar o input e o label no template da tela de login:
- Alterar o label de "Usuário" para **"Email"**
- `id="user-email"` (ou `id="email"`) para acessibilidade
- `type="email"` para validação HTML5
- `[value]="userEmail()"` e `(input)="userEmail.set($any($event.target).value)"`
- `placeholder="Seu e-mail"` para deixar claro o formato esperado
- Manter `required` e `autocomplete="email"`

### Observação 1.5

A tela de login passa a solicitar explicitamente **Email** em vez de "Usuário", alinhando a interface ao contrato do backend (`user_email`) e evitando ambiguidade para o usuário.

---

### Pensamento 6

Segurança: o redirect usa `window.location.href` com URL retornada pelo backend. O backend monta a URL internamente (tenant_slug + product_slug + state) — não confia em URL externa (ADR-010, evita Open Redirect). O frontend apenas executa o redirect para a URL confiável retornada pelo backend.

### Ação 1.6

Não adicionar validação de URL no frontend — o backend é a fonte da verdade. O frontend confia na `redirect_url` retornada pelo backend autenticado.

### Observação 1.6

O risco de Open Redirect seria se o backend aceitasse `redirect_uri` como parâmetro do cliente. O ADR-010 especifica que o backend monta a URL internamente. O frontend está seguro ao seguir o redirect.

---

### Pensamento 7

Acessibilidade e testes: o bloco de erro de login deve ter `data-testid="login-error"` e `role="alert"`. O botão de submit já tem `data-testid="button-login"`. Em estado de loading, o botão pode ter `aria-busy="true"`.

### Ação 1.7

Incluir no checklist os data-testid e atributos de acessibilidade para o bloco de erro de login.

### Observação 1.7

Conformidade com a Fase 5 da feature 011 (componentes, acessibilidade).

---

## Decisão Final Fase 3

**Entregáveis:**
1. `LoginPageComponent` injeta `AuthService`
2. `onSubmit()` monta body com tenant_slug, product_slug, user_email, password, state (valores dos signals)
3. Chama `authService.login(body).subscribe(...)`
4. Sucesso: `window.location.href = res.redirect_url`
5. Erro: `loginError.set('Credenciais inválidas. Tente novamente.')`, `isSubmitting.set(false)`
6. Signal `isSubmitting` para desabilitar botão durante requisição
7. Signal `loginError` para exibir mensagem no template
8. Alterar tela de login para Email (em vez de Usuário): label "Email", signal `userEmail`, input `type="email"`, placeholder "Seu e-mail"
9. Bloco de erro com `role="alert"` e `data-testid="login-error"`
10. Gerenciamento de subscription (takeUntilDestroyed ou unsubscribe) para evitar memory leak

---

## Checklist de Implementação

1. [ ] Injetar `AuthService` no `LoginPageComponent` via `inject()`
2. [ ] Adicionar signals `isSubmitting`, `loginError`
3. [ ] Alterar tela de login para Email: renomear signal `username` → `userEmail`, label "Usuário" → "Email", placeholder "Seu e-mail"
4. [ ] Implementar `onSubmit()` com validação de query params e montagem do body
5. [ ] Chamar `authService.login(body).subscribe({ next, error })`
6. [ ] Sucesso: `window.location.href = res.redirect_url`
7. [ ] Erro: mensagem genérica, `isSubmitting.set(false)`
8. [ ] Template: exibir `loginError()` com `role="alert"` e `data-testid="login-error"`
9. [ ] Botão: `[disabled]="isSubmitting()"`, opcional "Entrando..." quando loading
10. [ ] Input de email: label "Email", `type="email"`, `required`, `autocomplete="email"`, `placeholder="Seu e-mail"`, `data-testid="input-email"`
11. [ ] Usar `takeUntilDestroyed()` ou `ngOnDestroy` para unsubscribe
12. [ ] Testar fluxo completo: params válidos + credenciais corretas → redirect
13. [ ] Testar fluxo de erro: credenciais inválidas → mensagem exibida

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docs/code_guidelines/frontend.md §4 | Signals para estado |
| docs/code_guidelines/frontend.md §8.2 | Mensagem genérica para erros de auth |
| ADR-002 §Mitigações | Nunca diferenciar usuário não encontrado vs senha incorreta |
| ADR-010 | Redirect para URL retornada pelo backend |
| docs/context.md | Fluxo de autenticação centralizado |

---

## Referências

- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [docs/context.md](../../../context.md)
- [ADR-010 — Fluxo Centralizado de Autenticação](../../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [Angular HttpClient](https://angular.dev/guide/http)
