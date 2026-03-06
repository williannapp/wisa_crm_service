# Fase 2 — Estrutura e Rota da Página de Login

## Objetivo

Criar o componente `LoginPage` na estrutura de features do projeto e configurar a rota lazy-loaded `/login`, em conformidade com `docs/code_guidelines/frontend.md` (organização por features, lazy loading) e ADR-002.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

As guidelines definem a estrutura:
```
src/app/
├── features/
│   ├── auth/
│   │   ├── login/
│   │   │   ├── login-page.component.ts
│   │   │   ├── login-page.component.html
│   │   │   └── login-page.component.spec.ts
```

O componente deve seguir convenções: `LoginPageComponent`, selector `app-login-page`, arquivos em kebab-case. A feature 010 criou a pasta `features/auth/` — verificar se existe. Se não existir, criar `features/auth/login/`.

### Ação 2.1

Criar o componente usando Angular CLI:
```bash
ng generate component features/auth/login/login-page --standalone --skip-tests
```

Ou manualmente criar os arquivos:
- `frontend/src/app/features/auth/login/login-page.component.ts`
- `frontend/src/app/features/auth/login/login-page.component.html`
- `frontend/src/app/features/auth/login/login-page.component.scss`

O componente deve ser standalone (padrão Angular 17+). A guidelines seção 2.1 indica que standalone é o padrão; não é necessário `standalone: true` explícito em Angular 20, mas incluí-lo não causa problema para compatibilidade com Angular 18.

### Observação 2.1

O projeto atual usa Angular 18. A guideline menciona Angular v20+; o código deve ser compatível com a versão instalada. Verificar `package.json` para versão exata.

---

### Pensamento 2

As guidelines (seção 5.1) recomendam lazy loading para feature routes:

```typescript
export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login/login-page').then(m => m.LoginPageComponent)
  },
  ...
];
```

O `loadComponent` retorna uma Promise que resolve para o módulo/componente. Em Angular 18, a sintaxe é `then(m => m.LoginPageComponent)` assumindo que o componente seja exportado como `LoginPageComponent`. O caminho deve ser relativo ao arquivo de rotas (`app.routes.ts`).

### Ação 2.2

Atualizar `frontend/src/app/app.routes.ts`:

```typescript
import { Routes } from '@angular/router';

export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () =>
      import('./features/auth/login/login-page.component').then(m => m.LoginPageComponent),
  },
  {
    path: '',
    redirectTo: 'login',
    pathMatch: 'full',
  },
];
```

Redirecionar `''` para `login` garante que a aplicação mostre a tela de login ao acessar a raiz. Quando a home existir, ajustar para outra rota padrão.

### Observação 2.2

O redirecionamento para `/login` é adequado para um portal de autenticação cujo ponto de entrada é a tela de login. Conforme `docs/context.md`, o usuário é redirecionado para a página de login do centralizador quando não autenticado.

---

### Pensamento 3

O componente `LoginPageComponent` deve ter estrutura mínima inicial: template vazio ou com placeholder, para validar que a rota funciona. A implementação visual completa virá nas Fases 3 e 4.

### Ação 2.3

Estrutura inicial do componente:

```typescript
// login-page.component.ts
import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router'; // apenas se necessário

@Component({
  selector: 'app-login-page',
  standalone: true,
  imports: [],
  templateUrl: './login-page.component.html',
  styleUrl: './login-page.component.scss',
})
export class LoginPageComponent {}
```

Template inicial (para validar rota):
```html
<div class="login-page">
  <p>Página de Login (em construção)</p>
</div>
```

### Observação 2.3

Placeholder temporário. Será substituído pelo layout completo nas fases seguintes. O importante é que `ng serve` e navegação para `/login` funcionem.

---

### Pensamento 4

Verificar se `app.component.html` usa `<router-outlet>` para que as rotas filhas sejam renderizadas. O `app.component.ts` já importa `RouterOutlet` e o template deve ter `<router-outlet />` ou `<router-outlet></router-outlet>`.

### Ação 2.4

Validar `app.component.html` contém:
```html
<router-outlet />
```

Se não, adicionar. O `RouterOutlet` é o placeholder onde o Angular Router injeta o componente da rota ativa.

### Observação 2.4

Sem `router-outlet`, nenhuma rota filha será exibida. A feature 010 deve ter configurado isso; verificar antes de implementar.

---

### Decisão final Fase 2

**Entregáveis:**
1. Componente `LoginPageComponent` criado em `frontend/src/app/features/auth/login/`
2. Rota lazy-loaded `path: 'login'` em `app.routes.ts`
3. Redirecionamento `''` → `'login'`
4. Template placeholder para validar rota
5. Validação: `ng serve` e acessar `http://localhost:4200/login` exibe o componente

---

### Checklist de Implementação

1. [ ] Criar pasta `frontend/src/app/features/auth/login/` se não existir
2. [ ] Gerar/criar `LoginPageComponent` com arquivos .ts, .html, .scss
3. [ ] Atualizar `app.routes.ts` com rota `login` e lazy loading
4. [ ] Adicionar redirecionamento `''` → `'login'`
5. [ ] Verificar `app.component.html` tem `<router-outlet>`
6. [ ] Executar `ng serve` e validar navegação para `/login`
7. [ ] Garantir que o componente segue convenção `app-login-page` (selector)

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docs/code_guidelines/frontend.md §1.1 | Componente em `features/auth/login/` |
| docs/code_guidelines/frontend.md §5.1 | Lazy loading via `loadComponent` |
| docs/code_guidelines/frontend.md §6 | Nomenclatura `LoginPageComponent`, kebab-case |
| ADR-002 | Standalone component |

---

## Referências

- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [Angular Router - Lazy Loading](https://angular.dev/guide/routing/common-router-tasks#lazy-loading)
