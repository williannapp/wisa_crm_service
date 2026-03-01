# 📘 Arquitetura de Autenticação e Gestão de Assinaturas

## 📌 Visão Geral

Esta solução descreve uma arquitetura distribuída para um sistema de
vendas executado em múltiplas VPS (uma por cliente), com controle
centralizado de autenticação e gestão de assinaturas.

### Objetivos:

-   Garantir controle de acesso baseado em assinatura
-   Permitir bloqueio rápido em caso de inadimplência
-   Centralizar segurança e autenticação
-   Manter escalabilidade

------------------------------------------------------------------------

# 🏗 Arquitetura Geral

## 1️⃣ VPS Geral (Core Central)

Responsável por:

-   Auth Service (autenticação)
-   Gestão de assinaturas
-   Banco central de usuários e clientes
-   Geração e assinatura de JWT
-   Chave privada para assinatura de tokens

Atua como: **Identity Provider (IdP)**

------------------------------------------------------------------------

## 2️⃣ VPS por Cliente (Business Node)

Cada cliente possui sua própria VPS contendo:

-   Sistema de vendas
-   Banco de dados isolado
-   Middleware de validação de token
-   Chave pública para validação de JWT

Atua como: **Service Provider (SP)**

------------------------------------------------------------------------

# 🔐 Modelo de Autenticação

Autenticação centralizada baseada em JWT assinado.

As VPS clientes **não implementam login próprio**.

------------------------------------------------------------------------

# 🔁 Fluxo Completo de Autenticação (Online)

## 1️⃣ Acesso inicial

Usuário acessa:

https://cliente1.seusistema.com

O sistema verifica se há JWT válido. Caso não exista ou esteja expirado,
inicia o fluxo de autenticação.

------------------------------------------------------------------------

## 2️⃣ Redirecionamento para Auth Central

O sistema responde com:

HTTP 302 Redirect\
Location:
https://auth.seusistema.com/login?redirect=cliente1.seusistema.com

O navegador do usuário é redirecionado para o Auth Central.

------------------------------------------------------------------------

## 3️⃣ Login no Auth Central

O usuário:

-   Informa email e senha
-   O Auth valida credenciais
-   Identifica o tenant
-   Consulta o Billing Service

------------------------------------------------------------------------

## 4️⃣ Validação de Assinatura

O Billing verifica:

-   Assinatura ativa
-   Plano válido
-   Status de bloqueio
-   Período de tolerância

Se inválida → acesso negado\
Se válida → JWT é gerado

------------------------------------------------------------------------

## 5️⃣ Geração do JWT

Exemplo de payload:

{ "sub": "user_id", "tenant_id": "cliente1", "plan": "pro", "exp": "5-15
minutos", "iss": "auth.seusistema.com" }

O token é assinado com chave privada.

------------------------------------------------------------------------

## 6️⃣ Retorno ao Sistema do Cliente

O Auth redireciona o usuário de volta para:

https://cliente1.seusistema.com

Enviando o JWT via:

-   Cookie HTTP-only seguro (recomendado) ou
-   Authorization header

------------------------------------------------------------------------

## 7️⃣ Validação Local

A VPS do cliente:

-   Valida assinatura com chave pública
-   Valida expiração
-   Valida issuer

Se válido → acesso liberado\
Se inválido → 401 Unauthorized

------------------------------------------------------------------------

# 🔄 Renovação de Sessão

Quando o JWT expira:

1.  Próxima requisição retorna 401
2.  Frontend redireciona para Auth Central
3.  Novo JWT é emitido (se assinatura ativa)

------------------------------------------------------------------------

# 💰 Bloqueio por Inadimplência

Se o cliente parar de pagar:

-   O Auth não emite novo JWT
-   Quando o token expirar → acesso bloqueado
-   Bloqueio ocorre em poucos minutos

------------------------------------------------------------------------

# 🔒 Segurança

-   HTTPS obrigatório
-   JWT com expiração curta
-   Chave privada apenas no Auth
-   Chave pública distribuída para VPS clientes
-   Cookies HTTP-only e Secure
-   Rate limiting no Auth

------------------------------------------------------------------------

# 🚀 Vantagens

-   Autenticação centralizada
-   Controle financeiro forte
-   Bloqueio rápido
-   Escalabilidade
-   Preparado para MFA e SSO
-   Base sólida para evolução SaaS

------------------------------------------------------------------------

# 📌 Conclusão

Essa arquitetura separa claramente:

-   Domínio de Identidade
-   Domínio de Cobrança
-   Domínio de Negócio

Resultando em um sistema seguro, modular e escalável.
