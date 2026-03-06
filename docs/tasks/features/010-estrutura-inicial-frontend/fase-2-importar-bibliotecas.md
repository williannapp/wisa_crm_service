# Fase 2 — Importar Bibliotecas Necessárias

## Objetivo

Inicializar o projeto Angular v20+ dentro de `frontend/` e garantir que todas as dependências necessárias para a estrutura inicial estejam presentes no `package.json`, em conformidade com ADR-002 e supply chain seguro.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-002 especifica Angular v20+ como framework frontend. O Angular CLI, ao executar `ng new`, já instala as dependências base: `@angular/core`, `@angular/common`, `@angular/router`, `rxjs`, `zone.js`, `typescript`, etc. Para a estrutura inicial (sem login, sem chamadas HTTP complexas), as dependências são essencialmente as que o `ng new` traz.

### Ação 2.1

Identificar dependências para a estrutura inicial:

| Pacote | Finalidade | Inclusão |
|--------|------------|----------|
| @angular/core v20+ | Framework base | Via `ng new` |
| @angular/common | Diretivas, pipes, HttpClient | Via `ng new` |
| @angular/router | Roteamento, lazy loading | Via `ng new` (--routing) |
| @angular/platform-browser | Bootstrap da aplicação | Via `ng new` |
| rxjs | Programação reativa | Via `ng new` |
| zone.js | Change detection (legacy) | Via `ng new` |
| typescript 5.8+ | Linguagem | Via `ng new` |

### Observação 2.1

Angular 20 utiliza esbuild/Vite para builds mais rápidos. O `ng new` configura o `angular.json` apropriadamente. Nenhuma biblioteca adicional é estritamente necessária para a estrutura inicial. Bibliotecas como `@angular/forms`, `@angular/animations` podem ser adicionadas quando as features forem implementadas.

---

### Pensamento 2

Se a Fase 1 tiver criado apenas diretórios (sem `ng new`), a Fase 2 deve executar `ng new` para gerar o projeto. Se a Fase 1 já tiver executado `ng new`, a Fase 2 verifica/ajusta as dependências.

### Ação 2.2

Cenário A — Fase 1 criou apenas pastas:
```bash
# Na raiz do projeto
ng new frontend --directory=frontend --routing --style=scss
```
Isso cria `frontend/` com todo o conteúdo. Os diretórios criados na Fase 1 (se estiverem em `frontend/`) seriam sobrescritos. Portanto, a ordem correta é: **Fase 2 executa `ng new` primeiro**, gerando `frontend/` completo. Os diretórios adicionais (core, domain, etc.) são criados na Fase 1 **ou** na Fase 2 após o `ng new`.

Cenário B — Fase 1 já executou `ng new` e criou subdiretórios:
Fase 2 verifica `package.json` e adiciona dependências extras se necessário.

### Observação 2.2

Para evitar duplicação de esforço, recomenda-se que a **Fase 1** execute `ng new` e crie os subdiretórios, e a **Fase 2** se concentre em: (1) validar versões das dependências; (2) adicionar dependências opcionais para a estrutura (ex.: `@angular/forms` se forms reativos forem usados na tela de login futura); (3) executar `npm install` e validar o build. Na estrutura inicial pura, a Fase 2 pode ser resumida a **validar que o projeto compila**.

---

### Pensamento 3

O ADR-002 e as guidelines de segurança (supply chain) recomendam:
- `package-lock.json` versionado para builds determinísticos
- `npm audit` na CI para vulnerabilidades
- Mínimo de dependências para reduzir superfície de ataque

### Ação 2.3

Garantir no `package.json`:
- Versão do Angular alinhada a v20+
- Node.js engine: `"engines": { "node": ">=20.19.0" }` (conforme documentação Angular 20)
- Scripts: `build`, `serve`, `test`

### Observação 2.3

O `ng new` já define scripts padrão. Verificar que `ng build --configuration production` está disponível para uso no Dockerfile da Fase 4.

---

### Pensamento 4

Dependências de desenvolvimento (devDependencies) trazidas pelo `ng new`: `@angular-devkit/build-angular`, `karma`, `jasmine-core`, etc. Estas são adequadas. Nenhuma adição obrigatória para a estrutura inicial.

### Ação 2.4

Não adicionar bibliotecas de terceiros (UI kits, charts, etc.) nesta fase. Manter o projeto enxuto. Dependências adicionais (ex.: validação de formulários, interceptors HTTP) serão introduzidas nas features subsequentes.

### Observação 2.4

Consistente com o princípio de supply chain mínimo e com a feature 001 do backend (mínimo de dependências).

---

### Checklist de Implementação

1. [ ] Se Fase 1 não executou `ng new`: executar `ng new frontend --directory=frontend --routing --style=scss` na raiz
2. [ ] Verificar que `package.json` contém Angular v20+ (ou versão adequada)
3. [ ] Executar `npm install` (ou `npm ci` se `package-lock.json` existir)
4. [ ] Validar que `npm run build` executa sem erros
5. [ ] Validar que `ng serve` inicia o servidor de desenvolvimento
6. [ ] Documentar a versão do Node.js requerida (>=20.19.0) no README ou em docs
7. [ ] Garantir `package-lock.json` commitado para reprodutibilidade

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-002 Angular v20+ | Projeto criado com Angular 20+ |
| Supply chain | package-lock.json versionado; mínimo de deps |
| Segurança | npm audit sem vulnerabilidades CRÍTICAS (tratar na CI) |

---

## Referências

- [ADR-002 — Angular como framework frontend](../../../adrs/ADR-002-angular-como-framework-frontend.md)
- [Angular Installation](https://angular.dev/installation)
- [Angular CLI new](https://angular.dev/cli/new)
