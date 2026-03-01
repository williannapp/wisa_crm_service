# ADR-002 — Angular como Framework Frontend

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (frontend — portal de autenticação)

---

## Contexto

O `wisa-crm-service` expõe uma interface de usuário centralizada para autenticação. Essa interface é o único ponto de contato visual do usuário final com o sistema de auth — uma tela de login que:

- Coleta credenciais (email/senha)
- Exibe mensagens de erro de autenticação
- Pode exibir estados de validação (tenant inativo, assinatura vencida, etc.)
- Redireciona o usuário de volta ao sistema cliente após autenticação bem-sucedida
- Pode evoluir para suportar MFA (TOTP/SMS), SSO OAuth2 e auto-cadastro

O frontend opera em ambiente de alta responsabilidade de segurança: qualquer vulnerabilidade XSS, CSRF ou de injeção na tela de login compromete **toda a cadeia de autenticação** do sistema SaaS.

A escolha do framework frontend impacta diretamente:
- Superfície de ataque da aplicação
- Capacidade de escalar funcionalidades (MFA, SSO, página de gestão de assinatura)
- Consistência e manutenibilidade do código ao longo do tempo
- Integração com o backend Go e o fluxo de JWT

---

## Decisão

**Angular v20+ é o framework escolhido para o frontend do `wisa-crm-service`.**

A aplicação Angular será uma **SPA (Single Page Application)** servida pelo NGINX da VPS central, comunicando-se exclusivamente via HTTPS com o backend Go.

---

## Justificativa

### 1. Arquitetura opinativa reduz superfície de ataque

Angular é um framework **opinativo** — ele define como as coisas devem ser feitas, não apenas como podem ser feitas. Em um contexto de segurança, isso é uma vantagem crítica:

- **Template compiler** do Angular escapa automaticamente todo output no DOM, eliminando XSS por interpolação por design
- **DomSanitizer** força tratamento explícito de conteúdo dinâmico potencialmente inseguro
- **HttpClient** possui interceptors nativos para centralizar adição de headers (ex: tokens CSRF, Authorization)
- **Router Guards** fornecem proteção de rotas autenticadas de forma padronizada

Ao contrário de Vue (mais permissivo) ou React (biblioteca, não framework), Angular obriga o desenvolvedor a adotar padrões de segurança, reduzindo erros por omissão.

### 2. TypeScript estático como linguagem de primeira classe

Angular usa TypeScript obrigatoriamente. Em uma aplicação de autenticação:

- Tipagem dos modelos de claims JWT (interfaces TypeScript para o payload)
- Detecção em compile-time de erros ao manipular respostas do backend
- Contratos explícitos entre componentes via interfaces e types
- Tooling de refactoring seguro em IDEs

TypeScript elimina a categoria de erros de runtime causados por undefined/null em lógica crítica de validação de tokens.

### 3. Módulos e isolamento de contexto

A arquitetura de módulos do Angular (e os standalone components no Angular 17+) permite:

- Isolamento do módulo de autenticação (AuthModule) com seus próprios providers e interceptors
- Lazy loading de módulos secundários (gestão de assinatura, perfil, etc.) sem impacto no bundle inicial de login
- Injeção de dependência hierárquica que previne vazamento de estado entre contextos diferentes

### 4. Interceptors HTTP para gestão centralizada de tokens

O `HttpClient` do Angular permite interceptors que tratam centralizadamente:

```typescript
// Exemplo conceitual — não código de produção
@Injectable()
export class AuthInterceptor implements HttpInterceptor {
  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
    // Adiciona token, trata 401, redireciona para login
  }
}
```

Isso garante que **nenhuma requisição** escape do controle de autenticação por esquecimento do desenvolvedor.

### 5. Ecossistema e suporte de longo prazo

Angular é mantido pelo Google com:
- Ciclo de releases semestral com suporte LTS de 18 meses por versão major
- Roadmap público e breaking changes com migração assistida por `ng update`
- Angular v20 introduz melhorias significativas em Signals (reatividade granular sem overhead do Zone.js), reduzindo consumo de memória e melhorando performance percebida

A estabilidade do framework é crítica para um sistema de autenticação que deve ser mantido por anos.

### 6. Standalone Components e Tree-shaking no Angular 17+

A migração para Standalone Components (padrão no Angular 17+) permite:
- Bundle final menor via tree-shaking mais agressivo
- Remoção progressiva do Zone.js com Signals (Angular 18+)
- Componentes mais testáveis e desacoplados

Para uma tela de login com bundle inicial pequeno e crítico, isso representa melhoria real de performance de carregamento.

---

## Consequências

### Positivas
- XSS por interpolação eliminado por design do template compiler
- Padronização de acesso HTTP via interceptors centraliza controle de tokens
- TypeScript estático reduz bugs na manipulação de respostas do backend
- Estrutura clara para evolução (MFA, OAuth2, gestão de assinatura)
- Router Guards protegem rotas autenticadas de forma declarativa

### Negativas
- Bundle inicial maior que React/Vue para interfaces simples (mitigado por lazy loading e tree-shaking)
- Curva de aprendizado mais íngreme para novos desenvolvedores (DI, decorators, módulos)
- Verbosidade maior para componentes simples comparado ao Vue ou React
- Build mais lento em projetos grandes (mitigado pelo esbuild/Vite no Angular 17+)

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| XSS via uso incorreto de `bypassSecurityTrust*` | Média | Crítico | Alta |
| CSRF em requisições POST de login | Média | Alto | Alta |
| Exposição de token JWT no LocalStorage | Alta | Alto | Alta |
| Clickjacking na tela de login | Média | Alto | Média |
| Dependências npm vulneráveis | Média | Médio | Média |
| Leakage de informações em mensagens de erro verbosas | Alta | Médio | Média |

---

## Mitigações

### XSS
- **Nunca** usar `bypassSecurityTrustHtml`, `bypassSecurityTrustScript` ou equivalentes sem revisão de segurança formal
- Implementar **Content Security Policy (CSP)** via header NGINX:
  ```
  Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' https://api.wisa-crm.com; frame-ancestors 'none';
  ```
- Nunca renderizar HTML dinâmico recebido do backend

### Armazenamento de tokens
- **Não armazenar JWT no LocalStorage** — vulnerável a XSS
- Preferir **cookies HTTP-only + Secure + SameSite=Strict** para armazenar o JWT de sessão
- Se usar memory storage (variável em serviço Angular), considerar o comportamento em múltiplas abas

### CSRF
- Configurar cookies com `SameSite=Strict` (proteção primária)
- Implementar Double Submit Cookie pattern como proteção secundária
- O backend Go deve validar o header `Origin`/`Referer` nas requisições de autenticação

### Clickjacking
- Configurar `X-Frame-Options: DENY` no NGINX
- Equivalente moderno via CSP: `frame-ancestors 'none'`

### Mensagens de erro
- **Nunca** diferenciar no frontend/backend entre "usuário não encontrado" e "senha incorreta" — usar mensagem genérica: "Credenciais inválidas"
- Evitar exposição de stack traces ou detalhes técnicos ao usuário

### Supply chain
- Executar `npm audit` na CI e tratar vulnerabilidades CRÍTICAS como bloqueantes de deploy
- Usar `package-lock.json` versionado
- Considerar `npm ci` (install determinístico) ao invés de `npm install` na CI

---

## Configuração de Segurança Recomendada

### Headers HTTP obrigatórios (via NGINX)

```nginx
add_header X-Frame-Options "DENY" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Content-Security-Policy "default-src 'self'; ..." always;
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
```

### Build de produção
- Sempre gerar build com `ng build --configuration production` (minification, tree-shaking, Ahead-of-Time compilation)
- AOT (Ahead-of-Time) compilation elimina o template compiler do bundle de produção, reduzindo superfície de ataque

---

## Alternativas Consideradas

### React (com TypeScript)
- **Prós:** Ecossistema maior, flexibilidade, performance de renderização (VDOM + Fiber)
- **Contras:** É uma biblioteca, não um framework — requer curadoria de bibliotecas de roteamento, estado, HTTP, forms (aumenta superfície de ataque e risco de escolhas inadequadas); sem opinião sobre segurança por padrão

### Vue.js 3
- **Prós:** Curva de aprendizado menor, performance excelente, Composition API elegante
- **Contras:** Menos opinativo que Angular em relação a segurança, ecossistema corporativo menor, menor suporte a TypeScript nativo (melhorou no Vue 3 mas ainda inferior ao Angular)

### Server-Side Rendering (Go templates ou HTMX)
- **Prós:** Sem JavaScript complexo, menor superfície de ataque, melhor SEO (irrelevante para tela de login)
- **Contras:** Experiência de usuário inferior para fluxos dinâmicos (MFA, validação em tempo real), menor flexibilidade para evolução do produto (OAuth2, perfil de usuário, etc.)

**Angular v20+ oferece o melhor equilíbrio entre segurança por design, escalabilidade de funcionalidades e produtividade para um portal de autenticação SaaS de longo prazo.**

---

## Referências

- [Angular Security Guide](https://angular.dev/best-practices/security)
- [OWASP Angular Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Angular_Based_Application_Security_Cheat_Sheet.html)
- [Angular Standalone Components](https://angular.dev/guide/components/importing)
- [Angular Signals](https://angular.dev/guide/signals)
