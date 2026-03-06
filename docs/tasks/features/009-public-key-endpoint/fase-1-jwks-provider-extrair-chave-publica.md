# Fase 1 — JWKS Provider: Extrair Chave Pública RSA e Formatar em JWKS

## Objetivo

Criar um provedor de JWKS (JSON Web Key Set) que, a partir da chave privada RSA usada para assinar JWTs, extraia a chave pública e a formate no padrão RFC 7517. O provedor deve suportar múltiplas chaves para permitir rotação segura.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O `RSAJWTService` já carrega a chave privada RSA de um arquivo PEM e usa `jwt.ParseRSAPrivateKeyFromPEM` para obter `*rsa.PrivateKey`. Da chave privada RSA, a chave pública é obtida via `privateKey.Public()`, que retorna `*rsa.PublicKey`. O JWK para RSA requer os campos `n` (modulus) e `e` (exponent) em encoding base64url (RFC 4648), além de `kty`, `use`, `alg`, `kid`.

**Conclusão:** É necessário criar um serviço ou método que:
1. Receba o par (chave privada ou pública, kid)
2. Extraia `N` e `E` da chave pública RSA
3. Codifique em base64url (sem padding)
4. Retorne a estrutura JWK completa

### Ação 1.1

Pesquisar na biblioteca `github.com/golang-jwt/jwt` ou em pacotes Go se existe suporte nativo para serializar RSA public key em JWK. Alternativa: usar `encoding/base64` com `RawURLEncoding` e `crypto/rsa` diretamente.

```go
// crypto/rsa: PublicKey tem campos N e E (big.Int)
// base64.RawURLEncoding.EncodeToString(n.Bytes())
```

### Observação 1.1

O pacote `github.com/golang-jwt/jwt/v5` não fornece JWK out-of-the-box. Usar `crypto/rsa` e `encoding/base64` é suficiente. O módulo `N` em big.Int deve ser codificado como inteiro unsigned big-endian (sem zeros à esquerda removidos, exceto o byte mais significativo se for >= 0x80 para manter representação positiva em ASN.1). RFC 7518: para RSA, `n` é o modulus como base64url; `e` é o exponent — normalmente 65537 (`AQAB` em base64url).

### Pensamento 2

Para o encoding correto de `n` e `e` em JWK RSA:
- **n (modulus):** `N` do `rsa.PublicKey` é `*big.Int`. Deve ser convertido para bytes big-endian. O RFC 7518 especifica que inteiros são representados como array de bytes unsigned big-endian. `N.FillBytes(make([]byte, (N.BitLen()+7)/8)` ou `N.Bytes()` —注意: `Bytes()` pode retornar menos bytes se o valor tem leading zeros. O RFC diz que o valor deve ser o inteiro positivo. `FillBytes` (Go 1.15+) garante o tamanho correto. Para `e`, tipicamente 65537 — 3 bytes: 0x01 0x00 0x01.
- Biblioteca alternativa: `github.com/lestrrat-go/jwx/v2/jwk` fornece suporte JWK completo e pode simplificar. Verificar se já existe no go.mod ou se adicionar dependência é aceitável.

### Ação 1.2

Verificar dependências atuais do backend:

```bash
cd backend && go list -m all
```

Decisão: Se `lestrrat-go/jwx` ou similar não estiver presente, implementar manualmente com `crypto/rsa` e `encoding/base64` para evitar nova dependência pesada. A implementação é direta.

### Observação 1.2

Conformidade com code guidelines: evitar dependências desnecessárias. A serialização JWK para RSA é simples (~30 linhas). Preferir implementação manual no pacote `infrastructure/crypto`.

### Pensamento 3

Onde colocar a lógica? Opções:
- a) Método `GetJWKS()` no `RSAJWTService` — o serviço já possui a chave privada e o kid
- b) Novo `JWKSProvider` em `infrastructure/crypto` que recebe `RSAJWTService` ou os dados da chave
- c) Interface `domain/service/JWKSProvider` com implementação em crypto

O `RSAJWTService` já está injetado no main. Adicionar um método `GetJWKS(ctx) ([]JWK, error)` ou `GetJWKSJSON() ([]byte, error)` ao RSAJWTService manteria a coesão: um único serviço conhece a chave e o kid. Porém, para **múltiplas chaves**, o RSAJWTService atual possui apenas uma chave. A rotação futura exigiria um agregador de chaves.

**Decisão:** Criar interface `JWKSProvider` em `domain/service` com método `GetKeys(ctx) ([]JWK, error)`. Implementação `RSAJWKSProvider` em `infrastructure/crypto` que:
- Recebe a mesma config (PrivateKeyPath, KeyID) ou recebe o RSAJWTService e delega a ele a obtenção da chave
- Ou: criar `RSAJWKSProvider` que lê o mesmo arquivo de chave e produz JWKS

Para múltiplas chaves: a interface retorna `[]JWK`. A implementação inicial lê uma chave (a mesma do JWT). Futuramente, config `JWT_EXTRA_PUBLIC_KEYS` (caminho para JSON com chaves adicionais) pode ser adicionada. Por ora, retornar um slice com um único JWK.

### Ação 1.3

Definir interface e structs em `internal/domain/service/jwks_provider.go`:

```go
// JWK representa uma chave no formato JSON Web Key (RFC 7517)
type JWK struct {
    Kty string `json:"kty"`
    Use string `json:"use"`
    Alg string `json:"alg"`
    Kid string `json:"kid"`
    N   string `json:"n"`
    E   string `json:"e"`
}

// JWKS é o conjunto de chaves (RFC 7517)
type JWKS struct {
    Keys []JWK `json:"keys"`
}

// JWKSProvider fornece as chaves públicas para validação de JWT
type JWKSProvider interface {
    GetKeys(ctx context.Context) ([]JWK, error)
}
```

### Observação 1.3

O `ctx` permite cache ou fetching assíncrono no futuro. Para a implementação inicial, pode não ser usado.

### Pensamento 4

Implementação em `infrastructure/crypto/rsa_jwks_provider.go`:
- Carregar chave privada do mesmo path usado pelo RSAJWTService (ou receber o path na config)
- Parsear com `jwt.ParseRSAPrivateKeyFromPEM` ou `x509.ParsePKCS1PrivateKey` / `ParsePKCS8PrivateKey`
- Extrair `publicKey := privateKey.Public().(*rsa.PublicKey)`
- Para `n`: `nBytes := publicKey.N.FillBytes(make([]byte, (publicKey.N.BitLen()+7)/8))` — `FillBytes` escreve em big-endian
- Para `e`: o exponent 65537 em bytes = `[]byte{0x01, 0x00, 0x01}`. Caso E seja outro valor, usar `big.NewInt(int64(publicKey.E)).FillBytes(...)`
- Codificar ambos com `base64.RawURLEncoding.EncodeToString(nBytes)`
- Montar JWK com kty="RSA", use="sig", alg="RS256", kid do config

### Ação 1.4

Implementar `RSAJWKSProvider`:

```go
func NewRSAJWKSProvider(cfg RSAJWTConfig) (JWKSProvider, error) {
    // Reutilizar lógica de carregamento do RSAJWTService
    // Ou aceitar RSAJWTConfig com PrivateKeyPath e KeyID
}
func (p *rsaJWKSProvider) GetKeys(ctx context.Context) ([]JWK, error)
```

Garantir que o `kid` usado seja o mesmo de `JWT_KEY_ID` (configuração compartilhada).

### Observação 1.4

O `RSAJWTConfig` já existe e contém `PrivateKeyPath` e `KeyID`. O `RSAJWKSProvider` pode usar o mesmo config. Não duplicar leitura do arquivo: podemos ter um construtor que recebe `*RSAJWTService` e extrai a chave dele — mas o RSAJWTService guarda `privateKey []byte` e não expõe a chave parseada. A alternativa é o JWKSProvider ler o arquivo independentemente. Duplicação mínima de I/O (uma leitura no startup do RSAJWTService e uma no JWKSProvider — ou podemos fazer o JWKSProvider receber os bytes da chave no construtor, injetados pelo main que já carrega o arquivo). Para simplificar: JWKSProvider lê o arquivo. Em produção, o arquivo está em disco e a leitura é feita no startup; o GetKeys pode cachear o resultado em memória (a chave não muda durante a execução). Por ora, fazer o parse da chave em cada GetKeys seria ineficiente — cachear o []JWK no provider após primeira chamada. Ou: como o JWKS é estático durante a vida do processo, o provider pode computar no `New` e armazenar; `GetKeys` retorna o slice armazenado.

### Pensamento 5 (Múltiplas chaves)

O requisito exige suporte a múltiplas chaves para rotação. A implementação inicial pode retornar `keys: [uma chave]`. A estrutura já está pronta. Para adicionar uma segunda chave no futuro, seria necessário:
- Variável `JWT_EXTRA_PUBLIC_KEYS_PATH` — caminho para arquivo JSON com array de JWKs adicionais (apenas chaves públicas, não privadas)
- Ou `JWT_JWKS_PATH` — caminho para arquivo JWKS completo que substitui/estende a chave derivada da privada

Para a Fase 1, documentar na interface que `GetKeys` retorna todas as chaves ativas. A implementação retorna a chave derivada da privada. O design permite extensão futura sem quebrar a API.

### Ação 1.5

Não adicionar configuração de chaves extras nesta fase. Garantir que o slice retornado tenha pelo menos um elemento. Em caso de erro ao carregar a chave, `NewRSAJWKSProvider` deve retornar erro (fail-fast).

### Observação 1.5

Conformidade com ADR-006: rotação sem downtime. O formato `{"keys": [...]}` já suporta múltiplas entradas.

### Pensamento 6 (Segurança)

A chave pública pode ser exposta — não há risco de segurança em distribuí-la. O endpoint será público. Garantir que apenas a chave pública (n, e) seja exposta, nunca a privada. O provider lê o arquivo da chave privada e deriva a pública — a privada nunca deixa o servidor. O arquivo de chave privada deve estar em path restrito (ADR-009).

### Observação 1.6

Sem novas considerações de segurança além das já documentadas.

### Pensamento 7 (Encoding base64url)

Em Go, `base64.RawURLEncoding` produz base64 sem padding, conforme RFC 4648 base64url. O RFC 7517 usa base64url. Correto.

Para inteiros com leading zeros: `big.Int` em Go, quando o valor tem bit mais significativo 1, o primeiro byte pode ser 0 para manter o valor positivo em representação de complemento de dois. Em JWK, os inteiros são unsigned. `FillBytes` com tamanho `(N.BitLen()+7)/8` garante o número correto de bytes. Se `N.BitLen()` for 4096, teremos 512 bytes.

### Ação 1.6

Implementar helper para converter rsa.PublicKey em JWK:

```go
func rsaPublicKeyToJWK(pub *rsa.PublicKey, kid string) (JWK, error) {
    nBytes := pub.N.FillBytes(make([]byte, (pub.N.BitLen()+7)/8))
    eBig := big.NewInt(int64(pub.E))
    eBytes := eBig.FillBytes(make([]byte, (eBig.BitLen()+7)/8))
    if len(eBytes) == 0 {
        eBytes = []byte{0x01, 0x00, 0x01}
    }
    return JWK{
        Kty: "RSA",
        Use: "sig",
        Alg: "RS256",
        Kid: kid,
        N:   base64.RawURLEncoding.EncodeToString(nBytes),
        E:   base64.RawURLEncoding.EncodeToString(eBytes),
    }, nil
}
```

### Observação 1.7

Para `E=65537`, `eBytes` deveria ser `[]byte{0x01, 0x00, 0x01}`. `FillBytes` com tamanho correto produzirá isso. Validar com teste unitário.

---

## Checklist de Implementação

- [ ] 1. Criar `internal/domain/service/jwks_provider.go` com interface JWKSProvider e structs JWK, JWKS
- [ ] 2. Criar `internal/infrastructure/crypto/rsa_jwks_provider.go` com RSAJWKSProvider
- [ ] 3. Implementar conversão rsa.PublicKey → JWK (n, e em base64url)
- [ ] 4. Usar RSAJWTConfig (PrivateKeyPath, KeyID) no construtor
- [ ] 5. Cachear resultado no construtor (computar uma vez)
- [ ] 6. Teste unitário: verificar que n e e são válidos e decodificáveis

---

## Referências

- RFC 7517 — JSON Web Key (JWK)
- RFC 7518 — JSON Web Algorithms (JWA) — RSA
- ADR-006 — JWT com Assinatura Assimétrica
- docs/code_guidelines/backend.md
- backend/internal/infrastructure/crypto/rsa_jwt_service.go
