# Fase 1 — Estrutura de Diretórios do Frontend

## Objetivo

Criar a estrutura de diretórios do frontend dentro da pasta `frontend/`, em conformidade com `docs/code_guidelines/frontend.md`, ADR-002 (Angular como framework) e a abordagem feature-based com Clean Architecture aplicada ao frontend.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O documento `docs/code_guidelines/frontend.md` define duas visões complementares: (1) organização por features e (2) separação de camadas (domain, application, infrastructure, presentation). O ADR-002 especifica Angular v20+ com standalone components como padrão. A estrutura deve refletir esses princípios desde o início.

### Ação 1.1

Consultar a estrutura exata definida nas guidelines:

**Organização por Features:**
- `src/app/core/` — Singleton services, guards, interceptors
- `src/app/features/` — Módulos por funcionalidade (auth, home, etc.)
- `src/app/shared/` — Componentes e pipes reutilizáveis

**Clean Architecture aplicada ao Frontend:**
- `src/app/domain/` — Entidades e interfaces de negócio
- `src/app/application/` — Casos de uso e serviços de aplicação
- `src/app/infrastructure/` — Implementações de repositórios e adapters
- `src/app/presentation/` — Componentes, páginas, rotas (ou integrado a features)

### Observação 1.1

As guidelines sugerem que `presentation/` pode conter `features/`. Na prática, pode-se unificar: `features/` ficam em `src/app/features/` e as camadas domain/application/infrastructure coexistem. O importante é manter a regra de dependência: domain não importa infrastructure; application orquestra.

---

### Pensamento 2

O Angular CLI (`ng new`) gera uma estrutura padrão que pode não coincidir exatamente com as guidelines. Duas abordagens:

A) Criar manualmente os diretórios conforme especificação e depois integrar com o output do `ng new`.
B) Usar `ng new` com opções mínimas e, em seguida, reorganizar/criar os diretórios adicionais (domain, application, infrastructure) que o CLI não gera por padrão.

### Ação 1.2

A abordagem recomendada é **B**: executar `ng new` dentro de `frontend/` (ou na raiz temporariamente e mover) para obter a base do projeto (package.json, angular.json, tsconfig, etc.) e, em seguida, criar os diretórios adicionais conforme guidelines. O `ng new` pode ser executado como `ng new frontend --directory=frontend` na raiz do repositório, gerando `frontend/` com a estrutura base.

### Observação 1.2

O comando `ng new frontend --directory=frontend --routing --style=scss` cria a pasta `frontend/` com estrutura Angular padrão. A flag `--no-standalone` não deve ser usada — Angular 20 tem standalone como padrão. Após a criação, adicionar os diretórios `domain/`, `application/`, e ajustar a organização de `features/` e `core/`.

---

### Pensamento 3

Na fase inicial (estrutura somente), o requisito explicitamente proíbe implementação de código. Portanto:
- Criar a estrutura de pastas
- Os arquivos gerados pelo `ng new` são mínimos (app.component, main, app.config, index.html) — isso é "scaffolding", não "implementação de features"
- A decisão: criar diretórios vazios com `.gitkeep` OU usar `ng new` que gera arquivos base. O `ng new` gera código de bootstrap — isso é necessário para "estrutura inicial". O requisito diz "não implemente nenhum código na estrutura"; interpreta-se que não se deve implementar lógica de negócio (login, auth, etc.), mas o scaffold mínimo do Angular é parte da "estrutura".

### Ação 1.3

Definir escopo da Fase 1:
1. Executar `ng new frontend --directory=frontend` (ou equivalente) na raiz do projeto
2. Após scaffold, criar os diretórios adicionais conforme guidelines: `domain/`, `application/`, ajustar `core/`, `features/`, `shared/`
3. Não adicionar componentes de features (login, home) — apenas a estrutura de pastas

### Observação 1.3

Se o planejamento for seguido à risca "não implemente nenhum código", a Fase 1 pode ser restrita a: **apenas criar a árvore de diretórios** (manual ou via script), sem executar `ng new`. O `ng new` seria parte da Fase 2 (importar bibliotecas). Assim, na Fase 1 criam-se somente as pastas com `.gitkeep`. **Decisão:** Fase 1 = criar diretórios; Fase 2 = `ng new` + dependências. Isso mantém consistência com a feature 001 do backend (Fase 1 apenas pastas).

---

### Pensamento 4

Estrutura final desejada para `frontend/`:

```
frontend/
├── src/
│   ├── app/
│   │   ├── core/                 # Auth, HTTP config, singletons
│   │   │   ├── auth/
│   │   │   └── http/
│   │   ├── domain/               # Entidades e ports
│   │   │   ├── models/
│   │   │   └── ports/
│   │   ├── application/          # Use cases
│   │   │   └── use-cases/
│   │   ├── infrastructure/       # Implementações
│   │   │   ├── http/
│   │   │   └── storage/
│   │   ├── features/             # Módulos por funcionalidade
│   │   │   ├── auth/
│   │   │   └── home/
│   │   ├── shared/               # Componentes e pipes reutilizáveis
│   │   │   ├── components/
│   │   │   └── pipes/
│   │   ├── app.component.ts
│   │   └── app.config.ts
│   ├── main.ts
│   └── index.html
├── angular.json
├── package.json
├── tsconfig.json
└── ...
```

O `ng new` já cria `src/`, `app/`, `main.ts`, `index.html`, `angular.json`, etc. Os diretórios `core/`, `domain/`, `application/`, `infrastructure/`, `features/`, `shared/` devem ser criados manualmente dentro de `src/app/`.

### Ação 1.4

Checklist de diretórios a criar (dentro de `frontend/src/app/`):

- `core/auth/`
- `core/http/`
- `domain/models/`
- `domain/ports/`
- `application/use-cases/`
- `infrastructure/http/`
- `infrastructure/storage/`
- `features/auth/`
- `features/home/`
- `shared/components/`
- `shared/pipes/`

### Observação 1.4

Cada diretório vazio pode ter um `.gitkeep` para garantir versionamento no Git. Ou, se a Fase 2 executar `ng new` e criar a estrutura, alguns desses diretórios podem ser criados pelo Angular ao gerar componentes. Para Fase 1 isolada (apenas pastas), usar `.gitkeep`.

---

### Decisão final Fase 1

**Opção A (Fase 1 pura):** Criar apenas a árvore de diretórios em `frontend/` com `.gitkeep`, sem `ng new`. A estrutura base (`package.json`, `angular.json`) virá na Fase 2.

**Opção B (Fase 1 + scaffold mínimo):** Executar `ng new` na Fase 1 para ter a base e, em seguida, criar os subdiretórios adicionais. O `ng new` é considerado parte da "estrutura" e não da "implementação".

Recomendação: **Opção A** para manter paralelismo com a feature 001 (Fase 1 = só diretórios). Na Fase 2, `ng new` criará `src/`, `app/`, etc., e o implementador adicionará os subdiretórios que faltarem. Se `ng new` for executado na Fase 2, a estrutura `frontend/` existirá e o CLI pode sobrescrever — portanto, **Fase 1 deve criar `frontend/` vazio ou com subpastas que o `ng new` não cria**. O `ng new --directory=frontend` cria a pasta `frontend/` inteiramente. Então, na Fase 1, **criar manualmente a pasta `frontend/`** e dentro dela apenas os diretórios que desejamos, **sem** `ng new`. O `ng new` na Fase 2 substituiria ou mesclaria. 

**Ajuste final:** Na Fase 1, criar `frontend/` e dentro dela a estrutura de diretórios esperada. Como o `ng new` cria sua própria estrutura, o plano mais limpo é:
- **Fase 1:** Criar `frontend/` com a árvore de subdiretórios conforme guidelines. Usar `.gitkeep` onde necessário. **Não** executar `ng new` — isso será feito na Fase 2. A Fase 1 estabelece o "esqueleto" desejado.
- **Fase 2:** Executar `ng new frontend --directory=.` **dentro** de `frontend/` (em um dir vazio) ou `ng new wisa-crm-frontend --directory=frontend` na raiz. O Angular CLI pode criar `frontend/` do zero — nesse caso, a Fase 1 teria criado `frontend/` com subpastas que seriam **fundidas** ou o `ng new` seria executado em `frontend/` vazio. A abordagem mais segura: Fase 1 cria `frontend/` com subpastas `src/app/...`; Fase 2 executa `ng new` com `--directory=.` de forma que popule `frontend/` sem conflito. Na verdade, `ng new my-app` cria `my-app/` com todo conteúdo. Para ter `frontend/` como pasta, usa-se `ng new frontend` — isso cria `frontend/` com tudo dentro.

**Decisão pragmática:** Fase 1 cria a estrutura de pastas dentro de `frontend/src/app/` **assumindo que `frontend/` já existirá com o resultado do `ng new`**. Ou seja: Fase 2 roda `ng new` primeiro; Fase 1 documenta quais diretórios adicionais criar **após** o `ng new`. A ordem de execução seria: Fase 2 (ng new) → Fase 1 (criar diretórios adicionais). Para evitar confusão, **reordena-se logicamente**: Fase 1 = "Criar estrutura" = executar `ng new` + criar diretórios adicionais em um único passo. Assim, a Fase 1 inclui tanto o scaffold do Angular quanto os diretórios extra. O requisito "não implemente código" refere-se a lógica de negócio — o scaffold do Angular é infraestrutura da estrutura.

**Decisão final:** Fase 1 cria a estrutura completa: (1) executar `ng new frontend` para obter base; (2) criar os subdiretórios `core/`, `domain/`, `application/`, `infrastructure/`, `features/`, `shared/` dentro de `src/app/`; (3) não implementar componentes de login, auth, etc. Apenas a estrutura vazia.

---

### Checklist de Implementação

1. [ ] Garantir Node.js v20+ e Angular CLI v20 instalado (`npm install -g @angular/cli`)
2. [ ] Na raiz do projeto, executar `ng new frontend --directory=frontend --routing --style=scss` (ou equivalente para Angular 20)
3. [ ] Após criação, adicionar dentro de `frontend/src/app/`:
   - [ ] `core/auth/`
   - [ ] `core/http/`
   - [ ] `domain/models/`
   - [ ] `domain/ports/`
   - [ ] `application/use-cases/`
   - [ ] `infrastructure/http/`
   - [ ] `infrastructure/storage/`
   - [ ] `features/auth/`
   - [ ] `features/home/`
   - [ ] `shared/components/`
   - [ ] `shared/pipes/`
4. [ ] Adicionar `.gitkeep` em diretórios vazios (opcional)
5. [ ] Validar que `ng build` (ou `ng serve`) executa sem erros com a estrutura base

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docs/code_guidelines/frontend.md | Estrutura por features + camadas (core, domain, application, infrastructure, shared) |
| ADR-002 Angular v20+ | Projeto criado com Angular 20+ e standalone como padrão |
| ADR-005 Clean Architecture | Separação domain / application / infrastructure |
| Segurança | Nenhum arquivo sensível; estrutura não expõe dados |

---

## Referências

- [docs/code_guidelines/frontend.md](../../../code_guidelines/frontend.md)
- [docs/adrs/ADR-002-angular-como-framework-frontend.md](../../../adrs/ADR-002-angular-como-framework-frontend.md)
