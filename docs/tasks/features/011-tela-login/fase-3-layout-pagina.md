# Fase 3 — Layout da Página de Login

## Objetivo

Implementar o layout da página de login: container full-screen, gradient de fundo e overlay de textura, replicando o visual do protótipo Login-Wisa. Usar as variáveis CSS definidas na Fase 1.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O protótipo possui a seguinte estrutura de layout:
```html
<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-[#0f172a] via-[#1e3a8a] to-[#3b82f6] p-4 font-sans">
  <div class="absolute inset-0 bg-[url(...)] opacity-20 mix-blend-overlay pointer-events-none"></div>
  <Card class="...">...</Card>
</div>
```

Elementos principais:
- Container: `min-h-screen`, flex center, padding
- Gradient: `bg-gradient-to-br` (bottom-right) com 3 cores
- Overlay: `absolute inset-0`, texture com opacity 20%, `mix-blend-overlay`, `pointer-events-none`
- Card: posicionado com `relative z-10` para ficar acima do overlay

Em SCSS, o gradient é:
```scss
background: linear-gradient(to bottom right, var(--login-bg-start), var(--login-bg-mid) 50%, var(--login-bg-end));
```

### Ação 3.1

Estruturar o template do `LoginPageComponent`:

```html
<div class="login-page__container">
  <div class="login-page__overlay" aria-hidden="true"></div>
  <!-- Card será adicionado na Fase 4 -->
  <div class="login-page__card-placeholder">
    Card e formulário (Fase 4)
  </div>
</div>
```

E no `login-page.component.scss`:
```scss
.login-page__container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  position: relative;
  background: linear-gradient(
    to bottom right,
    var(--login-bg-start),
    var(--login-bg-mid) 50%,
    var(--login-bg-end)
  );
}

.login-page__overlay {
  position: absolute;
  inset: 0;
  background-image: var(--login-noise-url);
  opacity: 0.2;
  mix-blend-mode: overlay;
  pointer-events: none;
}
```

### Observação 3.1

O `:host` pode precisar de `display: block` para que o container ocupe a altura total. Adicionar no componente:
```scss
:host {
  display: block;
}
```
O `login-page__card-placeholder` será substituído pelo card real na Fase 4. Por ora, pode ter estilos mínimos para visualização (ex.: fundo branco translúcido) para validar o layout.

---

### Pensamento 2

A variável `--login-noise-url` deve estar definida. Se na Fase 1 não foi adicionado o asset `noise.svg`, o overlay pode ficar sem background e ser invisível. Alternativa: usar um fallback ou pattern CSS. Exemplo de fallback:
```scss
background-image: var(--login-noise-url, none);
```
Se `--login-noise-url` não estiver definida, não quebra. Ou usar um padrão CSS para grain:
```scss
// Fallback: subtle noise via CSS (não idêntico ao SVG, mas aceitável)
background-image: url("data:image/svg+xml,..."); // inline SVG de noise
```

Para simplicidade, assumir que `_login-tokens.scss` define `--login-noise-url: url('/images/noise.svg')` e o asset existe em `public/images/noise.svg`. Se não existir, o overlay será transparente — o gradient ainda será visível.

### Ação 3.2

Garantir que `_login-tokens.scss` (Fase 1) inclua:
```scss
--login-noise-url: url('/images/noise.svg');
```
Na Fase 3, se o asset não existir, criar um arquivo mínimo ou documentar que o overlay pode ser opcional. Para validação visual, um `rgba(0,0,0,0.02)` como overlay alternativo pode simular subtle grain se o SVG não estiver disponível.

### Observação 3.2

O protótipo usa URL externa. Para independência, hospedar localmente é preferível. O Angular serve arquivos de `public/` na raiz; `/images/noise.svg` resolve para `public/images/noise.svg`.

---

### Pensamento 3

Acessibilidade: o overlay é decorativo e não deve ser lido por screen readers. `aria-hidden="true"` no overlay está correto. O container principal deve permitir que o foco vá para os elementos interativos do formulário (Fase 4). O layout em si não introduce elementos interativos — apenas estrutura visual.

### Ação 3.3

Adicionar `aria-hidden="true"` ao overlay. O container pode ter `role="main"` para landmark, ou o componente de página já está implícito como main se for o único conteúdo. Para página de login, o conteúdo principal é o formulário — o container é apenas wrapper. Opcional: `<main class="login-page__container">` para semântica.

### Observação 3.3

Conformidade com WCAG: estrutura semântica adequada. O `<main>` será útil quando houver skip link ou múltiplas regiões.

---

### Pensamento 4

Responsividade: o protótipo usa `p-4` (16px padding). Em mobile, o padding evita que o card cole nas bordas. O `min-h-screen` garante altura completa em viewports variados. O flex center mantém o card centralizado vertical e horizontalmente. Não há breakpoints especiais no protótipo para o container — o design é fluido.

### Ação 3.4

Manter layout fluido. O card (Fase 4) terá `max-width` para não ficar excessivamente largo em telas grandes. O container já está adequado para responsividade básica.

### Observação 3.4

Sem necessidade de media queries na Fase 3 para o container. O card pode precisar de ajustes em Fase 5.

---

### Decisão final Fase 3

**Entregáveis:**
1. Container full-screen com gradient de fundo
2. Overlay de textura (com fallback se asset indisponível)
3. Centralização do conteúdo (flex)
4. Placeholder para o card (será substituído na Fase 4)
5. Estilos encapsulados no `login-page.component.scss`
6. Variáveis CSS usadas para cores (sem hardcode)
7. Validação visual: gradient e overlay visíveis ao acessar `/login`

---

### Checklist de Implementação

1. [ ] Adicionar `:host { display: block; }` no componente
2. [ ] Implementar `.login-page__container` com min-height, flex, padding, gradient
3. [ ] Implementar `.login-page__overlay` com position absolute, inset, opacity, mix-blend-mode
4. [ ] Garantir que `--login-noise-url` está definida em tokens (ou fallback)
5. [ ] Adicionar `aria-hidden="true"` ao overlay
6. [ ] Considerar `<main>` para container (semântica)
7. [ ] Validar visualmente em viewport desktop e mobile (redimensionar janela)
8. [ ] Verificar que `ng build` executa sem erros

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Protótipo Login-Wisa | Layout idêntico: gradient, overlay, centralização |
| Fase 1 | Uso das variáveis `--login-bg-*`, `--login-noise-url` |
| Acessibilidade | Overlay com aria-hidden, estrutura semântica |
| Guidelines | SCSS, sem inline styles críticos |

---

## Referências

- [Protótipo login.tsx](/Login-Wisa/client/src/pages/login.tsx)
- [fase-1-design-tokens-estilos.md](./fase-1-design-tokens-estilos.md)
