# Fase 4 — Card e Formulário (Estrutura Visual)

## Objetivo

Implementar o card com formulário de login: estrutura visual completa (campos usuário e senha, botão Entrar, links Esqueceu senha / Fale conosco), replicando o design do protótipo Login-Wisa. **Sem lógica de autenticação** — apenas UI/UX visual.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O protótipo possui um Card com:
- Barra gradient no topo: `h-2 bg-gradient-to-r from-blue-600 to-cyan-400`
- CardHeader: título "Bem-vindo de volta", descrição "Acesse sua conta Wisa Labs para continuar"
- CardContent: formulário com Label + Input para usuário e senha
- Campo usuário: ícone UserIcon à esquerda, placeholder "Seu nome de usuário"
- Campo senha: ícone LockIcon à esquerda, botão toggle Eye/EyeOff à direita, placeholder "••••••••"
- Link "Esqueceu a senha?" ao lado do label da senha
- Botão "Entrar" full-width, azul
- CardFooter: "Ainda não tem uma conta? Fale conosco"

O backend usa **email** como credencial (conforme context.md: "usuário informa seu usuário e senha" — na prática, o endpoint de login valida email). O protótipo usa "Usuário" e placeholder "Seu nome de usuário". Manter a terminologia do protótipo para design; a integração futura pode ajustar label para "E-mail" se necessário. O foco aqui é design, não contrato de API.

### Ação 4.1

Estruturar o HTML do card:

```html
<div class="login-page__card">
  <div class="login-page__card-bar"></div>
  <header class="login-page__card-header">
    <h1 class="login-page__card-title">Bem-vindo de volta</h1>
    <p class="login-page__card-description">Acesse sua conta Wisa Labs para continuar</p>
  </header>
  <div class="login-page__card-content">
    <form class="login-page__form" (ngSubmit)="onSubmit()">
      <!-- Campos e botão -->
    </form>
  </div>
  <footer class="login-page__card-footer">
    <p>Ainda não tem uma conta? <a href="#">Fale conosco</a></p>
  </footer>
</div>
```

O `(ngSubmit)` pode chamar um método vazio ou `console.log` por ora — sem chamada HTTP. O `href="#"` nos links é placeholder; em features futuras, substituir por routerLink ou href real.

### Observação 4.1

O formulário precisa de `FormsModule` ou `ReactiveFormsModule` se usarmos `ngModel` ou `formControlName`. Para inputs controlados (two-way binding), usar `[(ngModel)]` requer `FormsModule`. Para apenas design estático, inputs podem ser não controlados; porém, para toggle de senha e possível validação visual, usar `ngModel` ou `formControl` é útil. Incluir `FormsModule` nos imports do componente.

---

### Pensamento 2

O componente precisa de estado para:
- `username` e `password` (valores dos inputs)
- `showPassword` (boolean para toggle tipo text/password)

As guidelines (seção 4) recomendam **Signals** para estado. Usar `signal()` para esses valores:

```typescript
username = signal('');
password = signal('');
showPassword = signal(false);
```

Para two-way binding com inputs, temos duas opções: (A) usar `[value]="username()"` e `(input)="username.set($event.target.value)"`; (B) usar `ngModel` com model signals (Angular 17+). O Angular 18 suporta `ngModel` tradicional. Para simplicidade e compatibilidade, usar `ngModel` com FormsModule é mais direto para formulários. As guidelines preferem Signals para estado local — então usar `[value]` + `(input)` com signals mantém conformidade.

### Ação 4.2

Implementar bindings com Signals:

```html
<input
  type="text"
  [value]="username()"
  (input)="username.set($any($event.target).value)"
  placeholder="Seu nome de usuário"
/>
<input
  [type]="showPassword() ? 'text' : 'password'"
  [value]="password()"
  (input)="password.set($any($event.target).value)"
  placeholder="••••••••"
/>
<button type="button" (click)="showPassword.update(v => !v)">...</button>
```

O `onSubmit()` pode ser vazio ou logar no console — sem side effects de rede.

### Observação 4.2

Conformidade com guidelines §4 (Signals). O `$any()` evita erro de tipo no `$event.target`. Alternativa: `(input)="username.set(usr.value)"` com `#usr` template ref.

---

### Pensamento 3

Estilos do card conforme protótipo:
- Card: `bg-white/95` → `background: rgba(255,255,255,0.95)`, `backdrop-filter: blur()`, `shadow-2xl`, `border-white/20`, `max-w-md`, `w-full`
- Barra topo: `h-2` (8px), gradient `from-blue-600 to-cyan-400`
- Título: `text-2xl font-bold text-slate-900`
- Descrição: `text-slate-500`
- Input: `pl-10` (padding left para ícone), `bg-slate-50`, `border-slate-200`, focus ring `blue-500`
- Botão: `bg-blue-600 hover:bg-blue-700`, `w-full`, `py-2.5`, `rounded-lg`, `shadow-md`

Usar variáveis do `_login-tokens.scss` em todos os valores.

### Ação 4.3

Estilos SCSS para o card:

```scss
.login-page__card {
  width: 100%;
  max-width: 28rem; // 448px - max-w-md
  background: var(--login-card-bg);
  backdrop-filter: blur(12px);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 0.5rem;
  position: relative;
  z-index: 10;
  overflow: hidden;
}

.login-page__card-bar {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 8px;
  background: linear-gradient(to right, var(--login-card-bar-start), var(--login-card-bar-end));
}

.login-page__card-header {
  padding-top: 2.5rem;
  padding-left: 1.5rem;
  padding-right: 1.5rem;
  text-align: center;
}

.login-page__card-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--login-title);
}

.login-page__card-description {
  color: var(--login-description);
  margin-top: 0.25rem;
}

// ... inputs, botão, footer usando variáveis
```

### Observação 4.3

O `backdrop-filter` tem suporte amplo em navegadores modernos. Fallback: sem blur, o card ainda será legível. O `overflow: hidden` evita que a barra gradient vaze nos cantos arredondados.

---

### Pensamento 4

O protótipo usa ícones Lucide (UserIcon, LockIcon, EyeIcon, EyeOffIcon). O projeto Angular não tem Lucide instalado. Opções: (A) instalar `lucide-angular`; (B) usar Angular Material Icons; (C) usar SVGs inline. Para fidelidade ao protótipo, Lucide é a opção mais próxima. O `lucide-angular` é o port oficial para Angular.

### Ação 4.4

Incluir no planejamento a instalação de `lucide-angular`:
```bash
npm install lucide-angular
```

Uso no componente:
```typescript
import { User, Lock, Eye, EyeOff } from 'lucide-angular';
// No template:
<lucide-icon name="user" [size]="18"></lucide-icon>
```

A API do lucide-angular pode variar — verificar documentação. Alternativa: usar SVG inline com os paths dos ícones para evitar dependência. Para manter fidelidade e produtividade, recomendar `lucide-angular`. Se o projeto preferir mínimo de deps, SVGs inline são aceitáveis.

### Observação 4.4

O `lucide-angular` adiciona ~X kB ao bundle. Os ícones são tree-shakable. Para 4 ícones, o impacto é pequeno. Decisão: incluir `lucide-angular` na Fase 4 ou adiar para Fase 5. Fase 4 foca em estrutura; Fase 5 em "componentes visuais" — ícones podem ser Fase 5. Para entregar um card visualmente completo na Fase 4, os ícones fazem parte. Incluir ícones na Fase 4.

---

### Pensamento 5

Segurança (ADR-002, guidelines §8.2): Nunca diferenciar "usuário não encontrado" vs "senha incorreta". Nesta fase, não há chamada de API — apenas UI. Os placeholders e labels são neutros. O campo está como "Usuário" no protótipo; o backend usa email. Manter "Usuário" no design por ora; a integração futura pode alterar.

O link "Esqueceu a senha?" com `href="#"` é placeholder. O link "Fale conosco" idem. Não implementar navegação real nesta feature.

### Ação 4.5

Manter links como `href="#"` ou `routerLink="#"` com `(click)="$event.preventDefault()"` para evitar scroll. Ou usar `href="javascript:void(0)"` — menos recomendado. O `#` com `preventDefault` no click é mais limpo se adicionarmos handler futuro. Por ora, `href="#"` é suficiente.

### Observação 4.5

Nenhuma vulnerabilidade introduzida. Sem dados sensíveis expostos. O formulário não submete para lugar nenhum — apenas estrutura visual.

---

### Decisão final Fase 4

**Entregáveis:**
1. Card com barra gradient no topo
2. Header com título e descrição
3. Formulário com campos usuário e senha (com ícones)
4. Toggle mostrar/ocultar senha
5. Link "Esqueceu a senha?"
6. Botão "Entrar"
7. Footer com "Ainda não tem uma conta? Fale conosco"
8. Estado com Signals (username, password, showPassword)
9. Estilos usando variáveis CSS dos tokens
10. `onSubmit` vazio ou console.log — sem HTTP
11. Validação visual: formulário idêntico ao protótipo

---

### Checklist de Implementação

1. [ ] Instalar `lucide-angular` (ou usar SVGs inline)
2. [ ] Implementar estrutura HTML do card (bar, header, content, footer)
3. [ ] Adicionar formulário com inputs para usuário e senha
4. [ ] Implementar Signals: username, password, showPassword
5. [ ] Adicionar ícones User, Lock, Eye, EyeOff
6. [ ] Implementar toggle de senha
7. [ ] Adicionar link "Esqueceu a senha?" e "Fale conosco"
8. [ ] Implementar botão "Entrar" com `onSubmit` placeholder
9. [ ] Aplicar estilos SCSS usando variáveis `--login-*`
10. [ ] Importar `FormsModule` se usar ngModel (ou manter binding manual)
11. [ ] Validar visualmente contra protótipo
12. [ ] Garantir que `ng build` executa sem erros

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Protótipo Login-Wisa | Estrutura idêntica do card e formulário |
| Guidelines §4 | Signals para estado (username, password, showPassword) |
| Guidelines §6 | Nomenclatura LoginPageComponent, kebab-case |
| ADR-002 | Sem lógica de auth; mensagens genéricas quando houver |
| Escopo | Design apenas — sem chamada HTTP |

---

## Referências

- [Protótipo login.tsx](/Login-Wisa/client/src/pages/login.tsx)
- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [fase-1-design-tokens-estilos.md](./fase-1-design-tokens-estilos.md)
- [Lucide Angular](https://github.com/lucide-icons/lucide/tree/main/packages/angular)
