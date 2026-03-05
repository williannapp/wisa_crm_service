# Fase 2 — Catálogo de Códigos de Erro e Mapeamento Domain→AppError

## Objetivo

Expandir o catálogo de códigos de erro padronizados no package `pkg/errors` e definir a lógica de mapeamento de erros de domínio para `AppError`, garantindo que mensagens sensíveis nunca vazem para o cliente.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-005 e code guidelines definem erros de domínio em `domain/errors.go`:
- `ErrInvalidCredentials`
- `ErrTenantNotFound`
- `ErrSubscriptionExpired`
- `ErrAccountLocked`
- `ErrUserNotFound` (usado internamente, mas NUNCA exposto — sempre mapeado para ErrInvalidCredentials na autenticação)

O ADR-010 e tasks_planning.md definem códigos e mensagens para cenários específicos:
- `INVALID_CREDENTIALS` — credenciais inválidas (401)
- `SUBSCRIPTION_SUSPENDED` — assinatura suspensa (provavelmente 403)
- `SUBSCRIPTION_CANCELED` — assinatura cancelada
- `RATE_LIMIT_EXCEEDED` — muitas tentativas (429)
- `TENANT_NOT_FOUND` — tenant não encontrado (404)
- `ACCOUNT_LOCKED` — conta bloqueada (403)

### Ação 1.1

Definir o catálogo completo em `pkg/errors/codes.go`:
- Constantes para cada código (INVALID_CREDENTIALS, SUBSCRIPTION_SUSPENDED, etc.)
- Constantes ou struct de configuração para mensagem e detalhe padrão de cada código
- Mapeamento código → HTTP status

Estrutura proposta: funções helper ou mapa de `ErrorSpec`:
```go
type ErrorSpec struct {
    Code       string
    Message    string
    Details    string
    HTTPStatus int
}
// Ou funções NewInvalidCredentials(), NewSubscriptionSuspended(), etc.
```

### Observação 1.1

Funções como `NewInvalidCredentials() *AppError` encapsulam mensagens padrão e evitam repetição. O catálogo centraliza todas as mensagens expostas ao cliente em um único lugar — facilita auditoria de segurança e i18n futura.

---

### Pensamento 2

O mapeamento domain→AppError deve ocorrer na **camada delivery**, pois:
- A delivery conhece tanto domain quanto pkg/errors
- Domain não conhece HTTP nem AppError (Dependency Rule)
- O ErrorMapper é um adaptador: converte conceitos de domínio em DTO de saída HTTP

O ErrorMapper terá assinatura:
```go
func MapToAppError(err error) *AppError
```
Retorna `*AppError` correspondente ao domínio, ou `AppError` genérico (500 Internal Server Error) para erros desconhecidos.

### Ação 1.2

Definir localização do ErrorMapper:
- **Caminho:** `internal/delivery/http/errors/mapper.go` ou `internal/delivery/http/mapper/error_mapper.go`
- O package `delivery/http` ou submódulo `delivery/http/errors` importa `domain` e `pkg/errors`
- Função `MapToAppError(err error) *AppError` usando `errors.Is()` para cada domain error

### Observação 1.2

A delivery faz a ponte entre domínio e infraestrutura HTTP. O mapeamento aqui respeita Clean Architecture: domain → (mapper na delivery) → AppError → JSON response.

---

### Pensamento 3

Regras de segurança para o mapeamento:
1. **Erros de autenticação:** ErrInvalidCredentials, ErrUserNotFound (do repo), ErrTenantNotFound → SEMPRE retornar o MESMO AppError genérico (INVALID_CREDENTIALS, 401). Nunca revelar "email não encontrado" vs "senha incorreta".
2. **Erros de assinatura:** ErrSubscriptionExpired, ErrSubscriptionSuspended → mensagens específicas conforme tasks_planning.md (SUBSCRIPTION_SUSPENDED, SUBSCRIPTION_CANCELED).
3. **Erros desconhecidos:** Qualquer outro `err` (GORM, bcrypt, etc.) → AppError genérico INTERNAL_ERROR (500) com mensagem genérica "Erro interno. Tente novamente." — NUNCA expor err.Error().
4. **Logging:** O mapper pode logar o erro original internamente (com err.Error()) para debug, mas nunca incluir no AppError retornado ao cliente.

### Ação 1.3

Regras de mapeamento documentadas:

| Domain Error | AppError Code | HTTP | Message |
|--------------|---------------|------|---------|
| ErrInvalidCredentials | INVALID_CREDENTIALS | 401 | Credenciais inválidas. |
| ErrUserNotFound (em auth) | INVALID_CREDENTIALS | 401 | Credenciais inválidas. |
| ErrTenantNotFound | INVALID_CREDENTIALS | 401 | Credenciais inválidas. |
| ErrAccountLocked | ACCOUNT_LOCKED | 403 | Conta bloqueada. |
| ErrSubscriptionSuspended | SUBSCRIPTION_SUSPENDED | 403 | Acesso suspenso por pendência financeira. |
| ErrSubscriptionCanceled | SUBSCRIPTION_CANCELED | 403 | Assinatura cancelada. |
| ErrTenantNotFound (em outros contextos) | TENANT_NOT_FOUND | 404 | Tenant não encontrado. |
| ErrRateLimitExceeded | RATE_LIMIT_EXCEEDED | 429 | Muitas tentativas. Tente novamente mais tarde. |
| Outros | INTERNAL_ERROR | 500 | Erro interno. Tente novamente. |

*Nota: ErrTenantNotFound no contexto de login deve mapear para INVALID_CREDENTIALS (ADR-010). Em outros contextos (ex: API admin), pode mapear para TENANT_NOT_FOUND.*

### Observação 1.3

A distinção de contexto (login vs admin) pode exigir que o mapper receba contexto opcional ou que o use case retorne um tipo mais específico. Para simplificar a Fase 2, adotar regra conservadora: em dúvida, usar INVALID_CREDENTIALS para falhas de auth. O mapper pode ser ampliado posteriormente com parâmetros de contexto.

---

### Pensamento 4

O `domain/errors.go` ainda não existe (previsto em feature 001 fase posterior). Esta fase deve incluir a criação de `internal/domain/errors.go` com os erros documentados, ou assumir que será criado em paralelo. O TRACKER da feature 001 indica `errors.go` como "Será criado em fase posterior". A Feature 004 pode criar o domain/errors.go como parte da Fase 2, pois é pré-requisito para o mapper.

### Ação 1.4

Criar `internal/domain/errors.go` com:
```go
var (
    ErrInvalidCredentials   = errors.New("invalid credentials")
    ErrTenantNotFound       = errors.New("tenant not found")
    ErrUserNotFound         = errors.New("user not found")
    ErrSubscriptionExpired  = errors.New("subscription expired")
    ErrSubscriptionSuspended = errors.New("subscription suspended")
    ErrSubscriptionCanceled = errors.New("subscription canceled")
    ErrAccountLocked        = errors.New("account locked")
    ErrRateLimitExceeded    = errors.New("rate limit exceeded")
)
```

O package `domain` não importa pkg/errors nem net/http. Apenas `errors` da stdlib.

### Observação 1.4

A criação de domain/errors.go nesta feature é coerente: o package de erro padronizado inclui o mapeamento domain→AppError, e o domínio precisa definir seus erros para que o mapper funcione. Alternativamente, domain/errors.go poderia ser uma task separada, mas isso atrasaria a Feature 004. Incluir na Fase 2.

---

### Pensamento 5

O mapper deve tratar erros wrapped. Em Go, `errors.Is(err, domain.ErrInvalidCredentials)` funciona com `fmt.Errorf("... %w", domain.ErrInvalidCredentials)`. A ordem de verificação no mapper deve priorizar erros mais específicos antes dos genéricos. Exemplo: se err wrap ErrAccountLocked e ErrInvalidCredentials de alguma forma (improvável), verificar AccountLocked primeiro.

A ordem recomendada: verificar erros específicos primeiro (AccountLocked, SubscriptionSuspended, etc.), depois InvalidCredentials/TenantNotFound/UserNotFound, por último fallback para INTERNAL_ERROR.

### Ação 1.5

Implementar MapToAppError com cadeia de `if errors.Is(err, domain.ErrX) { return ... }`. Ordem:
1. ErrAccountLocked
2. ErrSubscriptionSuspended, ErrSubscriptionCanceled, ErrSubscriptionExpired
3. ErrRateLimitExceeded
4. ErrInvalidCredentials, ErrUserNotFound, ErrTenantNotFound (todos → INVALID_CREDENTIALS em contexto auth; para fase 2, assumir contexto auth como default para Tenant/User)
5. default → INTERNAL_ERROR

### Observação 1.5

Para evitar ambiguidade, o use case de autenticação deve retornar somente ErrInvalidCredentials para qualquer falha de login (user not found, wrong password, tenant not found) — conforme ADR-010. Assim o mapper não precisa distinguir: ErrInvalidCredentials já é o único erro de auth. O domain/errors.go pode manter ErrUserNotFound e ErrTenantNotFound para uso interno (ex: repositórios), mas o use case de auth converte todos em ErrInvalidCredentials antes de retornar.

### Decisão 1.5

O use case retorna ErrInvalidCredentials para falhas de login. Os erros ErrUserNotFound e ErrTenantNotFound permanecem para outros use cases (ex: admin busca user). O mapper mapeia:
- ErrInvalidCredentials → INVALID_CREDENTIALS
- ErrTenantNotFound → TENANT_NOT_FOUND (em contextos não-auth) ou criar wrapper no use case
- ErrUserNotFound → USER_NOT_FOUND (ex: admin) ou não exposto em auth

Para Fase 2, o escopo mínimo: mapear os erros listados. O tratamento de contexto (auth vs admin) pode ser refinado quando os use cases forem implementados.

---

### Decisão final Fase 2

**Implementar:**

1. Expandir `backend/pkg/errors/codes.go` com todos os códigos: INVALID_CREDENTIALS, ACCOUNT_LOCKED, SUBSCRIPTION_SUSPENDED, SUBSCRIPTION_CANCELED, TENANT_NOT_FOUND, RATE_LIMIT_EXCEEDED, INTERNAL_ERROR
2. Criar funções ou specs para cada código com mensagem e detalhe padrão (conforme tasks_planning.md)
3. Criar `internal/domain/errors.go` com var para cada domain error
4. Criar `internal/delivery/http/errors/mapper.go` (ou similar) com MapToAppError(err error) *AppError
5. Documentar regras de segurança no mapper (nunca expor err.Error(); fallback 500 genérico)

---

### Checklist de Implementação

1. [ ] Expandir pkg/errors/codes.go com catálogo completo
2. [ ] Criar internal/domain/errors.go
3. [ ] Criar internal/delivery/http/errors/mapper.go
4. [ ] Implementar MapToAppError com errors.Is para cada domain error
5. [ ] Fallback para INTERNAL_ERROR com mensagem genérica
6. [ ] Verificar que nenhum err.Error() interno vaze para o cliente

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Mapear domain→AppError | MapToAppError na delivery |
| Impedir vazamento | Fallback 500 genérico; mensagens pré-definidas |
| ADR-010 auth | ErrInvalidCredentials → mensagem única |
| Code guidelines §6.3 | ErrorMapper centralizado na delivery |
| Clean Architecture | Domain sem imports de delivery ou pkg |

---

## Referências

- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../adrs/ADR-005-clean-architecture.md)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [docs/DON'T READ/planned-tasks/tasks_planning.md](../../DON'T%20READ/planned-tasks/tasks_planning.md)
