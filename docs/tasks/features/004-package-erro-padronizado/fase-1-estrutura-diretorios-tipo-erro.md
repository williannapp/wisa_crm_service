# Fase 1 — Estrutura de Diretórios e Tipo Estruturado AppError

## Objetivo

Criar a estrutura de diretórios do package de erro padronizado e definir o tipo estruturado de erro com código interno, mensagem amigável, detalhe e status HTTP correspondente, em conformidade com Clean Architecture e code guidelines.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O projeto adota Clean Architecture (ADR-005) e code guidelines definem `pkg/` para utilitários compartilhados (logger, etc.). O package de erros será um utilitário consumido pela camada delivery e possivelmente por middlewares. O `pkg/errors` não deve importar `internal/domain`, `net/http` ou frameworks — apenas tipos padrão. A Dependency Rule exige que dependências apontem para dentro; `pkg` é raiz e não depende de camadas internas.

### Ação 1.1

Definir a localização do package:
- **Caminho:** `backend/pkg/errors/`
- **Motivo:** Consistente com `pkg/logger/`; utilitário compartilhado; não viola Dependency Rule
- **Alternativa descartada:** `internal/errors/` — menos explícito que pkg para código compartilhado entre delivery e infrastructure

### Observação 1.1

A escolha de `pkg/errors` segue o layout recomendado em code_guidelines/backend.md. O package será importado por `internal/delivery/http` e `internal/infrastructure/http/middleware`. Nenhuma violação de dependência.

---

### Pensamento 2

O tipo estruturado de erro deve conter:
1. **Código interno padronizado** (ex: `INVALID_CREDENTIALS`) — string constante para identificação programática e logs
2. **Mensagem amigável ao cliente** — texto curto e seguro para exposição na API
3. **Detalhe do erro** — descrição adicional (opcional); jamais stack traces ou dados internos
4. **Status HTTP correspondente** — 400, 401, 404, 429, 500 etc.

O formato de resposta JSON esperado (tasks_planning.md, ADR-010) é:
```json
{
  "status": "error",
  "error_code": "INVALID_CREDENTIALS",
  "message": "Credenciais inválidas.",
  "details": "Verifique usuário e senha."
}
```

O campo `status` pode ser inferido (sempre "error" em respostas de erro) ou explícito. O tipo AppError deve ser serializável para JSON.

### Ação 1.2

Definir a struct `AppError`:
```go
type AppError struct {
    Code       string // Código padronizado (ex: INVALID_CREDENTIALS)
    Message    string // Mensagem amigável ao cliente
    Details    string // Detalhe opcional; vazio se não houver
    HTTPStatus int    // Status HTTP (401, 404, etc.)
}
```

- `Code`: string, SCREAMING_SNAKE_CASE, exportada
- `Message`: string, nunca vazia em erros expostos ao cliente
- `Details`: string, opcional (pode ser vazia)
- `HTTPStatus`: int, valores válidos 4xx e 5xx

Para JSON com chaves em snake_case: usar tags `json:"error_code"`, `json:"message"`, `json:"details"`. O campo `status` pode ser adicionado na struct de resposta (Status string `json:"status"`) ou incluído no MarshalJSON customizado.

### Observação 1.2

A struct atende aos requisitos. O campo `Details` deve ser controlado para evitar vazamento — apenas detalhes aprovados e pré-definidos, nunca `err.Error()` de erros internos. O construtor `NewAppError` deve garantir que Message nunca seja vazia e que HTTPStatus seja válido.

---

### Pensamento 3

O tipo `AppError` deve implementar a interface `error` do Go para compatibilidade com código que trata erros. Método `Error() string` retornando a mensagem (ou código) para logs internos. **Importante:** O método `Error()` é para logs e debugging; a serialização JSON usa apenas `Code`, `Message` e `Details` — nunca o resultado de `Error()` em produção para o cliente.

### Ação 1.3

- Implementar `func (e *AppError) Error() string` retornando algo como `"[CODE] Message"` para logs
- Implementar `MarshalJSON` ou struct com tags JSON para resposta HTTP padronizada
- O `Details` no JSON deve ser omitido quando vazio (`omitempty`) para não poluir resposta

### Observação 1.3

Implementar `error` interface permite `return appErr` em use cases quando desejado; porém, na Clean Architecture, use cases retornam erros de domínio e o ErrorMapper na delivery faz a conversão. O AppError é primariamente um DTO de saída HTTP. A interface `error` é opcional mas útil para consistência.

---

### Pensamento 4

Estrutura de arquivos no package. Separar responsabilidades:
- `app_error.go`: struct AppError, NewAppError, implementação error, MarshalJSON
- `codes.go`: constantes de códigos (ErrorCode ou strings) e talvez constantes de HTTP status

A Fase 1 foca em estrutura e tipo; o catálogo completo de códigos pode ser expandido na Fase 2. Para Fase 1, criar pelo menos um código de exemplo (ex: `ErrCodeInvalidCredentials`) para validar o design.

### Ação 1.4

Arquivos a criar:
1. `backend/pkg/errors/app_error.go` — struct, construtor, Error(), MarshalJSON
2. `backend/pkg/errors/codes.go` — constantes de códigos e status HTTP (mínimo para validação)

### Observação 1.4

A separação mantém app_error.go focado no tipo e codes.go no catálogo. Conformidade com Single Responsibility (code guidelines §11).

---

### Pensamento 5

Segurança: impedir vazamento de informações internas. O construtor `NewAppError` deve:
- Aceitar apenas `Code`, `Message`, `Details` e `HTTPStatus` como parâmetros explícitos
- Nunca aceitar `err error` interno e expor `err.Error()` em Message ou Details
- O Details deve ser controlado — idealmente um enum ou conjunto fechado de mensagens pré-definidas

Para Fase 1, o contrato do construtor já bloqueia vazamento: quem chama deve passar mensagens explícitas e seguras.

### Ação 1.5

Contrato do construtor:
```go
func NewAppError(code, message string, details string, httpStatus int) *AppError
```
- Validar: `message` não vazio; `httpStatus` entre 400 e 599
- Se `details` vazio, usar string vazia (será omitempty no JSON)

### Observação 1.5

O design impede que erros de infraestrutura (GORM, bcrypt, etc.) sejam encadeados diretamente na resposta. O ErrorMapper (Fase 3) será o único ponto que converte domain errors em AppError, usando mensagens pré-definidas.

---

### Pensamento 6

Conformidade com ADRs e guidelines:
- ADR-005: Domain errors vivem em `domain/errors.go`; AppError vive em pkg e não conhece domain
- ADR-010: "Sempre retornar mensagem genérica para falha de autenticação" — o AppError permite Message genérica; a responsabilidade de usar mensagens corretas é do ErrorMapper
- Code guidelines §6.3: "Centralizar mapeamento em ErrorMapper na camada delivery" — Fase 3
- Code guidelines §12 (Credenciais): "Nunca diferenciar usuário não encontrado vs senha incorreta" — o ErrorMapper garantirá AppError único para ambos

### Ação 1.6

Checklist de conformidade Fase 1:
- [ ] pkg/errors não importa domain, net/http, gin
- [ ] AppError serializável para JSON com campos: status, error_code, message, details (omitempty)
- [ ] Construtor com validação básica
- [ ] Implementar interface error para consistência

### Observação 1.6

A Fase 1 é autocontida e não introduce dependências circulares ou violações de arquitetura.

---

### Decisão final Fase 1

**Implementar:**

1. Criar diretório `backend/pkg/errors/`
2. Criar `app_error.go`:
   - Struct `AppError` com: `Code`, `Message`, `Details`, `HTTPStatus`
   - Construtor `NewAppError(code, message, details string, httpStatus int) *AppError`
   - Método `Error() string` para interface `error`
   - Método `MarshalJSON` ou struct com tags JSON para `{"status":"error","error_code":"...","message":"...","details":"..."}` (details com omitempty)
3. Criar `codes.go` com constantes iniciais: pelo menos `InvalidCredentials` (401) como exemplo

---

### Checklist de Implementação

1. [ ] Criar `backend/pkg/errors/app_error.go`
2. [ ] Definir struct AppError com tags JSON
3. [ ] Implementar NewAppError com validação
4. [ ] Implementar Error() string
5. [ ] Criar `backend/pkg/errors/codes.go` com código de exemplo
6. [ ] Garantir que pkg/errors não tenha imports de internal ou gin

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| Estrutura de diretórios | `backend/pkg/errors/` |
| Código interno padronizado | Campo `Code` (ex: INVALID_CREDENTIALS) |
| Mensagem amigável | Campo `Message` |
| Detalhe do erro | Campo `Details` (opcional) |
| Status HTTP | Campo `HTTPStatus` |
| Clean Architecture | pkg sem dependência de internal |
| Evitar vazamento | Construtor exige parâmetros explícitos |

---

## Referências

- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [docs/adrs/ADR-005-clean-architecture.md](../../adrs/ADR-005-clean-architecture.md)
- [docs/adrs/ADR-010-fluxo-centralizado-de-autenticacao.md](../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
