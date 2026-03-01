# Code Guidelines — Frontend Angular v20

## 1. Estrutura de Pastas Recomendada

### 1.1 Organização por Features

Organize o projeto por áreas de funcionalidade (feature-based) em vez de por tipo de código:

```
src/
├── app/
│   ├── core/                    # Singleton services, guards, interceptors
│   │   ├── auth/
│   │   │   ├── auth.service.ts
│   │   │   ├── auth.guard.ts
│   │   │   └── auth.interceptor.ts
│   │   └── http/
│   │       └── api-config.ts
│   ├── features/
│   │   ├── auth/
│   │   │   ├── login/
│   │   │   │   ├── login-page.component.ts
│   │   │   │   ├── login-page.component.html
│   │   │   │   └── login-page.component.spec.ts
│   │   │   └── ...
│   │   └── home/
│   │       └── home-page/
│   │           ├── home-page.component.ts
│   │           └── ...
│   ├── shared/                  # Componentes e pipes reutilizáveis
│   │   ├── components/
│   │   └── pipes/
│   ├── app.component.ts
│   └── app.config.ts
├── main.ts
└── index.html
```

### 1.2 Separação de Camadas (Clean Architecture aplicada ao Frontend)

```
src/app/
├── domain/                      # Entidades e interfaces de negócio
│   ├── models/
│   │   ├── user.model.ts
│   │   └── tenant.model.ts
│   └── ports/
│       └── auth.repository.port.ts
├── application/                 # Casos de uso e serviços de aplicação
│   └── use-cases/
│       └── auth/
│           ├── authenticate.use-case.ts
│           └── authenticate.use-case.spec.ts
├── infrastructure/              # Implementações de repositórios e adapters
│   ├── http/
│   │   └── auth.repository.impl.ts
│   └── storage/
│       └── token.storage.impl.ts
└── presentation/                # Componentes, páginas, rotas
    └── features/
```

---

## 2. Standalone Components

### 2.1 Padrão Default

- **Angular v20:** Standalone components são o padrão. **NÃO** defina `standalone: true` explicitamente nos decorators — é comportamento implícito.
- Prefira standalone components a NgModules para todos os novos componentes.

### 2.2 Importação de Componentes

```typescript
import { ProfilePhoto } from './profile-photo';

@Component({
  imports: [ProfilePhoto],
  selector: 'app-user-profile',
  template: `<profile-photo [src]="userAvatar" />`
})
export class UserProfile { }
```

### 2.3 Bootstrap da Aplicação

```typescript
import { bootstrapApplication } from '@angular/platform-browser';
import { AppComponent } from './app/app.component';

bootstrapApplication(AppComponent, {
  providers: [
    { provide: API_BASE_URL, useValue: 'https://api.example.com' },
    provideHttpClient(withInterceptors([authInterceptor])),
  ]
});
```

---

## 3. Padrão de Services

### 3.1 Responsabilidade Única

- Design services com uma responsabilidade bem definida para promover modularidade e testabilidade.
- Services destinados a serem singletons: use `providedIn: 'root'`.

### 3.2 Configuração Global com InjectionToken

```typescript
export interface AppConfig {
  apiUrl: string;
  version: string;
  features: Record<string, boolean>;
}

export const APP_CONFIG = new InjectionToken<AppConfig>('app.config', {
  providedIn: 'root',
  factory: () => ({
    apiUrl: 'https://api.example.com',
    version: '1.0.0',
    features: { darkMode: true, analytics: false }
  })
});
```

### 3.3 Uso do inject()

- Prefira a função `inject()` para injeção de dependência em vez de constructor injection.
- Maior flexibilidade em standalone components e funções.

```typescript
@Component({
  selector: 'app-header',
  template: `<h1>Version: {{ config.version }}</h1>`
})
export class HeaderComponent {
  config = inject(APP_CONFIG);
}
```

---

## 4. Gerenciamento de Estado com Signals

### 4.1 Signals como Padrão

- Use **Signals** para gerenciamento de estado em componentes — reatividade granular sem overhead do Zone.js.
- Prefira `signal()`, `computed()` e `effect()` em vez de RxJS para estado local quando aplicável.

### 4.2 Estado Local com Signals

```typescript
import { Component, signal, computed } from '@angular/core';

@Component({
  selector: 'app-counter',
  template: `
    <p>Count: {{ count() }}</p>
    <p>Double Count: {{ doubleCount() }}</p>
    <button (click)="increment()">Increment</button>
  `
})
export class CounterComponent {
  count = signal(0);
  doubleCount = computed(() => this.count() * 2);

  increment() {
    this.count.update(value => value + 1);
  }
}
```

### 4.3 Computed Signals

- `computed()` deriva estado de forma reativa e read-only.
- Angular recalcula apenas quando dependências mudam.
- Use para cálculos, transformações e estado derivado complexo.

```typescript
const firstName = signal('John');
const lastName = signal('Doe');
const fullName = computed(() => `${firstName()} ${lastName()}`);
```

---

## 5. Boas Práticas de Performance

### 5.1 Lazy Loading de Rotas

- Implemente lazy loading para feature routes para reduzir bundle inicial.
- Use `loadComponent` para carregamento sob demanda.

```typescript
export const routes: Routes = [
  {
    path: 'login',
    loadComponent: () => import('./features/auth/login/login-page').then(m => m.LoginPage)
  },
  {
    path: '',
    loadComponent: () => import('./features/home/home-page').then(m => m.HomePage)
  }
];
```

### 5.2 NgOptimizedImage

- Use `NgOptimizedImage` para todas as imagens estáticas.
- **Não** use para imagens inline base64.

### 5.3 Tree-shakable Providers

- Use lightweight injection tokens e providedIn: 'root' para permitir tree-shaking de serviços não utilizados.

### 5.4 Host Bindings

- **NÃO** use decorators `@HostBinding` e `@HostListener`.
- Coloque host bindings no objeto `host` do decorator `@Component` ou `@Directive`.

---

## 6. Convenções de Nomenclatura

| Elemento | Convenção | Exemplo |
|----------|-----------|---------|
| Componentes | PascalCase + sufixo descritivo | `LoginPageComponent`, `UserProfileComponent` |
| Services | PascalCase + sufixo Service | `AuthService`, `TenantService` |
| Arquivos de componente | kebab-case | `login-page.component.ts` |
| Arquivos de service | kebab-case | `auth.service.ts` |
| Selectors | prefixo `app-` | `app-login-form`, `app-user-card` |
| Interfaces/Models | PascalCase | `User`, `Tenant`, `LoginRequest` |

---

## 7. Estrutura Ideal para Testes

### 7.1 TestBed com provideHttpClient

```typescript
TestBed.configureTestingModule({
  providers: [
    AuthService,
    provideHttpClient(withInterceptors([authInterceptor])),
    provideHttpClientTesting(),
  ],
});
```

### 7.2 Teste de Interceptors

```typescript
const service = TestBed.inject(AuthService);
const req = httpTesting.expectOne('/api/config');
expect(req.request.headers.get('X-Authentication-Token')).toEqual(service.getAuthToken());
```

### 7.3 Teste de Erros de Rede

```typescript
const req = httpTesting.expectOne('/api/config');
req.error(new ProgressEvent('network error!'));
// Assert que a aplicação tratou o erro corretamente
```

---

## 8. Tratamento de Erros

### 8.1 Centralização via Interceptors

- Use interceptors para tratar erros HTTP de forma centralizada (ex.: 401, 403, 500).
- Adicione token de autenticação em todas as requisições via interceptor.

### 8.2 Mensagens de Erro de Autenticação

- **Nunca** diferenciar entre "usuário não encontrado" e "senha incorreta".
- Use mensagem genérica: "Credenciais inválidas".
- Evite expor stack traces ou detalhes técnicos ao usuário.

---

## 9. Segurança

### 9.1 XSS

- O template compiler escapa automaticamente o output no DOM.
- **Nunca** usar `bypassSecurityTrustHtml`, `bypassSecurityTrustScript` sem revisão de segurança.
- Nunca renderizar HTML dinâmico recebido do backend.

### 9.2 Armazenamento de Tokens

- **Não** armazenar JWT no LocalStorage (vulnerável a XSS).
- Preferir cookies HTTP-only + Secure + SameSite=Strict.

### 9.3 Build de Produção

- Sempre `ng build --configuration production` (minification, tree-shaking, AOT).

---

## 10. Diretrizes de Escalabilidade

| Diretriz | Descrição |
|----------|-----------|
| Feature isolation | Cada feature deve ser autocontida com seus componentes, services e rotas |
| Lazy loading | Features secundárias carregadas sob demanda |
| Shared vs. Core | `shared/` para componentes reutilizáveis; `core/` para singletons de aplicação |
| Domain-first | Modelos e interfaces no domain; apresentação apenas consome |
| Testes por camada | Testar services e use cases isoladamente; mocks para repositórios |
