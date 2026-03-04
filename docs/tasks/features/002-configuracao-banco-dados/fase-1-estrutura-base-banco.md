# Fase 1 — Estrutura Base do Banco de Dados

## Objetivo

Adicionar as configurações básicas do banco de dados na estrutura do projeto: criar pastas e classes necessárias para execução de queries, realizar os imports necessários. **Não implementar nenhuma query ou tabela** — apenas a estrutura base para que a aplicação consiga se conectar ao banco.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-003 define PostgreSQL 16+ como banco. O ADR-004 define GORM como ORM. O code_guidelines/backend.md e ADR-005 definem Clean Architecture com camadas: domain, usecase, infrastructure/persistence. A **estrutura base** deve permitir conexão sem implementar repositórios ou entidades ainda.

### Ação 1.1

Identificar o que "estrutura base" significa na prática:

- Um pacote ou função que **abre a conexão** com PostgreSQL via GORM
- Configuração do **pool de conexões** (conforme guidelines: MaxOpenConns, MaxIdleConns, ConnMaxLifetime, ConnMaxIdleTime)
- **Injeção** do `*gorm.DB` no `main.go` para uso futuro pelos repositórios
- Nenhuma interface de Repository, entidade ou tabela nesta fase

### Observação 1.1

O domínio **não** deve importar GORM. A conexão vive em `infrastructure/persistence` ou em um pacote dedicado como `infrastructure/database`. O `main.go` importa a infra e obtém a instância.

---

### Pensamento 2

Onde posicionar o código de conexão? O code_guidelines mostra `infrastructure/persistence/` para implementações GORM (repositórios). A conexão em si pode ficar em:

1. **`infrastructure/persistence/database.go`** — mesma pasta dos futuros repositórios, pois a conexão é o recurso compartilhado
2. **`infrastructure/database/`** — pacote dedicado para conexão

A abordagem mais comum em Clean Architecture é ter `internal/infrastructure/database/` ou `internal/infrastructure/persistence/database.go`. O `gorm_user_repository.go` receberá o `*gorm.DB` via construtor. Para manter simplicidade, usar **`infrastructure/persistence/database.go`** — um único pacote de persistência contendo a função de conexão e os futuros repositórios.

### Ação 1.2

Definir a estrutura de arquivos:

```
backend/internal/infrastructure/persistence/
├── database.go    # Função NewDatabase(dsn string) (*gorm.DB, error)
└── .gitkeep ou placeholder para futuros gorm_*_repository.go
```

O `database.go` deve:
1. Receber DSN (connection string) como parâmetro
2. Abrir conexão via `gorm.Open(postgres.Open(dsn), &gorm.Config{})`
3. Configurar o pool via `sqlDB.SetMaxOpenConns(25)`, etc.
4. Retornar `*gorm.DB` ou erro

### Observação 1.2

Conforme ADR-004 e guidelines: usar `db.WithContext(ctx)` em todas as queries. O pool de conexões segue as recomendações do backend.md. O driver PostgreSQL para GORM é `gorm.io/driver/postgres`.

---

### Pensamento 3

O `main.go` atual não depende de banco. Para esta fase, o objetivo é que a **aplicação consiga se conectar** — ou seja, ao iniciar, a aplicação deve tentar conectar e, se falhar, pode encerrar com log fatal (conforme guidelines: "Fatal: apenas no startup, ex: banco inacessível"). Alternativamente, em desenvolvimento, pode-se permitir que a app inicie sem banco e o `/health` não verifique o banco ainda. O requisito diz "estrutura base para que a aplicação consiga se conectar" — isto implica que a conexão **deve** ser estabelecida. Se `DATABASE_URL` estiver vazio, a aplicação não deve conectar (ou deve falhar explicitamente).

### Ação 1.3

Regra para `main.go`:

- Ler `DATABASE_URL` do ambiente
- Se vazio em `production`, falhar com log fatal
- Em `development`, se vazio: log de aviso e **não** inicializar o DB (aplicação sobe sem banco, útil para desenvolvimento sem Docker)
- Se presente: chamar `persistence.NewDatabase(databaseURL)` e armazenar em variável `db` para uso futuro

### Observação 1.3

Flexibilidade em dev evita bloquear desenvolvedores que rodam apenas o backend sem subir o PostgreSQL. Em produção, `DATABASE_URL` é obrigatório (ADR-009).

---

### Pensamento 4

Imports necessários no projeto:

- `gorm.io/gorm` — core do GORM
- `gorm.io/driver/postgres` — driver PostgreSQL para GORM

Estes devem ser adicionados ao `go.mod` via `go get`.

### Ação 1.4

Executar:
```
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
```

### Observação 1.4

GORM usa o driver `pgx` ou `lib/pq` por baixo. O `gorm.io/driver/postgres` usa `github.com/jackc/pgx/v5` por padrão.

---

### Pensamento 5

Verificar conformidade com ADR-003 (PostgreSQL), ADR-004 (GORM), ADR-005 (Clean Architecture) e guidelines. A conexão não deve expor credenciais em logs. O DSN contém user:password — nunca logar o DSN completo.

### Ação 1.5

No `database.go`, em caso de erro na conexão, logar apenas a mensagem de erro, não o DSN. Em caso de sucesso, logar algo genérico como "Database connected successfully".

### Observação 1.5

Segurança: credenciais nunca em logs (guidelines e ADR-009).

---

### Decisão final Fase 1

**Implementar:**
1. `backend/internal/infrastructure/persistence/database.go` com função `NewDatabase(dsn string) (*gorm.DB, error)` que configura pool e retorna GORM DB
2. Adicionar GORM e driver postgres ao `go.mod`
3. No `main.go`: ler `DATABASE_URL`, e se presente, chamar `NewDatabase` e armazenar `db` (não injetar em handlers ainda — apenas conexão estabelecida)
4. Se `APP_ENV=production` e `DATABASE_URL` vazio: `log.Fatal("DATABASE_URL is required in production")`
5. Se `development` e `DATABASE_URL` vazio: log de aviso, `db = nil`, aplicação sobe normalmente

---

### Checklist de Implementação

1. [ ] Adicionar `gorm.io/gorm` e `gorm.io/driver/postgres` ao go.mod
2. [ ] Criar `backend/internal/infrastructure/persistence/database.go` com `NewDatabase(dsn string) (*gorm.DB, error)`
3. [ ] Configurar pool: MaxOpenConns(25), MaxIdleConns(10), ConnMaxLifetime(5*time.Minute), ConnMaxIdleTime(2*time.Minute)
4. [ ] No `main.go`: carregar DATABASE_URL, aplicar regra production/development acima
5. [ ] Garantir que nenhum log exponha credenciais ou DSN completo
6. [ ] Validar que `go build ./...` compila

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-003 PostgreSQL | Driver postgres configurado |
| ADR-004 GORM | GORM como ORM, pool configurado |
| ADR-005 Clean Architecture | Conexão em infrastructure/persistence |
| Code Guidelines backend.md | Pool conforme seção 12 |
| Segurança | Credenciais não logadas |
| Restrição do item | Nenhuma query ou tabela implementada |

---

## Referências

- [docs/code_guidelines/backend.md](../../../code_guidelines/backend.md)
- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [docs/adrs/ADR-004-gorm-como-orm.md](../../../adrs/ADR-004-gorm-como-orm.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../../adrs/ADR-005-clean-architecture.md)
