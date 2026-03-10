# Documentação de Contexto — Sistema Centralizador de Autenticação

**Software:** `wisa-crm-service`

## Objetivo do Sistema

Desenvolver o **wisa-crm-service**, um sistema centralizador de autenticação e validação de assinaturas que funcione como uma **camada central de segurança** para os softwares comercializados aos clientes.

---

## Ideia Principal

O **wisa-crm-service** será responsável por:

- **Autenticar usuários**
- **Validar assinaturas**
- **Emitir tokens JWT seguros** para permitir o acesso ao software principal

---

## Fluxo de Funcionamento

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────────────┐
│   Usuário   │────▶│  Software do     │────▶│   wisa-crm-service       │
│             │     │  Cliente         │     │  (Auth Centralizador)    │
└─────────────┘     └──────────────────┘     └─────────────────────────┘
```

### 1. Acesso Inicial

O usuário dá duplo clique e abre o software que adquiriu.

### 2. Requisição ao Back-end

A requisição inicial é direcionada para o back-end do sistema do cliente.

### 3. Redirecionamento para Login

Ao identificar que o usuário ainda não está autenticado, o sistema do cliente **redireciona o usuário** para a página de login do sistema centralizador de autenticação.

### 4. Credenciais na Interface Central

O usuário informa seu **usuário** e **senha** na interface própria do sistema centralizador.

### 5. Validações no Sistema Centralizador

O sistema centralizador executa as seguintes verificações:

| Etapa | O que é verificado |
|-------|--------------------|
| **Identificação do tenant** | Identifica qual é o cliente (tenant) |
| **Validação de assinatura** | Verifica se a assinatura está ativa e em dia |
| **Validação de credenciais** | Valida usuário e senha |

### 6. Emissão do JWT (Authorization Code Flow)

Se **todas as validações forem aprovadas**, o sistema centralizador:
- Gera um **authorization code** temporário (TTL 40s) e armazena no Redis;
- Redireciona o usuário (HTTP 302) para o callback do sistema do cliente com `?code=...&state=...`;
- O sistema do cliente troca o code por JWT via `POST /api/v1/auth/token`, obtendo o token sem expô-lo na URL.

### 7. Recepção e Validação pelo Cliente

O sistema do cliente:

- Recebe o token
- Valida sua **autenticidade**
- Valida sua **origem**

### 8. Liberação de Acesso

Após validar o token, o sistema permite que o usuário acesse normalmente sua API e utilize o software.

---

## Requisito de Segurança Importante

### Estrutura do JWT

O JWT gerado deve conter uma **assinatura criptográfica** que identifique claramente sua origem (issuer).

Além da assinatura digital padrão, o token deve incluir:

| Claim | Descrição |
|-------|-----------|
| `iss` (issuer) | Identifica o wisa-crm-service como emissor |
| — | Identificação do cliente (tenant) |
| — | Identificação do usuário |
| `exp` | Data de expiração |
| `aud` (audiência) | *Opcional* — Garante que o token só possa ser usado pelo sistema correto |

### Validação Obrigatória na Aplicação do Cliente

A aplicação do cliente, ao receber o JWT, **deve obrigatoriamente**:

1. **Validar a assinatura criptográfica** usando a chave pública do wisa-crm-service
2. **Verificar se o `iss`** corresponde ao emissor autorizado
3. **Validar a audiência (`aud`)** quando aplicável
4. **Verificar a expiração do token** (`exp`)

> **Importante:** Somente após **todas** essas validações o acesso deve ser concedido.

---

## Resumo do Modelo de Segurança

```
┌──────────────────────────────────────────────────────────────────────┐
│                    wisa-crm-service (Sistema Centralizador)            │
│  • Possui chave PRIVADA para assinar tokens                           │
│  • Valida credenciais + assinatura + tenant                           │
│  • Emite JWT assinado                                                 │
└──────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ JWT assinado
                                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│                    SISTEMA DO CLIENTE                                 │
│  • Possui chave PÚBLICA para validar tokens                           │
│  • Valida: assinatura + iss + aud + exp                               │
│  • Concede ou nega acesso                                             │
└──────────────────────────────────────────────────────────────────────┘
```

---

## Referências Relacionadas

- [Integração com o Fluxo de Authorization Code](./integration/auth-code-flow-integration.md) — Guia definitivo para desenvolvedores integrarem o login em suas aplicações
- [Test App](../test-app/README.md) — Aplicação de teste (Angular + Go) que demonstra a integração completa com o Auth Code Flow
- [NGINX para testes E2E](../nginx/README.md) — Configuração do NGINX como reverse proxy para testes com subdomínios (auth.wisa.labs.com.br, lingerie-maria.wisa.labs.com.br)
- [Solução 2 — Arquitetura de Autenticação e Gestão de Assinaturas](./DON'T READ/ideias/ideias.md)
