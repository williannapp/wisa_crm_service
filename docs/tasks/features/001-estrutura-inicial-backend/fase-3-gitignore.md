# Fase 3 — Adicionar .gitignore

## Objetivo

Adicionar o arquivo `.gitignore` para evitar que arquivos desnecessários, temporários, sensíveis ou gerados sejam enviados ao repositório.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O projeto possui backend em Go e documentação em Markdown. Um `.gitignore` deve cobrir:
- Artefatos de build do Go (binários, cache)
- Arquivos de ambiente com credenciais
- Arquivos do IDE/editor
- Arquivos temporários e de sistema
- Chaves privadas e secrets

### Ação 3.1

Pesquisar padrões de `.gitignore` para projetos Go e multi-componente.

### Observação 3.1

O [GitHub gitignore Go](https://github.com/github/gitignore/blob/main/Go.gitignore) é referência. Inclui: `*.exe`, `*.exe~`, `*.dll`, `*.so`, `*.dylib`, `*.test`, `*.out`, `/vendor/`, `go.work`, diretórios de IDE, etc.

---

### Pensamento 2

O requisito menciona ".gitignore" no singular. O projeto tem `backend/` e possivelmente futuro `frontend/` (Angular). Duas abordagens:
A) Um único `.gitignore` na raiz cobrindo Go, Node/Angular, IDE, etc.
B) `.gitignore` na raiz + um específico em `backend/` se houver particularidades.

Adota-se a **abordagem B**: `.gitignore` na raiz para regras gerais (IDE, OS, ambiente, frontend futuro) e um `.gitignore` específico em `backend/` para artefatos próprios do Go (binários, vendor, go.work, testes, etc.). Isso mantém as particularidades do backend isoladas e facilita a adição futura de um frontend com suas próprias regras em `frontend/.gitignore`.

### Ação 3.2

Criar dois arquivos `.gitignore`:
1. **Raiz** (`wisa-crm-service/.gitignore`): regras gerais — ambiente e secrets, IDEs, arquivos de sistema, futuro frontend Angular
2. **Backend** (`backend/.gitignore`): regras específicas do Go — binários, `vendor/`, `go.work`, artefatos de teste e build

### Observação 3.2

A divisão permite que o backend tenha suas regras (ex.: `*.test`, `wisa-crm-service` binário) sem poluir a raiz. O `.gitignore` da raiz continua cobrindo `.env`, `*.pem`, IDEs, etc., que são transversais ao projeto.

---

### Pensamento 3

Segurança crítica (ADR-009, contexto do projeto): chaves privadas, `.env` com credenciais e arquivos de configuração sensíveis **nunca** devem ser commitados.

### Ação 3.3

Garantir entradas explícitas:
```
.env
.env.local
.env.*.local
*.pem
*.key
secrets/
```

### Observação 3.3

O `.env.example` (template sem valores reais) **deve** ser versionado. O `.gitignore` não deve ignorar `.env.example`. A regra `.env` ignora apenas o arquivo literal `.env`; `.env.example` continua rastreável.

---

### Pensamento 4

O requisito original diz "Criar um `.venv`" — em projetos Python, `.venv` é o diretório do ambiente virtual. Em Go não há `.venv`. O `.gitignore` deve incluir `.venv` por precaução, caso o projeto tenha scripts Python ou ferramentas auxiliares que usem venv.

### Ação 3.4

Adicionar `.venv/` ao `.gitignore` para compatibilidade e caso existam utilitários Python.

### Observação 3.4

Consistente com boas práticas; não causa efeitos colaterais no backend Go.

---

### Conteúdo Proposto do .gitignore

**Raiz (`/.gitignore`):**
```
# Binaries (genérico)
*.exe
*.exe~
*.dll
*.so
*.dylib

# Environment and secrets
.env
.env.local
.env.*.local
*.pem
*.key
secrets/
.venv/

# IDEs
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Angular / Frontend (para fase futura)
node_modules/
dist/
.angular/
*.tgz

# Build artifacts (genérico)
build/
out/
```

**Backend (`backend/.gitignore`):**
```
# Go — binários e artefatos
wisa-crm-service
wisa-crm-service.exe
*.test
*.out
vendor/
go.work
go.work.sum
```

### Observação sobre caminhos

No `backend/.gitignore`, os caminhos são relativos ao próprio `backend/`. O binário `wisa-crm-service` gerado em `backend/` será ignorado corretamente. Ajustar conforme o nome real do binário definido no Makefile ou script de build.

---

### Checklist de Implementação

1. [ ] Criar ou atualizar `.gitignore` na raiz do projeto (regras gerais)
2. [ ] Criar ou atualizar `.gitignore` em `backend/` (regras específicas do Go)
3. [ ] Na raiz: incluir seções Environment/Secrets, IDEs, OS, Angular (opcional)
4. [ ] Em backend: incluir binários, vendor/, go.work, artefatos de teste
5. [ ] Garantir que `.env` está ignorado e `.env.example` **não** está
6. [ ] Verificar que nenhum arquivo sensível já commitado será removido (usar `git rm --cached` se necessário)
7. [ ] Validar com `git status` que arquivos esperados não aparecem como untracked indesejados

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-009 (chaves privadas) | *.pem, *.key, .env ignorados |
| Segurança | Nenhum secret versionado |
| Code Guidelines | Nenhum conflito com estrutura |

---

## Referências

- [ADR-009 — Infraestrutura VPS](../../../adrs/ADR-009-infraestrutura-vps-linux.md)
- [GitHub Go .gitignore](https://github.com/github/gitignore/blob/main/Go.gitignore)
