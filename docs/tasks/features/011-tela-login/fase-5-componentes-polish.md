# Fase 5 — Componentes Visuais, Responsividade e Acessibilidade

## Objetivo

Finalizar a tela de login com ícones (se não incluídos na Fase 4), responsividade em diferentes viewports, acessibilidade (ARIA, labels, navegação por teclado) e atributos `data-testid` para testes automatizados. Garantir conformidade com ADR-002 (segurança) e guidelines de acessibilidade.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O protótipo possui `data-testid` nos elementos interativos:
- `data-testid="input-username"`
- `data-testid="input-password"`
- `data-testid="button-toggle-password"`
- `data-testid="button-login"`

Esses atributos facilitam testes E2E (Cypress, Playwright) e testes unitários com `TestingLibrary` que usam `getByTestId`. Incluir na implementação.

### Ação 5.1

Adicionar `data-testid` em todos os elementos interativos do formulário:

```html
<input data-testid="input-username" ... />
<input data-testid="input-password" ... />
<button type="button" data-testid="button-toggle-password" ... >
<button type="submit" data-testid="button-login" ... >
```

### Observação 5.1

Conformidade com boas práticas de testabilidade. Os nomes devem ser consistentes com o protótipo para facilitar migração de testes existentes (se houver).

---

### Pensamento 2

Acessibilidade (WCAG 2.1):
- Todos os inputs devem ter `<label>` associado via `for`/`id` ou wrapping
- O botão de toggle de senha deve ter `aria-label` descritivo ("Mostrar senha" / "Ocultar senha")
- O formulário deve ter `aria-describedby` ou `aria-label` se o contexto não for óbvio
- O título da página deve ser adequado (`<h1>` ou role="heading")
- Contraste de cores: verificar que texto e fundo atendem AA (4.5:1 para texto normal)
- Navegação por teclado: Tab order lógico, Enter para submit

O protótipo usa `<Label htmlFor="username">` e `<Label htmlFor="password">`. Garantir que os inputs tenham `id="username"` e `id="password"` correspondentes.

### Ação 5.2

Checklist de acessibilidade:
1. Labels: `<label for="username">Usuário</label>` e `<input id="username" ...>`
2. Toggle senha: `[attr.aria-label]="showPassword() ? 'Ocultar senha' : 'Mostrar senha'"`
3. Título: `<h1 class="login-page__card-title">` — único h1 na página
4. Link "Esqueceu a senha?": adicionar `aria-label="Esqueceu a senha?"` se o texto visível não for suficiente (o texto já é descritivo)
5. Foco visível: garantir que `:focus-visible` tenha outline adequado nos inputs e botões
6. Contraste: as variáveis do protótipo (slate-900 sobre branco, blue-600 sobre branco) atendem WCAG AA

### Observação 5.2

O Angular escapa output automaticamente — sem risco de XSS. Não usar `bypassSecurityTrust*` em nenhum conteúdo dinâmico. Conforme ADR-002 e guidelines §8.

---

### Pensamento 3

Responsividade: o protótipo usa `p-4` no container e `max-w-md` no card. Em mobile, o card deve ter margens adequadas. Em telas muito pequenas (< 320px), o formulário pode precisar de ajustes. Breakpoints típicos:
- Mobile: < 640px
- Tablet: 640px - 1024px
- Desktop: > 1024px

O layout atual (flex center, padding 1rem, card max-width 28rem) deve funcionar bem. Verificar que:
- Em 320px de largura, o card não overflow horizontal
- Em altura pequena (ex.: 568px), o conteúdo não fica cortado — scroll se necessário
- Inputs e botão têm tamanho touch adequado (min 44x44px) em mobile

### Ação 5.3

Ajustes SCSS para responsividade:
```scss
.login-page__container {
  padding: 1rem;
  @media (min-width: 640px) {
    padding: 1.5rem;
  }
}

.login-page__card {
  margin: 0 1rem;
  @media (min-width: 640px) {
    margin: 0;
  }
}

// Inputs: min-height para touch targets
.login-page__input {
  min-height: 44px;
}
```

### Observação 5.3

O `min-height: 44px` nos inputs atende recomendação de touch targets (Apple HIG, Material). O padding do container garante que o card nunca cole nas bordas.

---

### Pensamento 4

O link "Fale conosco" e "Esqueceu a senha?" com `href="#"` causam scroll para o topo ao clicar em alguns navegadores. Para evitar, usar `(click)="$event.preventDefault()"` ou `routerLink` com rota inexistente e preventDefault. Como não há rotas para essas ações, manter `href="#"` e `(click)="$event.preventDefault()"` é suficiente. Ou usar `role="button"` com `(click)` se forem tratados como botões — mas semanticamente são links (levariam a outra página). Manter como links com href placeholder.

### Ação 5.4

Para links placeholder:
```html
<a href="#" (click)="$event.preventDefault()" class="login-page__link">Esqueceu a senha?</a>
<a href="#" (click)="$event.preventDefault()" class="login-page__link">Fale conosco</a>
```

Ou, se o projeto usar routerLink para rotas futuras:
```html
<a routerLink="/forgot-password" class="login-page__link">Esqueceu a senha?</a>
```
Como a rota não existe, resultaria em 404. Para esta feature (design apenas), `href="#"` + preventDefault é adequado.

### Observação 5.4

Sem impacto em segurança. Os links são claramente placeholder. Em feature futura de "Esqueci minha senha", substituir por rota real.

---

### Pensamento 5

Validação visual de formulário: mesmo sem lógica de auth, pode ser útil mostrar estados de validação (ex.: required, minlength) para preparar a UI. Porém, o escopo é "design apenas". Incluir `required` nos inputs é HTML nativo e não implementa lógica — apenas dá feedback do browser. Adicionar `required` nos inputs é adequado e melhora UX.

### Ação 5.5

Adicionar `required` nos inputs de usuário e senha. O botão de submit acionará validação nativa do HTML5 se o formulário tiver `action` ou o Angular Forms estiver configurado. Para formulário sem FormsModule, o `required` ainda exibe mensagem do browser ao submit. Se usarmos FormsModule com validação, seria mais rico — mas para design, `required` é suficiente.

### Observação 5.5

Não expor mensagens que diferenciem "usuário não existe" vs "senha errada". Nesta fase, nenhuma mensagem de erro de auth é exibida — apenas validação HTML5 (campo obrigatório). OK.

---

### Pensamento 6

O `autocomplete` nos inputs: para campo de login, `autocomplete="username"` e `autocomplete="current-password"` ajudam gerenciadores de senha e preenchem credenciais salvas. Isso melhora UX e é recomendado pelo WCAG (3.3.7 - Redução de erros de digitação).

### Ação 5.6

Adicionar atributos autocomplete:
```html
<input id="username" autocomplete="username" ... />
<input id="password" autocomplete="current-password" ... />
```

### Observação 5.6

O `autocomplete="username"` pode ser "email" se o backend esperar email. O protótipo diz "Usuário" — "username" é genérico e aceitável. Quando integrar com backend (email), alterar para `autocomplete="email"`.

---

### Decisão final Fase 5

**Entregáveis:**
1. `data-testid` em input-username, input-password, button-toggle-password, button-login
2. Labels associados aos inputs (for/id)
3. aria-label no toggle de senha (dinâmico)
4. Focus visible nos elementos interativos (outline)
5. Responsividade: padding, margin, min-height touch targets
6. Links com preventDefault para evitar scroll
7. `required` nos inputs
8. `autocomplete="username"` e `autocomplete="current-password"`
9. Validação final: build, testes manuais de acessibilidade (navegação por teclado, screen reader básico)
10. Documentação no README ou na feature sobre como testar a tela

---

### Checklist de Implementação

1. [ ] Adicionar `data-testid` em todos os elementos interativos
2. [ ] Garantir labels associados (for/id) nos inputs
3. [ ] Adicionar `aria-label` dinâmico no toggle de senha
4. [ ] Verificar/ajustar `:focus-visible` para outline visível
5. [ ] Implementar media queries para responsividade (padding, margin)
6. [ ] Adicionar `min-height: 44px` nos inputs (touch targets)
7. [ ] Adicionar `(click)="$event.preventDefault()"` nos links placeholder
8. [ ] Adicionar `required` nos inputs
9. [ ] Adicionar `autocomplete="username"` e `autocomplete="current-password"`
10. [ ] Executar `ng build --configuration production` e validar
11. [ ] Testar navegação por Tab e Enter
12. [ ] Verificar contraste com ferramenta (ex.: Lighthouse, axe)
13. [ ] Atualizar TRACKER da feature 011 com status das fases

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Protótipo | data-testid compatíveis |
| WCAG 2.1 AA | Labels, aria-label, focus, contraste |
| Guidelines §7 | Estrutura para testes (data-testid) |
| ADR-002 | Sem mensagens que exponham dados de auth |
| Mobile | Touch targets 44px, layout responsivo |

---

## Referências

- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [ADR-002 — Angular e Segurança](../../../adrs/ADR-002-angular-como-framework-frontend.md)
- [WCAG 2.1](https://www.w3.org/WAI/WCAG21/quickref/)
- [Protótipo login.tsx](/Login-Wisa/client/src/pages/login.tsx)
