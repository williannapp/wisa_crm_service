# Fase 2 — Variáveis de Ambiente para Banco de Dados

## Objetivo

Criar as variáveis de ambiente necessárias para conexão com o banco de dados. O `.env.example` da Fase 4 da Feature 001 já possui um placeholder comentado para `DATABASE_URL`. Esta fase formaliza e documenta as variáveis de banco.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O formato de connection string PostgreSQL segue o padrão:
`postgres://user:password@host:port/dbname?sslmode=...`

Conforme ADR-003, em produção deve-se usar SSL (`sslmode=require` ou `verify-full`). Em desenvolvimento local com Docker, `sslmode=disable` é aceitável.

### Ação 2.1

Definir as variáveis de ambiente para banco:

| Variável | Exemplo | Obrigatória | Descrição |
|----------|---------|-------------|-----------|
| `DATABASE_URL` | `postgres://user:pass@localhost:5432/wisa_crm_db?sslmode=disable` | Em production: sim | Connection string completa para PostgreSQL |

Alternativa: variáveis separadas (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`) — mais verboso, mas permite maior flexibilidade em deploy. O padrão amplamente adotado é `DATABASE_URL` única (usada por Heroku, Railway, Docker, etc.).

### Observação 2.1

`DATABASE_URL` única é mais portável. Manter apenas esta variável. Se no futuro for necessário suportar variáveis separadas (ex: secrets em Kubernetes), pode-se criar um pacote de config que monte a URL a partir delas.

---

### Pensamento 2

O `.env.example` atual (Feature 001, Fase 4) já tem:
```
# Database (para fases futuras - opcional agora)
# DATABASE_URL=postgres://user:pass@localhost:5432/wisa_crm_db?sslmode=disable
```

Para a Fase 2, deve-se **descomentar e documentar** a variável, tornando-a oficial. O valor de exemplo deve refletir o que será usado com o docker-compose (a ser criado na Fase 3).

### Ação 2.2

Atualizar `backend/.env.example`:

```
# Server
PORT=8080
APP_ENV=development

# Database
# Formato: postgres://user:password@host:port/database?sslmode=disable
# Em produção use sslmode=require ou verify-full (ADR-003)
DATABASE_URL=postgres://wisa_crm:wisa_crm_secret@localhost:5432/wisa_crm_db?sslmode=disable
```

Os valores `wisa_crm` e `wisa_crm_secret` devem coincidir com o que será configurado no docker-compose (Fase 3) para evitar confusão.

### Observação 2.2

O `.env` está no .gitignore. O `.env.example` NUNCA deve conter senhas reais de produção — apenas exemplos seguros para desenvolvimento local.

---

### Pensamento 3

Conforme ADR-009, em produção as variáveis vêm de `EnvironmentFile=/etc/wisa-crm/env`. O arquivo `/etc/wisa-crm/env` no servidor contém `DATABASE_URL` com credenciais reais. O `.env.example` serve apenas como template para desenvolvimento.

### Ação 2.3

Garantir que a documentação (Fase 4) explique:
- Desenvolvimento: copiar `.env.example` para `.env` e ajustar se necessário
- Produção: configurar `DATABASE_URL` no EnvironmentFile do systemd (já previsto em ADR-009)

### Observação 2.3

Não criar variáveis adicionais desnecessárias (DB_POOL_SIZE, etc.) — YAGNI. O pool é configurado no código conforme guidelines.

---

### Pensamento 4

Variáveis opcionais relacionadas a migrations? O `golang-migrate` usa a mesma `DATABASE_URL` para rodar migrations. Não é necessário variável adicional.

### Ação 2.4

Manter apenas `DATABASE_URL`. O comando de migration receberá a mesma URL do ambiente.

### Observação 2.4

Consistência: uma única fonte de verdade para conexão.

---

### Decisão final Fase 2

**Implementar:**
1. Atualizar `backend/.env.example` com `DATABASE_URL` descomentada e documentada
2. Usar valores de exemplo compatíveis com o docker-compose (Fase 3): `wisa_crm`, `wisa_crm_secret`, `wisa_crm_db`
3. Incluir comentário sobre `sslmode` para produção
4. Garantir que `backend/.env` permaneça no .gitignore

---

### Checklist de Implementação

1. [ ] Atualizar `backend/.env.example` com `DATABASE_URL` oficial e documentação
2. [ ] Valores de exemplo alinhados com credenciais do docker-compose (Fase 3)
3. [ ] Comentário sobre sslmode para produção (ADR-003)
4. [ ] Verificar que `.env` está no .gitignore

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-003 | Formato de conexão PostgreSQL; sslmode documentado |
| ADR-009 | Variáveis sensíveis não em .env.example (apenas exemplos) |
| Segurança | Nenhum secret real no template |
| Code Guidelines | Simplicidade mantida |

---

## Dependência

Esta fase pode ser executada em paralelo ou antes da Fase 3 (Containers). Os valores de exemplo devem ser consistentes com o que será definido no `docker-compose.yml`.

---

## Referências

- [docs/adrs/ADR-003-postgresql-como-banco-de-dados.md](../../../adrs/ADR-003-postgresql-como-banco-de-dados.md)
- [docs/adrs/ADR-009-infraestrutura-vps-linux.md](../../../adrs/ADR-009-infraestrutura-vps-linux.md)
- [Feature 001 - Fase 4 - Variáveis de ambiente](../001-estrutura-inicial-backend/fase-4-variaveis-ambiente.md)
