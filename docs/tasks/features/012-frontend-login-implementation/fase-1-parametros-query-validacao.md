# Fase 1 — Parâmetros de Query e Validação no Carregamento

## Objetivo

Implementar a leitura dos parâmetros `tenant_slug`, `product_slug` e `state` na rota `/login` e validar sua presença no carregamento da página. Caso algum parâmetro obrigatório esteja ausente, exibir mensagem de erro **sem enviar requisição ao backend**, conforme especificação da feature.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O sistema cliente redireciona o usuário para o auth central com query params na URL:
```
https://auth.wisa.labs.com.br/login?tenant_slug=cliente1&product_slug=gestao-pocket&state=abc123xyz
```

O frontend deve receber esses parâmetros e validá-los **antes** de permitir o uso do formulário. Se faltar algum, exibir mensagem de erro imediatamente — sem chamar o backend. Isso evita requisições desnecessárias e oferece feedback rápido ao usuário.

### Ação 1.1

Identificar como obter query params no Angular. O `ActivatedRoute` fornece `queryParams` (Observable) ou `snapshot.queryParams`. Para componente que carrega uma vez (login), `snapshot` é suficiente. Para reatividade (ex.: usuário altera URL), `queryParams` Observable é preferível.

```typescript
// Injetar ActivatedRoute
constructor(private route: ActivatedRoute) {}

// Ou com inject()
route = inject(ActivatedRoute);
params = this.route.snapshot.queryParams;
```

### Observação 1.1

O `ActivatedRoute` deve ser injetado no `LoginPageComponent`. A guidelines (frontend.md §3.3) recomenda `inject()` em vez de constructor injection. Usar `inject(ActivatedRoute)`.

---

### Pensamento 2

Os parâmetros esperados são:
- `tenant_slug` — obrigatório
- `product_slug` — obrigatório
- `state` — obrigatório (token CSRF para validação no callback — ADR-010)

Todos devem estar presentes e não vazios. Validação: `param && param.trim().length > 0`.

### Ação 1.2

Criar lógica de validação no componente. Usar Signals para estado:
- `tenantSlug`, `productSlug`, `state` — valores lidos da URL
- `queryParamsError` — signal com mensagem de erro ou `null` se válido
- `isParamsValid` — computed baseado na validação

```typescript
// Exemplo de estrutura
tenantSlug = signal('');
productSlug = signal('');
state = signal('');
queryParamsError = signal<string | null>(null);
isParamsValid = computed(() => !this.queryParamsError());
```

A validação deve rodar no `ngOnInit` ou em `effect()` que reage à rota. Como os query params vêm da rota, o ideal é validar no `ngOnInit` após ler o snapshot.

### Observação 1.2

Conforme guidelines §4, usar Signals para estado. O `computed` evita duplicação de lógica. A mensagem de erro deve ser genérica e amigável: "Parâmetros inválidos. Acesso deve ser feito através do sistema do cliente." — evita expor detalhes técnicos (ADR-002, segurança).

---

### Pensamento 3

Onde exibir a mensagem de erro? Duas opções:
- (A) Substituir o card do formulário por um bloco de erro
- (B) Exibir mensagem acima do card, mantendo o layout

A opção (A) é mais clara: se os params são inválidos, o formulário não faz sentido. O usuário não pode "corrigir" — ele precisa acessar via o link correto do sistema cliente.

### Ação 1.3

No template, usar `@if` para condicional:
- Se `queryParamsError()` tem valor → exibir mensagem de erro (card de erro ou alerta)
- Se `queryParamsError()` é null → exibir o formulário de login normalmente

O card de erro pode reutilizar o mesmo container visual (gradient, overlay) para consistência, com conteúdo diferente.

### Observação 1.3

Manter acessibilidade: o bloco de erro deve ter `role="alert"` e ser anunciado por leitores de tela. Incluir `data-testid="login-params-error"` para testes automatizados.

---

### Pensamento 4

O formulário atual usa `username` e `password`. O backend espera `user_email` (campo deve ser email). O label "Usuário" pode permanecer, mas o input deve ter `type="email"` e `autocomplete="username"` (ou `email`) para UX. A validação de email pode ser feita no submit ou via Validators. Para esta fase, o foco é apenas os query params — o formulário em si será tratado na Fase 3.

### Ação 1.4

Nesta fase, apenas garantir que o formulário **não seja exibido** quando os params forem inválidos. O botão "Entrar" e os campos ficam desabilitados/ocultos. Não alterar a estrutura do formulário nesta fase.

### Observação 1.4

A ordem de implementação: Fase 1 (params) → Fase 2 (serviço HTTP) → Fase 3 (submit + redirect). A Fase 1 isola a validação de query params.

---

### Pensamento 5

Risco de segurança: os parâmetros vêm da URL. Não confiar em valores para renderização de HTML dinâmico. Apenas exibir a mensagem de erro estática. Nunca usar `bypassSecurityTrustHtml` com valores da URL (ADR-002, XSS).

### Ação 1.5

A mensagem de erro deve ser string fixa no código, não interpolada com os valores da URL. Ex.: "Parâmetros inválidos. Acesso deve ser feito através do sistema do cliente."

### Observação 1.5

Conformidade com ADR-002 e code_guidelines frontend §9.1 (XSS).

---

## Decisão Final Fase 1

**Entregáveis:**
1. `LoginPageComponent` lê `tenant_slug`, `product_slug` e `state` de `ActivatedRoute.queryParams` (ou `snapshot.queryParams`)
2. Validação: os três parâmetros devem estar presentes e não vazios
3. Signal `queryParamsError`: `null` se válido, mensagem de erro se inválido
4. Template: se `queryParamsError()` tem valor → exibir bloco de erro (não exibir formulário)
5. Mensagem de erro genérica, sem expor detalhes técnicos
6. `data-testid="login-params-error"` no bloco de erro
7. `role="alert"` para acessibilidade

---

## Checklist de Implementação

1. [ ] Injetar `ActivatedRoute` no `LoginPageComponent` via `inject()`
2. [ ] Criar signals: `tenantSlug`, `productSlug`, `state`, `queryParamsError`
3. [ ] Implementar validação no `ngOnInit` (ou `effect`): ler snapshot, validar, setar `queryParamsError`
4. [ ] Criar `computed` `isParamsValid` se necessário para o template
5. [ ] Atualizar template: `@if (queryParamsError()) { ... } @else { ... formulário ... }`
6. [ ] Bloco de erro com `role="alert"` e `data-testid="login-params-error"`
7. [ ] Mensagem: "Parâmetros inválidos. Acesso deve ser feito através do sistema do cliente."
8. [ ] Testar: acessar `/login` sem params → exibir erro
9. [ ] Testar: acessar `/login?tenant_slug=x&product_slug=y&state=z` → exibir formulário

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docs/code_guidelines/frontend.md §4 | Signals para estado |
| docs/code_guidelines/frontend.md §3.3 | `inject()` para DI |
| ADR-002 §Mitigações | Mensagem genérica, sem detalhes técnicos |
| ADR-010 | State obrigatório (CSRF) |

---

## Referências

- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [docs/context.md](../../../context.md)
- [ADR-010 — Fluxo Centralizado de Autenticação](../../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [Angular ActivatedRoute](https://angular.dev/api/router/ActivatedRoute)
