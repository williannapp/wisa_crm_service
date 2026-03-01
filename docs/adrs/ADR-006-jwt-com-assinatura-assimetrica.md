# ADR-006 — JWT com Assinatura Assimétrica (RS256 / ES256)

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (segurança — emissão e validação de tokens)

---

## Contexto

O `wisa-crm-service` atua como Identity Provider (IdP) centralizado: emite tokens JWT que são validados por múltiplos sistemas clientes distribuídos em VPS independentes. Este é o mecanismo central de segurança de toda a plataforma.

O design do JWT e do sistema de assinatura precisa responder a perguntas críticas:

1. **Como garantir que um sistema cliente consiga validar um JWT sem precisar comunicar-se com o `wisa-crm-service` em cada requisição?** (autenticação offline/local)
2. **Como garantir que nenhum sistema cliente possa forjar um JWT válido?** (não-repúdio)
3. **Como garantir que um JWT emitido para o cliente A não seja aceito pelo cliente B?** (isolamento de audience)
4. **O que acontece se a chave de assinatura for comprometida?** (plano de rotação)
5. **Quanto tempo um JWT deve ser válido?** (balanceamento entre UX e segurança)

O modelo de assinatura escolhido tem impacto direto em todos esses pontos.

---

## Decisão

**JWT com assinatura assimétrica usando o algoritmo RS256 (RSA + SHA-256) será o mecanismo de autenticação entre o `wisa-crm-service` e os sistemas clientes.**

**ES256 (ECDSA + P-256 + SHA-256) será adotado em versões futuras** como alternativa de menor tamanho de chave e maior performance criptográfica.

O `wisa-crm-service` mantém a **chave privada RSA de 4096 bits exclusivamente em ambiente controlado**. A chave pública correspondente é distribuída para todos os sistemas clientes via endpoint HTTPS autenticado (`/api/v1/public-key`).

---

## Justificativa

### 1. Por que assimétrico (RS256) ao invés de simétrico (HS256)?

**HS256 (HMAC-SHA256)** usa a mesma chave para assinar e verificar. Em um modelo com múltiplos sistemas clientes, isso significa:

- **Cada sistema cliente precisaria da chave secreta** — isso é inaceitável: se um único cliente for comprometido, a chave de **toda a plataforma** é exposta
- **Qualquer sistema cliente poderia forjar tokens** — um cliente comprometido poderia criar tokens válidos para qualquer outro tenant

**RS256 (RSA-SHA256)** usa par de chaves pública/privada:

- **Somente o `wisa-crm-service` possui a chave privada** — somente ele pode emitir tokens
- **Sistemas clientes recebem apenas a chave pública** — conseguem validar mas nunca forjar
- **Comprometimento de um cliente não afeta o `wisa-crm-service`** — a chave pública não serve para assinar

```
┌─────────────────────────────────────────────────────────────┐
│              wisa-crm-service                               │
│  Chave PRIVADA RSA 4096-bit                                 │
│  → Assina JWT (somente aqui)                                │
│  → Armazenada em arquivo protegido, fora do código          │
└─────────────────────────────────────────────────────────────┘
              │ distribui via HTTPS
              ▼
┌──────────────────────────────────┐  ┌──────────────────────────────────┐
│   Cliente A                      │  │   Cliente B                      │
│   Chave PÚBLICA (cópia)          │  │   Chave PÚBLICA (cópia)          │
│   → Valida assinatura do JWT     │  │   → Valida assinatura do JWT     │
│   → NÃO pode forjar JWT          │  │   → NÃO pode forjar JWT          │
└──────────────────────────────────┘  └──────────────────────────────────┘
```

### 2. Estrutura do JWT emitido

```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "key-2026-v1"
  },
  "payload": {
    "iss": "https://auth.wisa-crm.com",
    "sub": "usr_01HXYZ...",
    "aud": "cliente1.seusistema.com",
    "tenant_id": "ten_01HXYZ...",
    "plan": "pro",
    "jti": "01HXYZ...",
    "iat": 1740000000,
    "exp": 1740000900,
    "nbf": 1740000000
  }
}
```

**Justificativa de cada claim:**

| Claim | Valor | Justificativa |
|-------|-------|---------------|
| `iss` | URL do auth service | Permite ao cliente verificar que o token veio do emissor correto |
| `sub` | user_id (opaque ULID) | Identifica o usuário sem expor dados PII no token |
| `aud` | domínio do cliente | **Crítico:** impede que token do Cliente A seja aceito pelo Cliente B |
| `tenant_id` | tenant ULID | Identificação de tenant para lógica multi-tenant no cliente |
| `plan` | nome do plano | Permite controle de funcionalidades por plano sem consulta ao banco |
| `jti` | JWT ID único (ULID) | Permite revogação individual de token e prevenção de replay attacks |
| `iat` | issued at | Momento de emissão |
| `exp` | expiration | **Curta duração:** 15 minutos para tokens de acesso |
| `nbf` | not before | Previne uso de token antes do momento de emissão (proteção contra skew de relógio) |
| `kid` | key ID | Identifica qual par de chaves foi usado — essencial para rotação de chaves |

### 3. Duração do token e Refresh Tokens

**Access Token:** 15 minutos  
**Refresh Token:** 7 dias (armazenado como hash no banco, em tabela com controle de revogação)

**Justificativa da duração curta (15 min):**
- Em caso de comprometimento de um token, a janela de exploração é mínima
- O controle de bloqueio por inadimplência funciona em no máximo 15 minutos: quando o cliente para de pagar, o `wisa-crm-service` simplesmente não emite novo access token na renovação
- Alinhado com o requisito de bloqueio rápido documentado no contexto do sistema

**Fluxo de renovação:**
1. Access token expira
2. Frontend detecta 401 e envia refresh token para `/api/v1/auth/refresh`
3. `wisa-crm-service` valida o refresh token no banco, verifica se a assinatura ainda está ativa, emite novo access token
4. Se assinatura vencida → 402 Payment Required, usuário é redirecionado para renovação

### 4. `kid` (Key ID) para rotação segura de chaves

O campo `kid` no header do JWT permite que múltiplos par de chaves coexistam durante uma rotação:

```
Rotação planejada:
1. Gerar novo par de chaves (key-2026-v2)
2. Publicar nova chave pública no endpoint /jwks.json (JWKS format)
3. Novo tokens são emitidos com kid=key-2026-v2
4. Tokens antigos (kid=key-2026-v1) continuam válidos até expiração natural (15 min)
5. Após 15 minutos, a chave antiga pode ser removida do JWKS
```

Isso garante **zero downtime durante rotação de chaves**.

### 5. JWKS Endpoint para distribuição de chaves públicas

Ao invés de distribuir a chave pública manualmente para cada cliente, o `wisa-crm-service` expõe um endpoint padrão da indústria:

```
GET https://auth.wisa-crm.com/.well-known/jwks.json
```

Resposta:
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-2026-v1",
      "n": "...(modulus base64url)...",
      "e": "AQAB"
    }
  ]
}
```

Sistemas clientes podem fazer cache desta resposta (com TTL de 24h) e auto-atualizar durante rotação de chaves.

### 6. Algoritmo ES256 como evolução futura

ES256 (ECDSA com curva P-256) oferece vantagens sobre RS256:

- **Chaves menores:** 256-bit vs 4096-bit RSA com segurança equivalente (~128-bit)
- **Operações mais rápidas:** especialmente na assinatura
- **Tokens menores:** a assinatura ECDSA é ~64 bytes vs ~512 bytes RSA

A adoção de ES256 será feita em uma rotação futura de chaves, mantendo total retrocompatibilidade via `kid`.

---

## Consequências

### Positivas
- Isolamento total: compromisso de um cliente não afeta o `wisa-crm-service`
- Sistemas clientes validam tokens localmente sem consultar o auth service em cada requisição
- Bloqueio por inadimplência efetivo em no máximo 15 minutos (tempo de vida do access token)
- Rotação de chaves sem downtime via `kid` e JWKS
- `aud` previne uso de token de um cliente em outro sistema

### Negativas
- Operações RSA são mais lentas que HMAC — mitigado pela curta duração (poucas renovações) e hardware moderno
- Revogação imediata de access token não é possível sem um denylist — mitigado pela curta duração de 15 min
- Gerenciamento do arquivo de chave privada requer processo operacional seguro

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| Comprometimento da chave privada RSA | Baixa | Crítico | Crítica |
| Algorithm confusion attack (`alg: none` ou `alg: HS256`) | Média | Crítico | Alta |
| Replay attack com token válido interceptado | Baixa | Alto | Alta |
| Uso de token após bloqueio do tenant (dentro dos 15 min) | Alta | Médio | Média |
| Chave privada exposta em repositório git | Baixa | Crítico | Alta |
| Clock skew entre `wisa-crm-service` e sistemas clientes | Média | Baixo | Baixa |

---

## Mitigações

### Comprometimento da chave privada
- Armazenar a chave privada **fora do repositório** — nunca em variáveis de ambiente no código
- Usar arquivo de chave com permissão restrita: `chmod 400 /etc/wisa-crm/private.pem` e `chown wisa-crm-user:wisa-crm-user`
- Rotação de chaves semestral planejada (ou imediata em caso de suspeita de comprometimento)
- Considerar Hardware Security Module (HSM) ou AWS KMS para ambientes com requisitos de compliance mais rigorosos

### Algorithm confusion attack
- O backend Go **deve rejeitar qualquer token cujo header `alg` não seja `RS256`** — nunca aceitar `alg: none`
- Usar bibliotecas JWT que exijam especificação explícita do algoritmo aceito:

```go
// CORRETO — especificar algoritmo explicitamente
token, err := jwt.Parse(tokenString, keyFunc, jwt.WithValidMethods([]string{"RS256"}))

// INCORRETO — aceita qualquer algoritmo, incluindo none
token, err := jwt.Parse(tokenString, keyFunc)
```

### Replay attack
- O campo `jti` (JWT ID) único permite implementar um denylist de tokens usados
- Para o endpoint de logout, marcar o `jti` como revogado no banco
- Para tokens de vida curta (15 min), a janela de replay é mínima — custo/benefício de verificar `jti` em cada requisição deve ser avaliado vs performance

### Uso de token após bloqueio
- Aceitar a janela de 15 minutos como trade-off documentado e aprovado pelo negócio
- Para casos urgentes (fraude, comprometimento de conta), manter um denylist de `jti` com TTL de 15 minutos

### Chave em repositório git
- Adicionar `*.pem`, `*.key`, `private_key*` ao `.gitignore` globalmente
- Configurar `git-secrets` na CI para detectar padrões de chave privada em commits
- Executar `trufflehog` ou `gitleaks` na CI para varredura de segredos

### Clock skew
- Configurar NTP em todas as VPS
- Usar o campo `nbf` com tolerância de 30 segundos para clock skew
- A biblioteca JWT Go deve aceitar `leeway` de 30s ao validar `exp` e `nbf`

---

## Processo de Rotação de Chaves

```
1. Gerar novo par: openssl genrsa -out /tmp/new_private.pem 4096
2. Extrair pública: openssl rsa -in /tmp/new_private.pem -pubout -out /tmp/new_public.pem
3. Atualizar JWKS no banco/config com novo kid
4. Deploy do serviço com novo kid ativo para emissão
5. Aguardar 15 minutos (expiração de todos os tokens antigos)
6. Remover kid antigo do JWKS
7. Destruir chave privada antiga com shred: shred -u /path/to/old_private.pem
```

---

## Alternativas Consideradas

### HS256 (HMAC-SHA256)
- **Prós:** Simples, performático, sem gestão de par de chaves
- **Contras:** Requer compartilhar o segredo com todos os clientes — qualquer cliente comprometido expõe a chave de toda a plataforma; clientes podem forjar tokens; **inaceitável para modelo multi-tenant distribuído**

### Opaque Tokens (tokens opacos)
- **Prós:** Revogação imediata possível, sem informação no token
- **Contras:** Todo sistema cliente precisa consultar o `wisa-crm-service` para validar cada requisição — cria acoplamento forte e ponto único de falha; latência adicional em cada request; viola o requisito de validação local nos sistemas clientes

### PASETO (Platform-Agnostic Security Tokens)
- **Prós:** Mais moderno que JWT, evita armadilhas de implementação JWT (sem `alg: none`), versão v4 usa Ed25519
- **Contras:** Menor adoção e ecossistema de bibliotecas mais limitado, curva de aprendizado maior, não há suporte nativo amplo em sistemas clientes que possam ser de terceiros

**RS256 com as mitigações documentadas representa o equilíbrio ideal entre segurança, interoperabilidade e operabilidade para este sistema.**

---

## Referências

- [RFC 7519 — JSON Web Token (JWT)](https://tools.ietf.org/html/rfc7519)
- [RFC 7517 — JSON Web Key (JWK)](https://tools.ietf.org/html/rfc7517)
- [RFC 7518 — JSON Web Algorithms (JWA)](https://tools.ietf.org/html/rfc7518)
- [Critical vulnerabilities in JSON Web Token libraries (auth0)](https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/)
- [OWASP JWT Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [golang-jwt/jwt library](https://github.com/golang-jwt/jwt)
