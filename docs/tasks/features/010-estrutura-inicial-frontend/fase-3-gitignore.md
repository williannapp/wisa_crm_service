# Fase 3 — Adicionar .gitignore para o Frontend

## Objetivo

Adicionar ou atualizar o arquivo `.gitignore` para evitar que arquivos desnecessários do frontend Angular (node_modules, builds, cache, artefatos de IDE) sejam enviados ao repositório.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O projeto possui `.gitignore` na raiz que já inclui entradas para frontend (conforme verificado no repositório):

```
# Node (frontend)
node_modules/
frontend/node_modules/
.npm
.yarn/
.pnp.*
frontend/dist/
frontend/.angular/
frontend/.cache/
```

Portanto, a raiz já cobre os caminhos principais. A questão é: precisa-se de um `.gitignore` específico em `frontend/`?

### Ação 3.1

Analisar o padrão do backend: o backend possui `backend/.gitignore` com regras específicas de Go. A abordagem análoga para o frontend seria ter `frontend/.gitignore` com regras específicas de Angular/Node.

### Observação 3.1

O Angular CLI, ao executar `ng new`, gera automaticamente um `.gitignore` dentro do projeto com entradas típicas para Angular. Se a Fase 2 executar `ng new`, o `frontend/.gitignore` já existirá. A Fase 3 deve **validar** que esse arquivo contém as regras necessárias e **complementar** se faltar algo, ou **criar** `frontend/.gitignore` manualmente se a estrutura foi criada sem `ng new`.

---

### Pensamento 2

Entradas obrigatórias para um projeto Angular (baseadas em Angular default + boas práticas):

| Entrada | Motivo |
|---------|--------|
| `node_modules/` | Dependências (sempre ignorar) |
| `dist/` ou `*.tsbuildinfo` | Artefatos de build |
| `.angular/` | Cache do Angular CLI |
| `.cache/` | Cache de build |
| `coverage/` | Relatórios de cobertura de testes |
| `*.log` | Logs do npm |
| `.env` e `.env.*` (exceto .example) | Variáveis de ambiente com possíveis secrets |
| `.vscode/` (opcional) | Configurações de IDE — alguns preferem versionar |
| `.idea/` | IntelliJ/WebStorm |
| `*.tgz` | Pacotes npm locais |
| `npm-debug.log*` | Logs de debug |
| `.npm` | Cache npm |
| `.yarn` | Cache Yarn |
| `package-lock.json` | **NÃO** ignorar — deve ser versionado |
| `yarn.lock` | **NÃO** ignorar se usar Yarn |

### Ação 3.2

Garantir que o `.gitignore` do frontend (ou raiz) **não** ignore `package-lock.json`. O lock file é essencial para builds determinísticos e segurança de supply chain.

### Observação 3.2

O `.gitignore` padrão do Angular **não** ignora `package-lock.json`. Verificar regras que possam afetá-lo acidentalmente (ex.: `*.lock` ignoraria ambos `package-lock.json` e `yarn.lock` — evitar).

---

### Pensamento 3

O `.gitignore` da raiz já inclui `frontend/node_modules/`, `frontend/dist/`, etc. Os caminhos são relativos à raiz do repo. Se `frontend/` for um subdiretório, as regras na raiz são suficientes para ignorar `frontend/node_modules`, etc. Um `frontend/.gitignore` adicional pode ter regras **relativas ao contexto do frontend** (ex.: `node_modules/` sem prefixo — dentro de `frontend/`, isso ignora `frontend/node_modules/`). Git aplica `.gitignore` de forma hierárquica: o `.gitignore` em `frontend/` aplica-se a arquivos dentro de `frontend/`.

### Ação 3.3

Duas abordagens:

**A) Apenas raiz:** Manter todo o `.gitignore` na raiz. O existente já cobre frontend. Nenhuma ação em `frontend/`.

**B) Raiz + frontend específico:** O `ng new` cria `frontend/.gitignore`. Manter esse arquivo e garantir que contenha as entradas necessárias. A raiz pode manter regras gerais; o frontend tem suas específicas.

### Observação 3.3

O Angular CLI cria um `.gitignore` em `frontend/` ao executar `ng new`. Esse arquivo é adequado para a maioria dos casos. A Fase 3 deve: (1) Se `frontend/.gitignore` existir (via ng new), revisá-lo e complementar se necessário; (2) Se não existir, criar `frontend/.gitignore` com as entradas padrão. Também verificar que a raiz não duplica regras de forma conflitante.

---

### Pensamento 4

Segurança (ADR-002, ADR-009): arquivos de ambiente com possíveis secrets não devem ser commitados. O frontend pode ter `environment.ts` (dev) e `environment.prod.ts` — normalmente versionados com valores placeholder. Se houver `.env` para configurar API URL em desenvolvimento, este deve ser ignorado. O `.gitignore` da raiz já inclui `.env`. Garantir que `frontend/.env` também seja ignorado (a regra `.env` na raiz pode não cobrir `frontend/.env` dependendo do path — na verdade, `.env` sem prefixo ignora qualquer `.env` em qualquer subdiretório no Git).

### Ação 3.4

Adicionar em `frontend/.gitignore` (ou validar na raiz):
```
.env
.env.local
.env.*.local
!.env.example
```
A negação `!.env.example` garante que o template seja versionado. O `.gitignore` da raiz já tem `.env`; cobrirá `frontend/.env`.

### Observação 3.4

Nenhuma mudança crítica. Apenas garantir consistência.

---

### Checklist de Implementação

1. [ ] Verificar se `frontend/.gitignore` existe (criado por `ng new`)
2. [ ] Se não existir, criar `frontend/.gitignore` com entradas padrão Angular/Node
3. [ ] Garantir inclusão de: `node_modules/`, `dist/`, `.angular/`, `.cache/`, `coverage/`, `*.log`, `.env` (sem `.env.example`)
4. [ ] Garantir que `package-lock.json` **não** está ignorado
5. [ ] Verificar que o `.gitignore` da raiz não conflita (ex.: regras que ignoram `package-lock.json` globalmente)
6. [ ] Validar com `git status` que `node_modules` e `dist` não aparecem como untracked
7. [ ] Documentar no README da feature que `.env.example` (se usado) deve ser versionado

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-009 (secrets) | .env ignorado; .env.example versionado |
| Supply chain | package-lock.json não ignorado |
| Angular best practices | Entradas padrão do CLI mantidas |

---

## Referências

- [ADR-009 — Infraestrutura VPS](../../../adrs/ADR-009-infraestrutura-vps-linux.md)
- [GitHub Node .gitignore](https://github.com/github/gitignore/blob/main/Node.gitignore)
- [Angular .gitignore default](https://github.com/angular/angular-cli/blob/main/packages/angular_devkit/build_angular/src/utils/default-gitignore.ts)
