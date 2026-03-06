# Fase 1 — Design Tokens e Estilos Base

## Objetivo

Definir os design tokens (cores, fontes, variáveis CSS) e estilos globais necessários para replicar o visual do protótipo Login-Wisa, em conformidade com `docs/code_guidelines/frontend.md` e ADR-002.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O protótipo Login-Wisa utiliza Tailwind CSS e possui um design específico:
- **Background gradient:** `from-[#0f172a] via-[#1e3a8a] to-[#3b82f6]` (slate-900 → blue-900 → blue-500)
- **Card:** `bg-white/95 backdrop-blur-md`, barra superior `from-blue-600 to-cyan-400`
- **Tipografia:** Fonte Inter (Google Fonts)
- **Cores de texto:** `slate-900`, `slate-700`, `slate-500`
- **Inputs:** `slate-50` background, `slate-200` border, `blue-500` focus ring
- **Botão primário:** `blue-600` com hover `blue-700`
- **Links:** `blue-600` com hover `blue-800`

O frontend do projeto usa **SCSS** (definido em `angular.json`). As guidelines não especificam Tailwind. Duas abordagens: (A) adicionar Tailwind ao Angular; (B) usar SCSS com variáveis CSS para replicar os valores. A abordagem B mantém o stack atual e não adiciona dependências — recomendada para consistência.

### Ação 1.1

Extrair os valores exatos do protótipo e documentá-los como variáveis CSS em `frontend/src/styles.scss` (ou em arquivo `_variables.scss` importado):

```scss
// Design tokens baseados no protótipo Login-Wisa
:root {
  // Gradiente de fundo
  --login-bg-start: #0f172a;
  --login-bg-mid: #1e3a8a;
  --login-bg-end: #3b82f6;

  // Card
  --login-card-bg: rgba(255, 255, 255, 0.95);
  --login-card-bar-start: #2563eb;
  --login-card-bar-end: #22d3ee;

  // Tipografia
  --login-title: #0f172a;
  --login-description: #64748b;
  --login-label: #334155;
  --login-link: #2563eb;
  --login-link-hover: #1d4ed8;

  // Inputs
  --login-input-bg: #f8fafc;
  --login-input-border: #e2e8f0;
  --login-input-focus: #3b82f6;

  // Botão
  --login-btn-bg: #2563eb;
  --login-btn-hover: #1d4ed8;
  --login-btn-shadow: rgba(37, 99, 235, 0.2);
}
```

### Observação 1.1

As variáveis usam valores hexadecimais equivalentes às classes Tailwind do protótipo. O `backdrop-blur` precisa de `backdrop-filter: blur()` no CSS — suportado em navegadores modernos. Validar compatibilidade com os targets do projeto.

---

### Pensamento 2

O protótipo usa a fonte **Inter** (Google Fonts). O `index.html` do protótipo inclui:
```html
<link href="https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap" rel="stylesheet">
```

O ADR-002 recomenda CSP via NGINX em produção. Para desenvolvimento, carregar Inter via `<link>` em `index.html` ou via `@import` no `styles.scss` é aceitável. Em produção, a CSP deve permitir `fonts.googleapis.com` e `fonts.gstatic.com` se usarmos Google Fonts.

### Ação 1.2

Adicionar carregamento da fonte Inter em `frontend/src/index.html`:

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
```

E em `styles.scss`:
```scss
$font-sans: 'Inter', sans-serif;
html, body {
  font-family: $font-sans;
}
```

### Observação 1.2

Usar `preconnect` melhora performance. O subset `400;500;600;700` reduz o tamanho do download comparado à variável completa. Conforme ADR-002, CSP em produção deve incluir `fonts.googleapis.com` e `fonts.gstatic.com` na configuração NGINX.

---

### Pensamento 3

O protótipo usa um overlay de textura: `bg-[url('https://grainy-gradients.vercel.app/noise.svg')] opacity-20 mix-blend-overlay`. Isso depende de um recurso externo. Para evitar dependência de URL externa e garantir funcionamento offline/isolado, duas opções: (A) baixar o SVG e hospedar em `assets/`; (B) usar um padrão de noise via CSS ou imagem local.

### Ação 1.3

Criar pasta `frontend/public/` (ou `frontend/src/assets/`) e incluir um arquivo de textura. O Angular usa `public/` para assets estáticos (conforme `angular.json`). Baixar o SVG de noise de uma fonte confiável e armazenar em `frontend/public/images/noise.svg` (ou usar um pattern CSS alternativo como `repeating-conic-gradient` para simular grain). Documentar no plano que o asset será adicionado na Fase 3; na Fase 1, apenas declarar a variável CSS:

```scss
--login-noise-url: url('/images/noise.svg');
```

### Observação 1.3

A URL `/images/noise.svg` assume que o asset está em `public/images/`. O `angular.json` já define `"input": "public"` para assets. Verificar que a pasta `public` existe; se não, criá-la. O grain é sutil (opacity 20%) — se o asset não estiver disponível, a página ainda será visualmente aceitável sem ele.

---

### Pensamento 4

As guidelines (seção 5.2) mencionam `NgOptimizedImage` para imagens estáticas. A textura de noise é decorativa e não é uma "imagem de conteúdo" — pode ser aplicada via `background-image` em CSS, não necessitando `NgOptimizedImage`. Nenhuma imagem de logotipo foi explicitamente mostrada no protótipo (o `logo` importado em `login.tsx` não aparece no JSX renderizado). O protótipo foca em: gradient, card, inputs, botão. Sem logo obrigatório na Fase 1.

### Ação 1.4

Não incluir logo/ícone de marca na Fase 1. Se o design evoluir para incluir logo, será tratado em fase posterior. Foco: tokens de cor, fonte, variáveis de gradiente e card.

### Observação 1.4

Conformidade com o escopo: apenas design da tela. Logo pode ser placeholder ou adicionado em refinamento futuro.

---

### Decisão final Fase 1

**Entregáveis:**
1. Arquivo `frontend/src/styles/_login-tokens.scss` (ou seção em `styles.scss`) com variáveis CSS do design
2. `frontend/src/index.html` atualizado com preconnect e link para Inter
3. `frontend/src/styles.scss` importando os tokens e aplicando `font-family` base
4. Documentar em comentários os valores equivalentes ao protótipo para rastreabilidade
5. Opcional: criar `frontend/public/images/` e adicionar `noise.svg` (ou deixar para Fase 3)

---

### Checklist de Implementação

1. [ ] Criar `frontend/src/styles/_login-tokens.scss` com variáveis CSS (gradiente, card, tipografia, inputs, botão)
2. [ ] Adicionar `@import` em `frontend/src/styles.scss` para o arquivo de tokens
3. [ ] Adicionar preconnect e link da fonte Inter em `frontend/src/index.html`
4. [ ] Aplicar `font-family: 'Inter', sans-serif` no `body` em `styles.scss`
5. [ ] Criar `frontend/public/images/` se não existir
6. [ ] Validar que `ng build` executa sem erros
7. [ ] Documentar decisão de não usar Tailwind (manter SCSS nativo)

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docs/code_guidelines/frontend.md | Uso de SCSS, sem menção a Tailwind — OK |
| ADR-002 | Tipografia e tokens preparados para tela de auth |
| Protótipo Login-Wisa | Valores de cores e gradientes extraídos e documentados |
| Segurança | Sem recursos externos inseguros; fontes Google com preconnect |

---

## Referências

- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [Protótipo login.tsx](/Login-Wisa/client/src/pages/login.tsx)
- [Protótipo index.css](/Login-Wisa/client/src/index.css)
