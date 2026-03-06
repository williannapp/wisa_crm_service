# Frontend — wisa-crm-service

Projeto Angular 18 do portal de autenticação centralizado. Estrutura inicial com Clean Architecture (domain, application, infrastructure, features, core, shared).

**Requisitos:** Node.js >= 18.19.1 (recomendado: 20.x LTS)

## Estrutura de Diretórios

```
src/app/
├── core/           # Singletons: auth, http
├── domain/         # Entidades e ports (interfaces)
├── application/    # Use cases
├── infrastructure/ # Implementações (http, storage)
├── features/       # Módulos por funcionalidade (auth, home)
└── shared/         # Componentes e pipes reutilizáveis
```

## Development server

Run `ng serve` for a dev server. Navigate to `http://localhost:4200/`. The application will automatically reload if you change any of the source files.

**Proxy:** Requisições para `/api` são encaminhadas ao backend em `http://localhost:8080`. Suba o backend antes de testar o login.

**Login:** A rota `/login` exige query params `tenant_slug`, `product_slug` e `state`. Ex.: `http://localhost:4200/login?tenant_slug=cliente1&product_slug=gestao-pocket&state=abc123`

## Code scaffolding

Run `ng generate component component-name` to generate a new component. You can also use `ng generate directive|pipe|service|class|guard|interface|enum|module`.

## Build

Run `ng build` to build the project. The build artifacts will be stored in the `dist/` directory.

Produção: `ng build --configuration production`

## Docker

### Build da imagem

```bash
docker compose build frontend
```

### Executar stack completo

```bash
docker compose up
```

Frontend em http://localhost:4200 (NGINX serve os arquivos estáticos).

## Running unit tests

Run `ng test` to execute the unit tests via [Karma](https://karma-runner.github.io).

## Running end-to-end tests

Run `ng e2e` to execute the end-to-end tests via a platform of your choice. To use this command, you need to first add a package that implements end-to-end testing capabilities.

## Further help

To get more help on the Angular CLI use `ng help` or go check out the [Angular CLI Overview and Command Reference](https://angular.dev/tools/cli) page.
