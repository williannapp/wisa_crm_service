# Fase 4 — Documentação

## Objetivo

Criar documentação dentro de `docs/backend/` explicando como rodar o backend. Caso alguma configuração adicional seja necessária na VPS para que o código funcione conforme descrito nas ADRs, adicionar em `docs/backend/vps-configurations.md`.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O requisito tem duas partes:
1. **Documentação geral:** como rodar o backend
2. **Configurações VPS:** `docs/backend/vps-configurations.md` — apenas se houver configurações adicionais necessárias para a VPS

As ADRs (003, 004, 009) já descrevem PostgreSQL, GORM e infraestrutura VPS em alto nível. O `configurations.md` deve complementar com **passos práticos e específicos** que um operador executaria na VPS para fazer o backend funcionar com o banco.

### Ação 4.1

Definir o escopo da documentação (criar os dois documentos dentro de `docs/backend/`):

| Documento | Conteúdo |
|-----------|----------|
| `docs/backend/README.md` | Como rodar o backend em desenvolvimento local (com e sem Docker) |
| `docs/backend/vps-configurations.md` | Configurações adicionais na VPS (PostgreSQL, PgBouncer, env, etc.) quando aplicável às ADRs |

### Observação 4.1

A documentação do backend fica em `docs/backend/`. O TRACKER da Feature 001 não criou documentação específica de "como rodar". Criar agora.

---

### Pensamento 2

Conteúdo do "como rodar o backend":

1. **Pré-requisitos:** Go 1.25+, PostgreSQL 16 (ou Docker)
2. **Desenvolvimento local sem Docker:**
   - Configurar PostgreSQL localmente (ou usar instância remota)
   - Copiar `.env.example` para `.env`
   - Configurar `DATABASE_URL` em `.env`
   - `go run ./cmd/api` ou `go build && ./wisa-crm-service`
3. **Desenvolvimento local com Docker:**
   - `docker compose up postgres -d` (apenas banco)
   - Ou `docker compose up` (banco + backend)
   - Ou rodar backend local com `go run` apontando para `localhost:5432`
4. **Migrações:** como rodar `migrate up` e `migrate down` (após Fase 5)
5. **Health check:** endpoint `/health`

### Ação 4.2

Criar `docs/backend/README.md`. O título pode ser "Como Rodar o Backend" ou "Guia de Execução do Backend".

Estrutura sugerida:

```markdown
# Como Rodar o Backend

## Pré-requisitos
- Go 1.25 ou superior
- PostgreSQL 16+ (ou Docker)

## Desenvolvimento Local

### Opção 1: Com Docker (PostgreSQL em container)
1. Subir o banco: `docker compose up postgres -d`
2. Copiar variáveis: `cp backend/.env.example backend/.env`
3. Ajustar DATABASE_URL se necessário (padrão: localhost:5432)
4. Na pasta backend: `go run ./cmd/api`

### Opção 2: Tudo em Docker
1. `docker compose up`
2. Backend em http://localhost:8080
3. Health: http://localhost:8080/health

### Opção 3: PostgreSQL local (sem Docker)
1. Instalar PostgreSQL 16
2. Criar banco e usuário
3. Configurar .env com DATABASE_URL
4. `go run ./cmd/api`

## Variáveis de Ambiente
Ver backend/.env.example

## Migrações
(Seção a ser preenchida na Fase 5)

## Produção (VPS)
Ver docs/backend/vps-configurations.md
```

### Observação 4.2

Clareza e múltiplas opções para diferentes fluxos de trabalho dos desenvolvedores.

---

### Pensamento 3

Sobre `docs/backend/vps-configurations.md`: o ADR-009 já descreve hardening, systemd, UFW, etc. O `vps-configurations.md` deve ser um **guia operacional** que referencia as ADRs e detalha os passos concretos. Exemplos de conteúdo:

- Instalação do PostgreSQL 16 na VPS (apt install)
- Criação do banco `wisa_crm_db` e usuário `wisa_crm` com permissões
- Configuração do pg_hba.conf para conexões do backend (loopback)
- Arquivo `/etc/wisa-crm/env` com DATABASE_URL
- Comandos para rodar migrations em produção (migrate up antes do deploy)
- PgBouncer (se aplicável nesta fase — ADR-003 menciona para produção)

### Ação 4.3

Conteúdo mínimo de `docs/backend/vps-configurations.md`:

1. **PostgreSQL na VPS**
   - Instalação: `apt install postgresql-16`
   - Criar usuário e banco
   - Configurar pg_hba.conf para host 127.0.0.1
   - ssl = on conforme ADR-003

2. **Variáveis de ambiente do backend**
   - EnvironmentFile=/etc/wisa-crm/env
   - Exemplo do conteúdo (com placeholders para credenciais)
   - DATABASE_URL com host localhost ou 127.0.0.1

3. **Migrações em produção**
   - Ordem: rodar migrate up antes de iniciar o novo binário
   - Comando: migrate -path migrations -database "$DATABASE_URL" up
   - Rollback: migrate down 1 (quando necessário)

4. **Referências**
   - ADR-003, ADR-009

### Observação 4.3

O `vps-configurations.md` é específico para deploy na VPS. Desenvolvedores que só rodam localmente não precisam dele para desenvolvimento.

---

### Pensamento 4

A pasta `docs/backend/` pode não existir. Será criada junto com os documentos. O ADR-009 não exige um documento específico de configurações — o requisito da feature o solicita "caso alguma configuração adicional seja necessária". As configurações de PostgreSQL, env e migrations **são** necessárias para a VPS funcionar conforme as ADRs. Portanto, criar o documento.

### Ação 4.4

Criar `docs/backend/vps-configurations.md` com as seções definidas na Ação 4.3. Manter conciso; detalhes profundos ficam nas ADRs.

### Observação 4.4

Evitar duplicação: referenciar ADRs ao invés de copiar trechos longos.

---

### Pensamento 5

Indexação e descoberta: o `docs/context.md` ou um índice em `docs/README.md` poderia listar os documentos disponíveis. Para esta fase, criar os documentos principais. Um índice pode ser uma melhoria futura.

### Ação 4.5

Não criar índice de documentação nesta fase. Focar em `docs/backend/README.md` e `docs/backend/vps-configurations.md`.

### Observação 4.5

YAGNI para índice.

---

### Decisão final Fase 4

**Implementar:**
1. Criar `docs/backend/README.md` com guia completo de execução do backend
2. Criar `docs/backend/vps-configurations.md` com configurações necessárias na VPS (PostgreSQL, env, migrations)
3. Referenciar ADRs e .env.example onde apropriado
4. Incluir seção de migrações com placeholder ou conteúdo da Fase 5

---

### Checklist de Implementação

1. [ ] Criar `docs/backend/README.md`
2. [ ] Documentar pré-requisitos, opções de execução (Docker, local, híbrido)
3. [ ] Documentar variáveis de ambiente e referência ao .env.example
4. [ ] Criar `docs/backend/` (se não existir)
5. [ ] Criar `docs/backend/vps-configurations.md` com configurações VPS
6. [ ] Incluir seção de migrações (ou indicar que será adicionada na Fase 5)
7. [ ] Referenciar ADR-003, ADR-009 onde relevante

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Requisito | Documentação em docs/backend/ |
| Requisito | docs/backend/vps-configurations.md para configurações VPS |
| ADRs | Referências corretas |

---

## Dependências

- Fase 3 (Containers) — para documentar docker compose
- Fase 5 (ORM) — para documentar migrações com precisão

A documentação de migrações pode ser um placeholder até a conclusão da Fase 5.

---

## Referências

- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [docs/adrs/ADR-009-infraestrutura-vps-linux.md](../../../adrs/ADR-009-infraestrutura-vps-linux.md)
