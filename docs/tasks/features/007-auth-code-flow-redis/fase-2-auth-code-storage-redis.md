# Fase 2 — Armazenamento de Authorization Code no Redis

## Objetivo

Definir a interface de armazenamento de authorization codes (AuthCodeStore), implementar o armazenamento em Redis com TTL de 40 segundos e estrutura de dados contendo as informações necessárias para gerar o JWT. Seguir Clean Architecture e Dependency Rule (ADR-005).

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O authorization code é um token opaco gerado após autenticação bem-sucedida. Ele deve ser armazenado no Redis com as informações necessárias para que o endpoint `/auth/token` possa gerar o JWT sem precisar revalidar credenciais. Ou seja, o valor armazenado deve incluir tudo que o JWTClaims precisa: Subject (user_id), Audience, TenantID, UserAccessProfile.

Alternativa: armazenar o JWT já gerado no Redis e retorná-lo na troca — evita regenerar, mas duplica lógica e aumenta tamanho do valor. A abordagem mais limpa: armazenar os **claims** (não o JWT) e gerar o JWT no momento da troca. Assim, a rotação de chaves e a expiração do JWT são aplicadas corretamente no momento do uso.

### Ação 1.1

Definir struct de dados para o valor armazenado no Redis:

```go
// AuthCodeData contém os dados necessários para emitir o JWT na troca do code.
type AuthCodeData struct {
    Subject           string // user_id (ULID)
    Audience          string
    TenantID          string
    UserAccessProfile string
}
```

### Observação 1.1

A interface deve viver no domínio ou em um pacote de serviço (port). O domínio não conhece Redis; a interface AuthCodeStore é uma porta que a infraestrutura implementa. A struct AuthCodeData pode estar junto da interface ou em um pacote compartilhado.

---

### Pensamento 2

A interface precisa de três operações:
1. **Store(ctx, code, data, ttl)** — armazena o code com TTL
2. **GetAndDelete(ctx, code)** — obtém os dados e remove o code (single-use). Operação atômica para evitar race condition.
3. Opcional: **Exists(ctx, code)** — apenas para verificação sem consumir; não necessário se GetAndDelete cobrir o caso.

O Redis oferece `GET` + `DEL` em sequência, mas não atomicamente. Para atomicidade, usar `GETDEL` (Redis 6.2+) ou Lua script. O `GETDEL` retorna o valor e remove a chave em uma única operação atômica.

### Ação 1.2

Definir interface em `internal/domain/service/auth_code_store.go` (ou `internal/domain/repository/auth_code_repository.go`):

```go
package service

import "context"

// AuthCodeStore armazena temporariamente os dados associados a um authorization code.
type AuthCodeStore interface {
    Store(ctx context.Context, code string, data *AuthCodeData, ttlSeconds int) error
    GetAndDelete(ctx context.Context, code string) (*AuthCodeData, error)
}
```

### Observação 1.2

O nome "Store" evita acoplamento com Redis. O TTL em segundos torna a interface flexível. O `GetAndDelete` retorna `nil, err` quando o code não existe ou expirou — o chamador interpreta como "code inválido ou já utilizado".

---

### Pensamento 3

O formato da chave no Redis: `auth_code:{code}`. O valor deve ser serializado — JSON é simples e legível. O go-redis aceita `Set(ctx, key, value, expiration)` onde value pode ser string, []byte ou qualquer tipo que implemente encoding.BinaryMarshaler. Para JSON: `json.Marshal(data)` e armazenar como string. Na leitura: `json.Unmarshal`.

### Ação 1.3

Implementação em `internal/infrastructure/cache/redis_auth_code_store.go`:

- Struct `RedisAuthCodeStore` com campo `client *redis.Client`
- Método `Store`: chave `auth_code:` + code, valor JSON de AuthCodeData, `redis.WithExpiration(time.Duration(ttlSeconds)*time.Second)`
- Método `GetAndDelete`: usar `client.GetDel(ctx, key)` (Redis 6.2+). Se erro `redis.Nil`, retornar `nil, ErrCodeNotFound` (erro de domínio). Deserializar JSON e retornar *AuthCodeData.

### Observação 1.3

Se a versão do Redis for anterior a 6.2, usar Lua script ou `GET` + `DEL` em pipeline. O Redis 7 (imagem alpine) suporta GETDEL. Verificar compatibilidade da imagem Docker.

---

### Pensamento 4

Erros de domínio: quando o code não existe, expirou ou já foi usado, o `GetAndDelete` retorna um erro semântico. Criar `domain.ErrCodeNotFound` ou `domain.ErrCodeExpiredOrInvalid` para que o Use Case e o handler possam mapear para HTTP 401. Para simplificar, um único erro `ErrCodeInvalidOrExpired` cobre ambos os casos — não revelar ao cliente se foi "não encontrado" ou "já usado" (segurança).

### Ação 1.4

Adicionar em `internal/domain/errors.go`:

```go
ErrCodeInvalidOrExpired = errors.New("authorization code invalid or expired")
```

### Observação 1.4

Alinhado com code guidelines §6.1 (erros de domínio) e ADR-010 (não diferenciar detalhes que possam ajudar enumeração).

---

### Pensamento 5

O AuthCodeData precisa ser definido em um local que tanto o domain/service quanto a infraestrutura possam usar. Se AuthCodeStore está em `domain/service`, o AuthCodeData pode estar no mesmo pacote. A implementação Redis importa `domain/service` para a interface e o tipo — dependency rule respeitada (infraestrutura depende do domínio).

### Ação 1.5

Manter `AuthCodeData` em `domain/service` junto da interface `AuthCodeStore`. A implementação Redis em `infrastructure/cache` importa o pacote de serviço.

### Observação 1.5

Alternativa: colocar AuthCodeData em `domain/entity` se for considerado entidade. Para dados de transferência entre Store e JWT generation, `service` é adequado.

---

### Pensamento 6

Injeção no main: o `RedisAuthCodeStore` recebe o `*redis.Client` criado na Fase 1. O Use Case de login (Fase 3) e o Use Case de troca de code (Fase 4) dependerão do AuthCodeStore via interface. O main.go construirá `authCodeStore := cache.NewRedisAuthCodeStore(redisClient)` e injetará nos use cases.

### Ação 1.6

Construtor:

```go
func NewRedisAuthCodeStore(client *redis.Client) *RedisAuthCodeStore {
    return &RedisAuthCodeStore{client: client}
}
```

Garantir que `*RedisAuthCodeStore` implemente a interface `service.AuthCodeStore` (verificação em tempo de compilação: `var _ service.AuthCodeStore = (*RedisAuthCodeStore)(nil)`).

### Observação 1.6

Go permite verificação de implementação de interface implicitamente. O compilador garantirá que os métodos tenham as assinaturas corretas.

---

### Pensamento 7 (Segurança)

- **TTL 40 segundos:** Conforme especificação do usuário. Janela curta reduz risco de code interceptado.
- **Chave opaca:** O code deve ser gerado com `crypto/rand` (32 bytes em hex = 64 caracteres) para ser imprevisível.
- **Single-use:** GetAndDelete garante remoção atômica. Reutilização do mesmo code retorna "não encontrado".
- **Sem dados sensíveis:** AuthCodeData não contém senha nem tokens — apenas identificadores para emissão do JWT.

### Observação 1.7

O code em si não deve ser sequencial ou derivado de dados previsíveis. Será gerado na Fase 3 com `crypto/rand`.

---

## Checklist de Implementação

- [ ] 1. Criar `AuthCodeData` e interface `AuthCodeStore` em `domain/service`
- [ ] 2. Adicionar `ErrCodeInvalidOrExpired` em `domain/errors.go`
- [ ] 3. Implementar `RedisAuthCodeStore` em `infrastructure/cache/`
- [ ] 4. Usar chave `auth_code:{code}` e TTL configurável (40s para login)
- [ ] 5. `GetAndDelete` com `GetDel` (ou alternativa para Redis < 6.2)
- [ ] 6. Serializar/deserializar AuthCodeData em JSON
- [ ] 7. Injetar AuthCodeStore no main (wiring para Fases 3 e 4)

---

## Conclusão

A Fase 2 estabelece o contrato e a implementação de armazenamento de authorization codes. A interface segue Dependency Inversion; a implementação Redis é um detalhe isolado. O TTL de 40 segundos atende à especificação e reduz a janela de ataque.
