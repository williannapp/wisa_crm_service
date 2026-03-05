# Fase 3 — Serviços de Infraestrutura (Password e JWT)

## Objetivo

Implementar os serviços de criptografia necessários para o login: comparação de senha (bcrypt) e assinatura de JWT (RS256), em conformidade com ADR-006.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-006 define:
- JWT assinado com **RS256** (RSA + SHA-256)
- Chave privada RSA **4096 bits** em ambiente controlado
- Claims: iss, sub, aud, tenant_id, user_access_profile, jti, iat, exp, nbf
- Access token com duração de **15 minutos**
- Header com `kid` para rotação de chaves

O code guidelines §12 exige: nunca diferenciar "usuário não encontrado" vs "senha incorreta". O ADR-010 reforça: executar bcrypt mesmo quando o usuário não existe (comparar com hash dummy) para normalizar tempo de resposta e prevenir timing attacks.

### Ação 1.1

Definir interface PasswordService no domain (ou em um package de serviço que o use case depende):
```go
type PasswordService interface {
    Compare(plain, hashed string) bool
}
```
Implementação: BcryptPasswordService usa `golang.org/x/crypto/bcrypt.CompareHashAndPassword`.

### Observação 1.1

O use case sempre chamará Compare, mesmo quando user == nil. Passará um hash dummy (ex: `$2a$12$...` válido mas irrelevante) para normalizar o tempo. O bcrypt tem tempo constante por design.

---

### Pensamento 2

Para o JWT, a interface pode viver no domain ou em um package dedicado. O use case precisa de: dado um conjunto de claims, retornar um token string assinado. O domínio não deve conhecer RSA nem JWT library — apenas uma interface abstrata.

### Ação 1.2

Definir interface em `domain/service/` ou similar:
```go
type JWTService interface {
    Sign(ctx context.Context, claims JWTClaims) (string, error)
}
```

O tipo JWTClaims precisa ter os campos: Issuer, Subject, Audience, TenantID, UserAccessProfile, JTI, Iat, Exp, Nbf. O domínio pode definir uma struct de input para o Sign, ou o use case constrói um map. Para manter o domínio puro, JWTClaims pode ser uma struct em um package `domain/valueobject` ou no próprio use case. A interface JWTService recebe uma struct com os campos necessários — definida no boundary entre use case e infrastructure. Pode viver em `usecase/auth/dto` ou em um package `domain/auth`.

### Observação 1.2

A interface JWTService será implementada na infrastructure. O use case chamará `jwtService.Sign(ctx, claims)`. A struct de claims pode ser definida no mesmo package da interface (domain) para não criar acoplamento com biblioteca JWT.

---

### Pensamento 3

Estrutura do JWT conforme especificação do usuário:
- `iss`: identificador do wisa-crm-service. Configurável via env (ex: `JWT_ISSUER` ou `AUTH_ISSUER`). Valor default: `"wisa-crm-service"` para facilitar adaptação futura para domínio.
- `sub`: user_id (UUID ou ULID do usuário)
- `aud`: slug do tenant + domínio base (ex: cliente1.app.wisa-crm.com) — construído no use case
- `tenant_id`: UUID do tenant
- `user_access_profile`: admin, editor, viewer — o banco tem admin, operator, view. Usar valores do banco.
- `jti`: ID único do token (UUID ou ULID) para revogação
- `iat`: issued at (Unix timestamp)
- `exp`: expiration (iat + 15 min)
- `nbf`: not before (igual a iat)

### Ação 1.3

Claims struct:
```go
type JWTClaims struct {
    Issuer            string    `json:"iss"`
    Subject           string    `json:"sub"`
    Audience          string    `json:"aud"`
    TenantID          string    `json:"tenant_id"`
    UserAccessProfile string    `json:"user_access_profile"`
    JTI               string    `json:"jti"`
    IssuedAt          int64     `json:"iat"`
    ExpiresAt         int64     `json:"exp"`
    NotBefore         int64     `json:"nbf"`
}
```

A biblioteca `github.com/golang-jwt/jwt/v5` ou similar. Usar `jwt.RegisteredClaims` com claims customizados. O ADR-006 exige `kid` no header — a implementação deve definir um kid fixo (ex: "key-2026-v1") ou configurável.

### Observação 1.3

O `iss` sem domínio: usar `JWT_ISSUER` env com default `"wisa-crm-service"`. Quando o domínio estiver configurado, alterar para `"https://auth.wisa-crm.com"` ou similar. O código deve ler de config, não hardcodar.

---

### Pensamento 4

Implementação RSA: a chave privada será lida de arquivo. O caminho vem de env `JWT_PRIVATE_KEY_PATH`. Nunca commitar a chave. O .gitignore já deve ter `*.pem`. A implementação carrega a chave no startup e usa para assinar. Usar `crypto/rsa` e `crypto/x509` para parse da PEM.

### Ação 1.4

Criar `infrastructure/crypto/rsa_jwt_service.go`:
- Construtor recebe: path da chave privada, issuer, expiration (minutos), kid
- Método Sign(claims) → assina com RS256, retorna token string
- Gerar jti com uuid.New().String() ou ulid
- Calcular iat = time.Now().Unix(), exp = iat + 15*60, nbf = iat

### Observação 1.4

A biblioteca jwt-go exige que o signing key seja `*rsa.PrivateKey`. Carregar do arquivo PEM no construtor. Em caso de falha ao carregar, o serviço não deve iniciar (fail fast).

---

### Pensamento 5

O BcryptPasswordService: cost factor. O bcrypt usa um cost (ex: 12) para o hash. A comparação não precisa do cost — apenas CompareHashAndPassword. O service não precisa de config além do cost para futuros Create (hash de nova senha). Para Compare, nenhum config necessário.

### Ação 1.5

Criar `infrastructure/crypto/bcrypt_password_service.go`:
```go
type BcryptPasswordService struct{}
func (s *BcryptPasswordService) Compare(plain, hashed string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
    return err == nil
}
```
Simples e direto.

### Observação 1.5

Nenhuma dependência de config. O use case garante que hashed nunca seja vazio (usa dummy se user nil).

---

### Pensamento 6

Conformidade com ADR-006: algorithm confusion attack. O backend que emite o token deve usar explicitamente RS256. A biblioteca jwt ao assinar deve especificar `SigningMethodRS256`. Não há risco de confusion no emissor — o risco é no validador (cliente). O emissor apenas assina com RS256.

### Ação 1.6

Ao criar o token com jwt-go, usar `jwt.NewWithClaims(jwt.SigningMethodRS256, claims)`. Garantir que o método seja RS256 explicitamente.

### Observação 1.6

Segurança atendida. A chave privada nunca deve ser exposta. Documentar no .env.example: JWT_PRIVATE_KEY_PATH deve apontar para arquivo fora do repo.

---

### Decisão final Fase 3

**Implementar:**

1. **PasswordService**
   - Interface em `domain/service/password_service.go` ou `domain/repository` (se não, em usecase input)
   - Implementação `infrastructure/crypto/bcrypt_password_service.go` com Compare

2. **JWTService**
   - Interface em `domain/service/jwt_service.go` com Sign(claims) (string, error)
   - Struct JWTClaims no domain ou no mesmo package
   - Implementação `infrastructure/crypto/rsa_jwt_service.go`:
     - Carrega chave privada de arquivo (path via config)
     - Assina com RS256
     - Gera jti (UUID), iat, exp (15 min), nbf
     - Issuer e aud configuráveis

3. **Config**
   - Variáveis: JWT_PRIVATE_KEY_PATH, JWT_ISSUER (default "wisa-crm-service"), JWT_EXPIRATION_MINUTES (default 15), JWT_KEY_ID (default "key-2026-v1")

---

### Checklist de Implementação

1. [ ] Criar interface PasswordService e BcryptPasswordService
2. [ ] Criar interface JWTService e struct JWTClaims
3. [ ] Criar RSAJWTService com RS256
4. [ ] Carregar chave privada de arquivo
5. [ ] Gerar jti, iat, exp, nbf corretamente
6. [ ] Documentar JWT_PRIVATE_KEY_PATH em .env.example
7. [ ] Adicionar *.pem ao .gitignore se não existir

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-006 | RS256, 4096 bits, claims corretos |
| ADR-006 kid | Header kid para rotação |
| Code guidelines §12 | bcrypt com timing constante (use case) |
| Segurança | Chave em arquivo, fora do código |
| Configurável | iss e exp via env |

---

## Referências

- [docs/adrs/ADR-006-jwt-com-assinatura-assimetrica.md](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- [golang.org/x/crypto/bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
