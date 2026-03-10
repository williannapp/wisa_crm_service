# Fase 3 — Validação JWT e Rota Protegida

## Objetivo

Implementar a validação do JWT no backend da test-app conforme docs/integration/auth-code-flow-integration.md: obter a chave pública via JWKS, validar assinatura RS256, iss, aud, exp, nbf, kid, e expor uma rota protegida GET /api/hello que retorna "Hello World" somente quando o JWT é válido.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

A aplicação cliente deve validar o JWT antes de conceder acesso (doc § "Validação do JWT no Cliente"). A validação obrigatória inclui:
1. Assinatura RS256 usando chave pública do auth
2. Verificar iss (issuer)
3. Verificar aud (audiência)
4. Verificar exp (expiração)
5. Verificar nbf (not before), se presente
6. Usar kid do header para selecionar a chave correta no JWKS

A chave pública é obtida via GET {AUTH_SERVER_URL}/.well-known/jwks.json.

### Ação 1.1

Implementar um JWKS fetcher/cache no backend da test-app:
- GET https://auth.wisa.labs.com.br/.well-known/jwks.json
- O endpoint retorna Cache-Control: public, max-age=86400 (24h)
- Cachear a resposta em memória com TTL de 24h (ou respeitando max-age)
- Estrutura: array de keys com kty, use, alg, kid, n, e

### Observação 1.1

Conforme doc § "Obtenção da Chave Pública (JWKS)". A biblioteca JWT em Go (ex.: github.com/golang-jwt/jwt) aceita chave pública em formato RSA. É necessário converter o JWK (n, e em base64url) para rsa.PublicKey.

---

### Pensamento 2

O JWT no cookie é enviado pelo navegador nas requisições ao mesmo domínio. O backend precisa:
1. Ler o cookie `wisa_access_token` (ou equivalente)
2. Fazer parse do JWT
3. Extrair o kid do header
4. Buscar a chave correspondente no JWKS
5. Validar assinatura com a chave
6. Validar claims: iss, aud, exp, nbf
7. Se válido, permitir acesso; se inválido/expirado, retornar 401

### Ação 1.2

Criar um middleware ou função de validação que:
- Extrai o token do cookie (ou do header Authorization: Bearer, se o frontend enviar)
- Para API chamada pelo frontend via fetch/XHR, o cookie é enviado automaticamente se SameSite e domínio permitirem
- O frontend da test-app chama o backend no mesmo domínio (ou via proxy) — cookies serão enviados
- Validar com jwt.ParseWithClaims, método RS256 explícito (evitar algorithm confusion — ADR-006)

### Observação 1.2

ADR-006: "O backend Go deve rejeitar qualquer token cujo header alg não seja RS256". Usar jwt.WithValidMethods([]string{"RS256"}).

---

### Pensamento 3

Valores esperados para as claims:
- **iss:** Deve corresponder ao auth configurado. O JWT do wisa-crm-service usa iss configurável (ex.: "https://auth.wisa.labs.com.br" ou "wisa-crm-service"). Verificar qual valor o auth emite e configurar a validação para aceitar esse issuer.
- **aud:** Deve corresponder ao domínio do cliente. Para a test-app em localhost, o auth pode emitir aud como "localhost:8081" ou similar. A test-app deve validar que aud corresponde ao seu próprio domínio (APP_URL ou derivado).
- **exp, nbf:** Validar expiração; nbf com leeway de 30s para clock skew (ADR-006).

### Ação 1.3

Configurar issuer e audience como variáveis de ambiente ou derivadas:
- JWT_ISSUER: valor esperado no claim iss (ex.: "https://auth.wisa.labs.com.br")
- JWT_AUDIENCE: valor esperado no claim aud (ex.: "localhost:8081" para dev, ou domínio de produção da test-app)

O backend do wisa-crm-service emite aud baseado em tenant_slug e domínio. A test-app deve usar o valor que o auth emite para o tenant/produto de teste.

### Observação 1.3

Pode ser necessário consultar o código do backend principal para entender o formato exato de iss e aud nos JWTs emitidos.

---

### Pensamento 4

Rota protegida GET /api/hello:
- Requer JWT válido no cookie
- Se não houver cookie ou token inválido → 401 Unauthorized
- Se válido → 200 OK com body `{"message": "Hello World"}` ou similar

### Ação 1.4

Implementar handler GET /api/hello que:
1. Chama o middleware de validação JWT
2. Se válido, retorna JSON {"message": "Hello World"}
3. Se inválido, retorna 401 com corpo padronizado (opcional: {"error": "unauthorized"})

### Observação 1.4

Esta rota será consumida pelo frontend na Fase 4 para exibir "Hello World" após autenticação.

---

### Pensamento 5

Organização do código seguindo Clean Architecture:
- domain: interfaces ou tipos para validação (opcional, pode ser simples)
- infrastructure: JWKS fetcher, conversão JWK → RSA public key
- delivery: middleware JWT, handler /api/hello

O pacote github.com/golang-jwt/jwt/v5 oferece jwt.ParseWithClaims e jwt.Keyfunc para validação com múltiplas chaves.

### Ação 1.5

Estrutura sugerida:
- `internal/infrastructure/jwt/jwks_provider.go` — busca e cache do JWKS
- `internal/infrastructure/jwt/validator.go` — valida JWT usando JWKS
- `internal/delivery/http/middleware/jwt_auth.go` — middleware que valida cookie e injeta claims no context
- `internal/delivery/http/handler/hello_handler.go` — GET /api/hello

### Observação 1.5

Conformidade com code_guidelines/backend.md e ADR-005. Manter domain livre de dependências de infraestrutura.

---

## Decisão Final Fase 3

**Entregáveis:**

1. **JWKS Provider:**
   - Fetch GET {AUTH_SERVER_URL}/.well-known/jwks.json
   - Cache em memória (TTL 24h)
   - Função para obter rsa.PublicKey por kid

2. **JWT Validator:**
   - Parse JWT com alg RS256 explícito
   - Validação de iss, aud, exp, nbf
   - Leeway de 30s para exp/nbf (clock skew)

3. **Middleware JWT Auth:**
   - Extrai token do cookie wisa_access_token
   - Valida com JWKS + claims
   - Em sucesso: armazena claims no context, chama próximo handler
   - Em falha: 401 Unauthorized

4. **Rota GET /api/hello:**
   - Protegida pelo middleware JWT
   - Retorna 200 {"message": "Hello World"} quando autenticado
   - Retorna 401 quando não autenticado

5. **Variáveis de ambiente:**
   - JWT_ISSUER (opcional, default baseado em AUTH_SERVER_URL)
   - JWT_AUDIENCE (opcional, default baseado em APP_URL)

---

## Checklist de Implementação

1. [ ] Implementar JWKS fetcher com cache
2. [ ] Converter JWK para rsa.PublicKey
3. [ ] Implementar JWT validator (RS256, iss, aud, exp, nbf)
4. [ ] Implementar middleware JWT auth
5. [ ] Implementar GET /api/hello
6. [ ] Garantir que rota /login e /callback não exijam JWT
7. [ ] Testar fluxo: login → callback → GET /api/hello com cookie

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| auth-code-flow-integration.md §Validação JWT | 1-6 |
| ADR-006 | RS256 explícito, nunca alg:none |
| ADR-006 | Leeway 30s para clock skew |
| code_guidelines backend | Clean Architecture |

---

## Referências

- [docs/integration/auth-code-flow-integration.md](../../../integration/auth-code-flow-integration.md) § "Validação do JWT no Cliente", § "Obtenção da Chave Pública (JWKS)"
- [docs/adrs/ADR-006-jwt-com-assinatura-assimetrica.md](../../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
