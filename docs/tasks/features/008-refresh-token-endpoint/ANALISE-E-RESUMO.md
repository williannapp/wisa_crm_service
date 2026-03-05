# Feature 008 — Análise e Resumo do Planejamento

Este documento consolida a análise e o raciocínio ReAct aplicado ao planejamento completo da feature Refresh Token Endpoint.

---

## Item 1: Cliente deve chamar o software de autenticação quando o token expirar

**Pensamento:** O backend do cliente (sistema que consome o wisa-crm-service) deve interceptar respostas 401 em requisições à API protegida e, nesse momento, realizar uma chamada para o wisa-crm-service.

**Ação:** Documentar no guia de integração que o cliente deve implementar um interceptor (ou middleware) que:
1. Detecta 401 em resposta à API
2. Chama `POST /api/v1/auth/refresh` com refresh_token, tenant_slug e product_slug
3. Se 200: atualiza tokens e retenta a requisição original
4. Se 401 ou 402: redireciona para login

**Observação:** Esta é responsabilidade do **cliente**, não do wisa-crm-service. O wisa-crm-service fornece o endpoint; a documentação (Fase 4) orienta a implementação.

---

## Item 2: Software de autenticação — Receber e Validar

**Pensamento:** O endpoint POST /auth/refresh deve receber refresh_token, product_slug e tenant_slug. A validação exige que a hash exista no banco com base nesses três identificadores.

**Ação:** O endpoint valida:
1. Resolve tenant_slug → tenant_id (TenantRepository.FindBySlug)
2. Resolve product_slug → product_id (ProductRepository.FindBySlug)
3. Computa SHA-256 do refresh_token
4. Busca em refresh_tokens WHERE token_hash = hash AND tenant_id = X AND product_id = Y
5. Verifica: revoked_at IS NULL e expires_at > NOW()

**Observação:** A tabela atual tem tenant_id mas **não** product_id. Daí a necessidade da Fase 1 (migration).

---

## Item 3: Caso inválido → 401 Unauthorized com erro genérico

**Pensamento:** Conforme requisito e code guidelines (credenciais: nunca diferenciar), qualquer falha de validação deve retornar 401 com mensagem genérica.

**Ação:** Criar ErrRefreshTokenInvalid. Mapear para HTTP 401 com mensagem "Não autorizado" ou similar. Nunca revelar se o token é inexistente, revogado ou expirado.

**Observação:** Alinhado com OWASP e ADR-010.

---

## Item 4: Caso válido — Revogar, Gerar, Inserir

**Pensamento:** Em transação atômica:
1. Revogar token atual (revoked_at = NOW())
2. Gerar novo Access Token (15 min, não armazenado)
3. Gerar novo Refresh Token (SHA-256 hash, inserir no banco)
4. Retornar ambos ao cliente

**Ação:** Método Rotate no RefreshTokenRepository ou uso de transação com RevokeByID + Create. SELECT FOR UPDATE para evitar race condition.

**Observação:** Conformidade com ADR-003 (ACID), ADR-006 (rotativo, hash, 7 dias).

---

## Item 5: Access Token não armazenado em banco

**Pensamento:** O Access Token é stateless (JWT). Não persiste em banco. Apenas o Refresh Token é persistido como hash.

**Ação:** Nenhuma alteração de modelo. O fluxo já segue esse padrão.

**Observação:** Alinhado com ADR-006.

---

## Item 6: Refresh Token — Rotativo, SHA-256, 7 dias

**Pensamento:** Cada uso invalida o anterior. Armazenar como hash SHA-256. Expiração máxima 7 dias.

**Ação:** 
- Gerar token com crypto/rand (32 bytes)
- Hash com SHA-256, armazenar hex (64 chars) em token_hash
- expires_at = NOW() + 7 dias
- Ao renovar: revogar antigo, inserir novo

**Observação:** Implementação na Fase 2 e Fase 3.

---

## Análise: Alteração da tabela refresh_tokens

### Pergunta
Será necessário mudar a tabela de refresh_tokens para considerar product_slug e tenant_slug? Atualmente está composta por tenant_id somente.

### Resposta: **SIM**

**Justificativa:**
1. A validação no refresh usa refresh_token + product_slug + tenant_slug. Os slugs são resolvidos para IDs; a busca no banco usa tenant_id e **product_id**.
2. A tabela atual tem apenas tenant_id. Não possui product_id.
3. Um usuário pode ter acesso a múltiplos produtos no mesmo tenant. O refresh token deve ser escopado por (user, tenant, product) para que um token emitido no contexto do produto A não seja usado para renovar sessão do produto B.
4. O requisito exige validação por product_slug; portanto, product_id é obrigatório na tabela.

**Alteração necessária:** Adicionar coluna `product_id UUID NOT NULL REFERENCES products(id)` à tabela refresh_tokens (Fase 1).

---

## Conformidade com Guidelines e ADRs

| Documento | Conformidade |
|-----------|--------------|
| docs/code_guidelines/backend.md | Clean Architecture, repositórios, migrations versionadas, erros de domínio, SQL parametrizado |
| ADR-003 (PostgreSQL) | Transações ACID, índices, RLS |
| ADR-006 (JWT) | Access 15 min, Refresh 7 dias, hash SHA-256 |
| ADR-008 (Multi-tenant) | product_id para escopo (tenant, product) |
| ADR-010 (Fluxo Auth) | Refresh token rotation, verificação assinatura (402), rate limiting |

---

## Estrutura de Fases Criada

| Fase | Documento | Objetivo |
|------|-----------|----------|
| 1 | fase-1-migration-product-id-refresh-tokens.md | Migration 000008: product_id + índice |
| 2 | fase-2-refresh-token-no-token-exchange.md | AuthCodeData.ProductID, gerar/persistir refresh no /token |
| 3 | fase-3-endpoint-post-auth-refresh.md | POST /auth/refresh, validação, rotação, verificação assinatura |
| 4 | fase-4-documentacao-integracao-refresh.md | Guia para clientes integrarem o refresh |

---

## Próximos Passos

1. Implementar Fase 1 — migration
2. Implementar Fase 2 — refresh no token exchange
3. Implementar Fase 3 — endpoint /refresh
4. Implementar Fase 4 — documentação

Nenhum código foi implementado neste planejamento; apenas documentação e estrutura de tasks.
