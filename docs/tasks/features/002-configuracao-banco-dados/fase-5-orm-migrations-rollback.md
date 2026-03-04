# Fase 5 — ORM, Migrations e Rollback

## Objetivo

Realizar as adaptações no código para o ORM, adicionar a estrutura base do ORM ao projeto, implementar as migrations versionadas e implementar a funcionalidade de rollback do ORM.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-004 define GORM como ORM e **golang-migrate** para migrations (não AutoMigrate em produção). O code_guidelines especifica:
- Migrations em `migrations/` com nomenclatura `000001_descricao.up.sql` e `000001_descricao.down.sql`
- Não usar AutoMigrate em produção
- Migrations versionadas e reversíveis

A "estrutura base do ORM" já foi parcialmente coberta na Fase 1 (conexão GORM). Nesta fase, focar em:
1. Integração golang-migrate no projeto
2. Estrutura de pastas para migrations
3. Comandos/makefile para rodar migrate up e migrate down
4. Primeira migration (pode ser vazia ou criar tabela de controle)
5. Script ou binário para executar migrations no startup ou via CLI

### Ação 5.1

Definir os artefatos da Fase 5:

| Item | Implementação |
|------|---------------|
| Migrations | `backend/migrations/` com arquivos .up.sql e .down.sql |
| Ferramenta | golang-migrate CLI ou biblioteca embed |
| Rollback | migrate down 1 (ou down all) |
| Makefile/scripts | `make migrate-up`, `make migrate-down` |
| Migration inicial | 000001_init_schema.up.sql vazio ou com extensão uuid; 000001_init_schema.down.sql vazio |

### Observação 5.1

A migration inicial pode ser mínima: criar extensão `uuid-ossp` se necessário, ou criar uma tabela `schema_migrations` (o golang-migrate já gerencia isso internamente). Uma migration vazia é válida para estabelecer o baseline.

---

### Pensamento 2

O golang-migrate pode ser usado de duas formas:
1. **CLI** — binário `migrate` instalado no sistema, executado com `migrate -path ./migrations -database $DATABASE_URL up`
2. **Biblioteca Go** — importar `github.com/golang-migrate/migrate/v4` e executar programaticamente

Para o projeto, a abordagem CLI é mais comum e desacopla o processo de migration do binário da aplicação. O deploy pode rodar migrations como passo separado antes de iniciar o backend. Alternativamente, o backend pode rodar migrations no startup (menos recomendado para produção — preferir passo explícito).

### Ação 5.2

Adotar **CLI do golang-migrate** para executar migrations. O desenvolvedor e o processo de deploy executam manualmente:
- `migrate -path backend/migrations -database "$DATABASE_URL" up`
- `migrate -path backend/migrations -database "$DATABASE_URL" down 1`

Criar um **Makefile** na raiz ou em `backend/` com targets:
- `migrate-up`: sobe todas as migrations pendentes
- `migrate-down`: desce a última migration
- `migrate-down-all`: desce todas (cuidado em produção)

Opção: instalar migrate via `go install` e documentar no README/run-backend.

### Observação 5.2

O golang-migrate não é uma dependência do backend em runtime — é ferramenta de desenvolvimento/deploy. O `go.mod` do backend pode incluir `github.com/golang-migrate/migrate/v4` se quisermos um comando `go run ./cmd/migrate` que execute as migrations. Isso permitiria rodar migrations sem instalar o CLI. Ambas abordagens são válidas. Preferir **CLI** para simplicidade; o Makefile usa `migrate` se disponível no PATH, ou instrui a instalação.

### Ação 5.2 (revisão)

Incluir `golang-migrate` como dependência e criar `backend/cmd/migrate/main.go` que:
- Lê DATABASE_URL do env
- Executa migrate up por padrão
- Aceita flag `-down` para rollback

Isso evita dependência de ter o CLI instalado. O comando seria: `go run ./cmd/migrate` ou `go run ./cmd/migrate -down`.

---

### Pensamento 3

Estrutura das migrations. O golang-migrate suporta:
- Arquivos SQL: `000001_name.up.sql`, `000001_name.down.sql`
- Ou numbering: `1_up.sql`, `1_down.sql`

Conforme guidelines: `000001_descricao.up.sql`, `000001_descricao.down.sql`.

Para a migration inicial (baseline):
- `000001_init_schema.up.sql` — pode criar extensão uuid-ossp (CREATE EXTENSION IF NOT EXISTS "uuid-ossp";) ou ficar vazio
- `000001_init_schema.down.sql` — DROP EXTENSION IF EXISTS "uuid-ossp"; ou vazio

Não criar tabelas de domínio nesta fase — o requisito da feature é "estrutura base" e "implementar migrations/rollback". A primeira migration estabelece o padrão; tabelas de negócio serão adicionadas em features futuras.

### Ação 5.3

Criar `backend/migrations/000001_init_schema.up.sql`:
```sql
-- Migration inicial - baseline
-- Extensão para geração de UUIDs (nativo no PG13+ como gen_random_uuid())
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

Criar `backend/migrations/000001_init_schema.down.sql`:
```sql
DROP EXTENSION IF EXISTS "uuid-ossp";
```

### Observação 5.3

O PG 13+ tem `gen_random_uuid()` nativo, mas `uuid-ossp` oferece `uuid_generate_v4()` para compatibilidade. Manter a extensão para flexibilidade. O down remove a extensão — em ambiente limpo, funciona; se já houver tabelas usando uuid, o down falharia. Para a migration inicial isolada, é seguro.

---

### Pensamento 4

Implementação do comando de migrate. Duas opções finais:

**Opção A: CLI externo**
- Desenvolvedor instala: `go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`
- Makefile chama `migrate -path ... -database ... up`

**Opção B: Comando Go interno**
- `backend/cmd/migrate/main.go` usa a lib migrate para ler os arquivos e executar
- As migrations podem ser embedadas no binário via `//go:embed migrations/*.sql`
- Comando: `go run ./cmd/migrate` (up) ou com flag para down

A Opção B torna o projeto autossuficiente — não requer instalação externa. O binário de produção pode ter um subcomando ou um binário separado `wisa-crm-migrate` para rodar migrations. Para simplificar, criar um `cmd/migrate` que faz apenas up/down.

### Ação 5.4

Implementar `backend/cmd/migrate/main.go`:

```go
package main

import (
    "flag"
    "log"
    "os"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)
// Usa source/file apontando para ./migrations (relativo ao CWD)
// Lê DATABASE_URL do env
// Flag -down para rollback 1 step
```

Dependências: `go get github.com/golang-migrate/migrate/v4` e `go get github.com/golang-migrate/migrate/v4/database/postgres` e `go get github.com/golang-migrate/migrate/v4/source/file`.

O path das migrations pode ser `file://migrations` relativo ao CWD, ou absoluto. Garantir que ao rodar `go run ./cmd/migrate` de dentro de `backend/`, o path `./migrations` seja correto.

### Observação 5.4

O migrate requer que o path seja absoluto ou file:// com path absoluto em alguns casos. Verificar documentação. Alternativa: usar `source/file` com path `file://./migrations` — o migrate suporta paths relativos ao working directory.

---

### Pensamento 5

Adaptações no código para o ORM. A Fase 1 já adicionou GORM e a conexão. "Adaptações" podem significar:
- Garantir que os repositórios (quando criados) usem o padrão correto (WithContext, parametrização)
- Configurar o GORM para logging em desenvolvimento (opcional)
- Nenhuma entidade ou repositório nesta fase — apenas a infraestrutura de conexão e migrations

O code_guidelines já define os padrões. Não há adaptação adicional necessária além do que foi feito na Fase 1. A "estrutura base do ORM" = conexão + capacidade de migrations.

### Ação 5.5

Não criar entidades ou repositórios na Fase 5. A estrutura base do ORM inclui:
- Conexão GORM (Fase 1)
- Migrations com golang-migrate
- Comando migrate para up/down
- Makefile com targets de conveniência

### Observação 5.5

Restrição respeitada: apenas estrutura, sem tabelas de negócio.

---

### Pensamento 6

Rollback: o golang-migrate suporta `migrate.Down()` para descer uma ou todas as migrations. O comando `cmd/migrate` com flag `-down` executará `m.Down()` (1 step) ou `m.Steps(-1)`.

### Ação 5.6

No `cmd/migrate`, suportar:
- Sem flags: `migrate.Up()`
- Com `-down`: `migrate.Steps(-1)` (desce 1 migration)
- Opcional `-down-all`: `migrate.Down()` (desce todas)

### Observação 5.6

Rollback implementado via golang-migrate.

---

### Decisão final Fase 5

**Implementar:**
1. Adicionar `github.com/golang-migrate/migrate/v4` e drivers ao go.mod
2. Criar `backend/migrations/000001_init_schema.up.sql` (extensão uuid-ossp)
3. Criar `backend/migrations/000001_init_schema.down.sql` (drop extensão)
4. Criar `backend/cmd/migrate/main.go` que executa migrate up (padrão) ou down (flag -down)
5. Criar Makefile em `backend/` ou na raiz com `migrate-up` e `migrate-down`
6. Atualizar documentação (Fase 4) com comandos de migration
7. Garantir que migrate use DATABASE_URL do ambiente

---

### Checklist de Implementação

1. [ ] `go get github.com/golang-migrate/migrate/v4 github.com/golang-migrate/migrate/v4/database/postgres github.com/golang-migrate/migrate/v4/source/file`
2. [ ] Criar `backend/migrations/000001_init_schema.up.sql`
3. [ ] Criar `backend/migrations/000001_init_schema.down.sql`
4. [ ] Criar `backend/cmd/migrate/main.go` com suporte a up e down
5. [ ] Criar Makefile com targets migrate-up e migrate-down
6. [ ] Testar: migrate up, migrate down
7. [ ] Atualizar docs/backend/README.md e docs/backend/vps-configurations.md com instruções de migration
8. [ ] Verificar que go build ./cmd/migrate e go build ./cmd/api compilam

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-004 GORM | GORM para conexão; golang-migrate para migrations |
| Code Guidelines | Migrations versionadas, nomenclatura correta |
| Não AutoMigrate em prod | Migrations explícitas via golang-migrate |
| Rollback | migrate down implementado |
| Segurança | Migrations em SQL puro, sem concatenação |

---

## Referências

- [docs/code_guidelines/backend.md](../../../code_guidelines/backend.md)
- [docs/adrs/ADR-004-gorm-como-orm.md](../../../adrs/ADR-004-gorm-como-orm.md)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [golang-migrate CLI and Library](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
