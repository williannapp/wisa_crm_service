# Wisa CRM Service - Backend Implementation Guide

Este documento descreve a divisão da implementação do backend do projeto **Wisa CRM Service** em etapas organizadas.

---

# 📌 BACKEND - Basic Structure (Directories and Imports)

## 1. Estrutura Inicial do Projeto

1. Crie a estrutura de diretórios do backend dentro da pasta `backend`.
2. Importe as bibliotecas necessárias para iniciar o projeto.
3. Adicione o arquivo `.gitignore` para evitar que arquivos desnecessários sejam enviados ao repositório.
4. Crie um `.venv` para armazenar as variáveis de ambiente necessárias para rodar o projeto.
5. Crie um `Dockerfile` para executar o código do backend.
6. Crie um endpoint `health` para validar se a aplicação está rodando corretamente

---

# 📌 BACKEND - Database Configuration

## 1. Estrutura Base do Banco de Dados

- Adicione as configurações básicas do banco de dados na estrutura do projeto:
  - Crie as pastas e classes necessárias para execução de queries.
  - Realizer os imports necessários.
  - ⚠️ **IMPORTANTE:** Não implementar nenhuma query ou tabela neste momento. Implementar apenas a estrutura base para que a aplicação consiga se conectar ao banco.

## 2. Variáveis de Ambiente

- Crie as variáveis de ambiente necessárias para conexão com o banco de dados.

## 3. Containers

- Crie um `Dockerfile` separado para o banco de dados.
- Crie um `docker-compose.yml` na pasta raiz do projeto `wisa_crm_service`.

## 4. Documentação

- Crie uma documentação dentro de `@docs` explicando como rodar o backend.
- Caso alguma configuração adicional seja necessária na VPS para que o código funcione conforme descrito nas ADRs, adicione na documentação em: `@docs/vps/configurations.md`

## 5. ORM

- Realize as adpatações no código para o ORM.
- Adicione a estrutura base do ORM ao projeto.
- Implemente as migrations.
- Implemente a funcionalidade de rollback do ORM.

---

# 📌 BACKEND - Criação da estrutura de banco de dados

- Releia a documentação e implemente:
- Crie um schema no banco de dados com o nome: `wisa-labs-db`
- Implemente a estruturas de tabelas no banco de dados:
  - tenants
  - products
  - subscriptions
  - payments
  - users
  - user_product_access
  - refresh_tokens
  - audit_logs

---

# 📌 BACKEND - Criação de Package de Erro Padronizado


- Criar um package responsável por:
  - Padronizar respostas de erro HTTP
  - Mapear erros de domínio para erros de aplicação
  - Impedir vazamento de informações internas
  - Manter conformidade com Clean Architecture
  - Garantir consistência nas respostas da API

- Crie a esttrutura de diretorio para as exceptions
- Crie um tipo estruturado de Erro com ad seguintes informações:
    - código interno padronizado (ex: INVALID_CREDENTIALS)
    - mensagem amigável ao cliente
    - opcional (não expor informações sensíveis)
    - status HTTP correspondente


# 📌 BACKEND - Endpoint de Login

## 1. Implementação do Endpoint de Autenticação

O endpoint deve:

- Receber como parâmetros:
- `slug`
- `user_email`
- `password`
- Validar se o email e senha são válidos para o `slug` informado.
- Validar se o user está com status como `active`
- Obter as informações do tenant com base no `slug`.
- Verificar se o cliente possui assinatura com status:
- `active`
- `pending`

Se todas as validações forem positivas:

- Gerar um token JWT assinado.
- Retornar o token JWT como resposta.

### IMPORTANTE

- O campo `iss` deve identificar o `wisa-crm-service` como emissor.
- O campo `aud` deve conter o dominio do cliente
- O JWT deve conter:
    - iss
    - sub
    - aud
    - tenant_id
    - user_access_profile - enabled options: admin, editor, viewer
    - jti
    - iat
    - exp
    - nbf

---

# 🔎 Validações Obrigatórias

## 1. Email

- Validar se o email existe na base de dados para o `slug` informado.
- Caso não exista, retornar erro informando que o email não está cadastrado para o slug informado.

## 2. Senha

- Verificar se a senha informada é válida.
- Caso não seja, retornar erro informando que a senha é inválida.

## 3. User Status

- Verificar se o status do usuario está como ativo.
- Caso não esteja, retornar erro informando que o usuario não ter pemissões para acessar o sistema.

## 3. Assinatura

### Status: `suspended`

Significa que o acesso foi suspenso por falta de pagamento.

Retornar:

```json
{
    "status": "error",
    "error_code": "SUBSCRIPTION_SUSPENDED",
    "message": "Acesso suspenso por pendência financeira.",
    "details": "Sua assinatura não está ativa devido a pagamentos em aberto. Por favor, atualize sua forma de pagamento para acessar o software."
}

```

### Status: `canceled`

Significa que a assinatura foi cancelada.

Retornar:

```json
{
  "status": "error",
  "error_code": "SUBSCRIPTION_CANCELED",
  "message": "Assinatura cancelada.",
  "details": "Sua assinatura foi cancelada. Entre em contato com a equipe Wisa Labs para analisar o caso."
}

```
---

# 📌 BACKEND - Refresh Token


1) Implementar endpoint para renovação de sessão.

- O endpoint deve:

    - Receber o refresh token
    - Validar hash no banco

- Verificar:
    - Existência
    - revoked_at IS NULL
    - expires_at > NOW()
    - Caso inválido → retornar 401 Unauthorized com erro genérico.
    - Se válido:
        - Revogar o refresh token atual (revoked_at = NOW())
        - Gerar:
            - Novo Access Token (expiração de 15 minutos)
            - Novo Refresh Token (expiração até D+7)
            - Inserir novo refresh token no banco como hash
        - O Access Token deve:
            - Não ser armazenado em banco
        - O Refresh Token deve:
            - Ser rotativo (cada uso invalida o anterior)
            - Ser armazenado como SHA-256 hash
            - Ter expiração máxima de 7 dias
            - Caso o refresh token esteja expirado, revogado ou inexistente:
                - Retornar 401 Unauthorized
                - Não revelar o motivo específico


- Caso alguma configuração adicional seja necessária na VPS para que o código funcione conforme descrito nas ADRs, adicionar a documentação em: `@docs/vps/configurations.md`
---

# 📌 BACKEND - JWKS Endpoint para distribuição de chaves públicas


- Crie um endpoint público para distribuir a chave pulica para os clientes.
- O endpoint deve:
  - Retornar Content-Type application/json
  - Não exigir autenticação
  - Estar disponível exclusivamente via HTTPS
  - Permitir cache via Cache-Control
  - Deve ser retornado: 
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-2026-v1",
      "n": "base64url_modulus",
      "e": "AQAB"
    }
  ]
}

```
- O campo kid deve corresponder ao kid presente no header dos JWT emitidos.
- O endpoint deve suportar múltiplas chaves simultaneamente para permitir rotação segura. Exemplo: 

```json
{
  "keys": [
    { "kid": "key-2026-v1", ... },
    { "kid": "key-2026-v2", ... }
  ]
}

```

- Deve incluir header de cache:
Cache-Control: public, max-age=86400


- Caso alguma configuração adicional seja necessária na VPS para que o código funcione conforme descrito nas ADRs, adicionar a documentação em: `@docs/vps/configurations.md`


------

# 📌 BACKEND - Implementar Rate Limit No Login

- Se a tentantiva de login exeder um limite máximo de 5 tentativas por segundo por IP e tenant, sistema deve 
impedir novas tentativas até o tempo expirar.
- Implemente o valor de tentativas o valor de minutos em variaveis de ambiente para que seja facil aumentar ou diminuir caso seja necessário
- Utilize a bliblioteca go para a implementação: net/http
- Caso o numero de tentativas tenha se excedido retorne o erro:

```json
HTTP 429
{
  "status": "error",
  "error_code": "RATE_LIMIT_EXCEEDED",
  "message": "Muitas tentativas. Tente novamente mais tarde."
}
```
- Aplicar rate limit antes de chegar no endpoint de login
---

# 📌 BACKEND - Implementar Auditoria

1) Regristre auditoria para operações como:
    - login_success
    - login_failed
    - refresh_success
    - refresh_failed
    - subscription_blocked


# 📌 BACKEND - Cron Job Para Rotação de Chave privada


# 📌 BACKEND - Cron Job Para limpeza de audit logs


# 📌 BACKEND - Implementação de Testes Unitarios