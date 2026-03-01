# ADR-004 — GORM como ORM

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (camada de dados)

---

## Contexto

O backend em Go do `wisa-crm-service` precisa interagir com o PostgreSQL para todas as operações de persistência. As operações incluem:

- CRUD de tenants, usuários e assinaturas
- Consultas filtradas por `tenant_id` em todas as entidades
- Transações atômicas (ex: criar usuário + vincular assinatura)
- Execução de migrations de schema de forma versionada
- Consultas para audit log com filtros complexos de data e tipo de evento

A escolha da estratégia de acesso ao banco impacta diretamente:
- Segurança contra SQL Injection
- Manutenibilidade e legibilidade do código
- Performance das queries geradas
- Facilidade de testes unitários da camada de dados
- Aderência aos princípios de Clean Architecture

Existe uma tensão importante a considerar: **ORMs facilitam o desenvolvimento mas podem gerar queries ineficientes e abstrair demais o banco**, enquanto **SQL puro oferece controle total mas aumenta verbosidade e risco de erros de parametrização**.

---

## Decisão

**GORM v2 é o ORM escolhido para o `wisa-crm-service`.**

O uso do GORM será **disciplinado e bounded**: operações simples de CRUD usarão a API do GORM, enquanto queries complexas, críticas de performance ou que envolvam Row-Level Security utilizarão SQL raw parametrizado via `db.Raw()` ou `db.Exec()`.

A camada de acesso a dados será encapsulada em **Repositories** (padrão Repository da Clean Architecture), isolando completamente o GORM da lógica de domínio.

---

## Justificativa

### 1. Produtividade sem sacrificar controle

GORM v2 oferece uma API fluente que cobre 80–90% dos casos de uso sem escrever SQL manual:

```go
// Consulta com filtro multi-tenant — segura por padrão (query parametrizada)
var user User
result := db.Where("tenant_id = ? AND email = ?", tenantID, email).First(&user)
```

Para os 10–20% de casos complexos (queries de auditoria, relatórios de assinatura, agregações), GORM permite SQL raw parametrizado:

```go
// SQL raw com parametrização — seguro contra SQL Injection
db.Raw("SELECT u.*, s.expires_at FROM users u JOIN subscriptions s ON s.tenant_id = u.tenant_id WHERE u.tenant_id = ? AND s.status = 'active'", tenantID).Scan(&result)
```

Essa combinação oferece produtividade para o common case e controle para casos especiais.

### 2. Migrations versionadas nativas

GORM oferece **AutoMigrate** para desenvolvimento, mas para produção a abordagem recomendada é migrations explícitas. O ecossistema Go recomenda usar o `golang-migrate` em conjunto com GORM:

```
migrations/
  000001_create_tenants.up.sql
  000001_create_tenants.down.sql
  000002_create_users.up.sql
  000002_create_users.down.sql
  ...
```

Isso garante:
- Migrations versionadas e reversíveis
- Auditoria de mudanças de schema via git
- Deploy seguro com possibilidade de rollback

**AutoMigrate do GORM não deve ser usado em produção** — ele não detecta remoção de colunas e pode causar comportamento inesperado.

### 3. Proteção contra SQL Injection por padrão

GORM usa prepared statements e parametrização em todas as operações da API fluente. Um desenvolvedor precisaria sair do caminho padrão para introduzir SQL Injection:

```go
// SEGURO — parametrizado automaticamente
db.Where("email = ?", userInput).Find(&users)

// INSEGURO — deve ser proibido em code review
db.Where("email = '" + userInput + "'").Find(&users) // ← Nunca fazer isso
```

O risco está no `db.Raw()` com concatenação de strings — mitigado por regra explícita de code review e linting estático.

### 4. Suporte a transações explícitas

Para operações críticas de auth, GORM suporta transações explícitas com rollback automático em erros:

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // 1. Valida assinatura
    if err := tx.Where("tenant_id = ? AND status = 'active'", tenantID).First(&sub).Error; err != nil {
        return err // rollback automático
    }
    // 2. Registra evento de login
    if err := tx.Create(&auditEvent).Error; err != nil {
        return err // rollback automático
    }
    return nil // commit
})
```

### 5. Hooks para cross-cutting concerns

GORM permite hooks (`BeforeCreate`, `BeforeUpdate`, `AfterFind`) que podem ser usados para:

- Enforçar que `tenant_id` esteja presente antes de qualquer `INSERT`
- Auditar automaticamente `created_at` e `updated_at`
- Sanitizar campos antes de persistência

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.TenantID == uuid.Nil {
        return errors.New("tenant_id is required")
    }
    return nil
}
```

### 6. Integração com padrão Repository da Clean Architecture

GORM é facilmente encapsulado em interfaces de Repository:

```go
// Interface no domínio — sem dependência de GORM
type UserRepository interface {
    FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
    Create(ctx context.Context, user *User) error
    // ...
}

// Implementação na camada de infraestrutura — conhece GORM
type gormUserRepository struct {
    db *gorm.DB
}

func (r *gormUserRepository) FindByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error) {
    var user User
    result := r.db.WithContext(ctx).Where("tenant_id = ? AND email = ?", tenantID, email).First(&user)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, domain.ErrUserNotFound
    }
    return &user, result.Error
}
```

A lógica de domínio **nunca importa GORM diretamente** — apenas as implementações de Repository conhecem o ORM.

---

## Consequências

### Positivas
- Produtividade no desenvolvimento com API fluente
- SQL Injection prevenido por padrão na API principal
- Migrations explícitas garantem controle do schema em produção
- Encapsulamento via Repository mantém o domínio desacoplado da tecnologia de banco
- Transações explícitas garantem atomicidade em operações críticas

### Negativas
- N+1 query problem se não houver disciplina no uso de `Preload` — requer revisão ativa
- Queries geradas pelo GORM podem ser subótimas em casos complexos — mitigado com SQL raw quando necessário
- Reflexão (reflection) usada pelo GORM pode introduzir bugs sutis com tipos customizados
- AutoMigrate tentador para novos desenvolvedores — requer disciplina para não usar em produção

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| SQL Injection via `db.Raw()` com concatenação | Baixa | Crítico | Alta |
| N+1 queries degradando performance em pico | Média | Médio | Média |
| AutoMigrate usado acidentalmente em produção | Baixa | Alto | Média |
| Vazamento de conexões de banco por contexto sem timeout | Média | Médio | Média |
| Bypass de tenant_id por omissão em query GORM | Média | Crítico | Alta |

---

## Mitigações

### SQL Injection
- Lint estático com `gosec` para detectar concatenação de strings em `db.Raw()`
- Regra de code review: qualquer `db.Raw()` ou `db.Exec()` com variável concatenada é bloqueante
- Testar com inputs maliciosos nos testes de integração dos repositories

### N+1 queries
- Monitorar queries lentas via `pg_stat_statements` e `log_min_duration_statement`
- Usar `db.Preload()` explicitamente para associações necessárias
- Implementar `db.Joins()` para joins de alta frequência ao invés de Preload em N registros

### AutoMigrate em produção
- Remover qualquer chamada a `db.AutoMigrate()` do código de produção via lint
- Migrations gerenciadas exclusivamente via `golang-migrate`
- Processo de deploy documentado: "sempre rodar `migrate up` antes de iniciar novo binário"

### Contexto sem timeout
- Sempre usar `db.WithContext(ctx)` passando o contexto da requisição HTTP
- Configurar `ConnMaxLifetime` e `ConnMaxIdleTime` no pool de conexões:

```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
sqlDB.SetConnMaxIdleTime(2 * time.Minute)
```

### Bypass de tenant_id
- Implementar um **Global Scope** GORM que injeta automaticamente o filtro de tenant:

```go
func TenantScope(tenantID uuid.UUID) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("tenant_id = ?", tenantID)
    }
}
// Uso: db.Scopes(TenantScope(tenantID)).Find(&users)
```

- Testes de integração que verificam isolamento: buscar dados de tenant A com contexto de tenant B deve retornar vazio

---

## Padrões Obrigatórios de Uso

### 1. Sempre usar contexto
```go
// CORRETO
db.WithContext(ctx).Where(...).Find(...)

// INCORRETO — sem timeout, conexão pode ficar pendurada
db.Where(...).Find(...)
```

### 2. Nunca concatenar strings em queries
```go
// CORRETO
db.Where("email = ?", email)

// INCORRETO — SQL Injection
db.Where(fmt.Sprintf("email = '%s'", email))
```

### 3. Sempre verificar erros de não-encontrado
```go
if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, domain.ErrNotFound // mapear para erro de domínio
}
```

### 4. Usar transações para operações compostas
```go
// Qualquer operação que afete múltiplas tabelas deve usar transação
db.Transaction(func(tx *gorm.DB) error {
    // ...
})
```

---

## Alternativas Consideradas

### sqlx (SQL puro com scanning)
- **Prós:** Controle total de SQL, sem magia de reflection, performance máxima
- **Contras:** Verbosidade extrema para CRUD simples, sem suporte nativo a migrations, maior risco de erros de parametrização por omissão, sem hooks para cross-cutting concerns

### pgx (driver PostgreSQL nativo)
- **Prós:** Performance máxima, suporte nativo a tipos PostgreSQL avançados, prepared statements no protocolo de wire
- **Contras:** Muito baixo nível para operações cotidianas, sem abstração de ORM, código mais verboso, mais adequado como driver subjacente ao GORM/sqlx

### ent (entgo.io)
- **Prós:** Type-safe, geração de código, schema como código Go
- **Contras:** Geração de código adiciona complexidade ao build, curva de aprendizado específica, menos adotado que GORM, integração com PostgreSQL Row-Level Security mais complexa

### Decisão combinada
**GORM v2 com disciplina de uso oferece o melhor equilíbrio para o estágio atual do projeto.** A interface de Repository isola a lógica de domínio, e as regras de uso documentadas mitigam os principais riscos do ORM.

---

## Referências

- [GORM v2 Documentation](https://gorm.io/docs/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [GORM Security Practices](https://gorm.io/docs/security.html)
- [pgx - PostgreSQL Driver for Go](https://github.com/jackc/pgx)
