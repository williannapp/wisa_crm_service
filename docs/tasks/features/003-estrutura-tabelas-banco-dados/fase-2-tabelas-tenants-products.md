# Fase 2 — Tabelas tenants e products

## Objetivo

Implementar as tabelas `tenants` e `products` no schema `wisa-labs-db`, conforme modelagem do ADR-008. Estas são as entidades raiz do modelo de dados, sem dependências de outras tabelas de domínio.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A tabela `tenants` representa os clientes do sistema (empresas ou pessoas que contratam o wisa-crm-service). Conforme ADR-008:
- `id` UUID PRIMARY KEY
- `slug` VARCHAR(63) UNIQUE NOT NULL — identificador amigável para URLs (ex: cliente1)
- `name` VARCHAR(255) NOT NULL
- `tax_id` VARCHAR(18) NOT NULL — CNPJ (14 dígitos) ou CPF (11 dígitos)
- `type` tenant_type NOT NULL — person ou company
- `status` tenant_status NOT NULL DEFAULT 'active'
- `created_at`, `updated_at` TIMESTAMPTZ
- CHECK constraint: (type='company' AND LENGTH(tax_id)=14) OR (type='person' AND LENGTH(tax_id)=11)

A tabela `products` representa o catálogo global de produtos/planos:
- `id` UUID PRIMARY KEY
- `slug` VARCHAR(255) NOT NULL UNIQUE
- `name` VARCHAR(255) NOT NULL
- `status` product_status NOT NULL DEFAULT 'active'
- `created_at`, `updated_at` TIMESTAMPTZ

### Ação 2.1

Definir a migration 000003 para criar as duas tabelas. Qualificação: `wisa_crm_db.tenants` e `wisa_crm_db.products`.

### Observação 2.1

O ADR-008 usa `slug VARCHAR(63)` para products; vou manter consistência. O constraint de tax_id valida CNPJ/CPF brasileiro — adequado ao contexto do projeto.

---

### Pensamento 2

A coluna `tax_id` pode conter apenas dígitos (CPF/CNPJ sem formatação) ou com pontuação. O CHECK valida o **length** — assumindo que será armazenado sem pontuação (14 ou 11 caracteres). Se o input vier com pontos e traços, a aplicação deve sanitizar antes de inserir. O CHECK com LENGTH garante integridade no banco.

### Ação 2.2

Manter o CHECK como no ADR-008. Documentar que a aplicação deve armazenar tax_id apenas com dígitos (sem formatação) para o CHECK funcionar. Alternativa: LENGTH(REGEXP_REPLACE(tax_id, '\D', '', 'g')) = 14 ou 11 — mais flexível para aceitar input formatado. Por simplicidade, manter LENGTH(tax_id) e documentar o formato esperado.

### Observação 2.2

Conformidade com ADR-008. Se no futuro for necessário aceitar tax_id formatado, pode-se criar uma migration que altera o CHECK.

---

### Pensamento 3

Triggers de `updated_at`: o ADR-008 define a função `set_updated_at()` e triggers em cada tabela. Esses podem ser criados na Fase 6 (centralizados) ou em cada fase. Para manter a Fase 6 focada em índices e RLS, os triggers podem ser criados já em cada fase que adiciona tabelas. Alternativa: criar todos os triggers na Fase 6. A Fase 6 já planeja "triggers e função set_updated_at". Portanto, **não** criar triggers nesta fase — as tabelas terão coluna `updated_at` com DEFAULT NOW(), e o trigger será adicionado na Fase 6.

### Ação 2.3

Nesta fase: apenas CREATE TABLE. Sem triggers. A coluna `updated_at` terá `DEFAULT NOW()` mas não será atualizada automaticamente em UPDATE até a Fase 6. Como trade-off temporário, a aplicação pode usar GORM que atualiza o campo, ou a Fase 6 será implementada em sequência próxima.

### Observação 2.3

A ordem das fases coloca a Fase 6 por último. Entre Fase 2 e Fase 6, as tabelas existirão sem trigger — o GORM ao fazer Update() pode definir explicitamente UpdatedAt. Para desenvolvimento inicial, aceitável. O importante é que a estrutura da tabela esteja correta.

---

### Pensamento 4

Dependência da Fase 1: os ENUMs `tenant_type`, `tenant_status`, `product_status` devem existir no schema. A migration 000003 assume que 000002 foi executada. Ordem: 000002 (schema+enums) → 000003 (tenants+products).

### Ação 2.4

A migration 000003 deve qualificar os tipos: `wisa_crm_db.tenant_type`, etc. Ou, se o search_path da conexão do migrate incluir wisa_crm_db, os tipos são encontrados sem qualificação. Para robustez, qualificar: `wisa_crm_db.tenant_type`.

### Observação 2.4

Migrations autossuficientes com qualificação explícita evitam erros de ambiente.

---

### Pensamento 5

Rollback (down): DROP TABLE na ordem inversa das dependências. Products não referencia tenants. Tenants não referencia products. Ordem: DROP TABLE wisa_crm_db.products; DROP TABLE wisa_crm_db.tenants; (ou vice-versa, não há FK entre elas).

### Ação 2.5

Migration down: DROP TABLE IF EXISTS wisa_crm_db.products; DROP TABLE IF EXISTS wisa_crm_db.tenants;

### Observação 2.5

IF EXISTS garante idempotência do down.

---

### Pensamento 6

Segurança e guidelines:
- Nenhum dado sensível em tenants além de tax_id (dado empresarial, não credencial)
- Code guidelines: migrations versionadas, sem AutoMigrate
- ADR-008: modelagem row-level tenancy — tenants é a entidade raiz, não possui tenant_id (é o próprio tenant)

### Ação 2.6

Verificar que tenants não precisa de tenant_id — correto, pois representa o cliente raiz. Products também é global (catálogo de planos), não pertence a um tenant específico.

### Observação 2.6

Conformidade confirmada. Nenhuma alteração necessária.

---

### Decisão final Fase 2

**Implementar:**
1. Migration `000003_create_tenants_and_products.up.sql`:
   - CREATE TABLE wisa_crm_db.tenants com colunas conforme ADR-008
   - CREATE TABLE wisa_crm_db.products com colunas conforme ADR-008
2. Migration `000003_create_tenants_and_products.down.sql`:
   - DROP TABLE IF EXISTS wisa_crm_db.products;
   - DROP TABLE IF EXISTS wisa_crm_db.tenants;

---

### Checklist de Implementação

1. [ ] Criar `000003_create_tenants_and_products.up.sql` com DDL completo de tenants e products
2. [ ] Incluir CONSTRAINT chk_tax_id_length em tenants
3. [ ] Incluir UNIQUE em tenants(slug) e products(slug)
4. [ ] Criar `000003_create_tenants_and_products.down.sql`
5. [ ] Testar migrate up e migrate down
6. [ ] Validar que não há referências circulares

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-008 tenants | id, slug, name, tax_id, type, status, created_at, updated_at, CHECK tax_id |
| ADR-008 products | id, slug, name, status, created_at, updated_at |
| ADR-003 | Tabelas no schema, tipos apropriados |
| Code Guidelines | Migration versionada, sem AutoMigrate |

---

## Referências

- [docs/adrs/ADR-008-arquitetura-multi-tenant.md](../../../adrs/ADR-008-arquitetura-multi-tenant.md)
- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
