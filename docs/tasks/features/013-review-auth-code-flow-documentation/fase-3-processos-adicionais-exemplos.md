# Fase 3 — Processos Adicionais, Exemplos de Código e Tabela de Erros

## Objetivo

Completar a documentação revisada com: processos adicionais (obtenção JWKS, validação JWT, logout), exemplos de código em múltiplas linguagens/stacks quando pertinente, e tabela consolidada de erros. Garantir que o documento seja o guia definitivo para desenvolvedores.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

Processos que podem estar faltando ou sub-explorados:
1. **Obtenção da chave pública (JWKS)** — já existe, mas pode ser expandido com URL completa, formato de resposta e exemplo de uso
2. **Validação do JWT no cliente** — já existe (validar assinatura, iss, aud, exp, nbf)
3. **Logout** — O ADR-010 menciona POST /api/v1/auth/logout. Verificar se está implementado. Se não, documentar como "futuro" ou remover.
4. **Health check** — GET /health pode ser útil para o cliente verificar disponibilidade antes de redirect
5. **Fluxo SPA vs fluxo tradicional** — diferenciar quando o cliente usa SPA (Accept: application/json) vs form (redirect 302)

### Ação 1.1

Consultar se existe endpoint de logout no backend. Grep por "logout" no código.

### Observação 1.1

Se logout não existir, documentar que o "logout local" é limpar tokens e redirecionar; logout global seria futuro.

---

### Pensamento 2

Exemplos de código: a doc atual tem exemplo em Go para o callback e para iniciar redirect. Pode ser útil adicionar:
- Exemplo em **JavaScript/TypeScript** (Angular ou fetch) para o fluxo SPA
- Exemplo em **Go** já existe — revisar e corrigir (ex.: evitar concatenação de string no JSON do code — usar json.Marshal)
- Exemplo de **validação JWT** em Go (usando biblioteca jwt)
- Exemplo de **interceptor** que trata 401 e chama refresh

### Ação 1.2

O exemplo Go atual usa `strings.NewReader(\`{"code":"`+code+`"}\`)` — vulnerável se code contiver aspas (improvável para hex). Melhor usar `json.Marshal` ou struct. Incluir na lista de melhorias do exemplo.

### Observação 1.2

Conformidade com code_guidelines backend: nunca concatenar strings para JSON; usar encoding/json.

---

### Pensamento 3

Tabela consolidada de erros: uma única tabela ou seções por endpoint? Considerando que cada endpoint tem erros específicos, manter tabelas por endpoint (login, token, refresh) e uma tabela resumo com HTTP → ação geral.

### Ação 1.3

Estrutura proposta:
- **Tabela geral:** HTTP status → ação recomendada (401→login, 402→renovação, 429→retry, 503→indisponível, etc.)
- **Tabela por endpoint:** para login, token e refresh — error_code, message, situação

### Observação 1.3

Facilita a vida do desenvolvedor ao implementar tratamento de erros.

---

### Pensamento 4

Outros processos a considerar:
- **Pré-requisitos de integração:** variáveis de ambiente (AUTH_SERVER_URL), certificado TLS para produção
- **Ordem de implementação sugerida:** 1) Redirect + state, 2) Callback + token exchange, 3) Armazenamento de token, 4) Validação JWT no backend do cliente, 5) Interceptor + refresh
- **Testes:** como testar o fluxo localmente (curl, dados de teste)
- **Rate limiting:** documentar que login, token e refresh estão sob rate limit; em 429, cabeçalho Retry-After

### Ação 1.4

Incluir seção "Ordem de implementação recomendada" e "Pré-requisitos" ampliados. Rate limiting já está mencionado — garantir que esteja em cada endpoint afetado.

### Observação 1.4

Reduz dúvidas de quem começa a integrar do zero.

---

### Pensamento 5

Conformidade com ADRs e code_guidelines: o documento não implementa código, mas deve refletir as decisões (RS256, 15 min access, 7 dias refresh, rotação, etc.). Revisar se há alguma inconsistência entre doc e ADRs.

### Ação 1.5

Cross-check: ADR-006 (kid, JWKS, claims), ADR-010 (fluxo, state, redirect_url interna, lockout, mensagem genérica). Garantir que a doc cite os ADRs nas seções relevantes.

### Observação 1.5

Mantém rastreabilidade e autoridade do documento.

---

## Decisão Final Fase 3

**Entregáveis para a documentação revisada:**

1. **Processos adicionais**
   - Obtenção JWKS: revisar e expandir se necessário
   - Validação JWT: checklist completo (sig, iss, aud, exp, nbf, kid)
   - Logout: documentar logout local (limpar tokens, redirect); mencionar logout global como futuro se não implementado
   - Ordem de implementação recomendada para o desenvolvedor

2. **Exemplos de código**
   - Exemplo Go callback: corrigir para usar `json.Marshal` ao montar body do POST /auth/token
   - Exemplo Go redirect: manter/corrigir
   - Exemplo TypeScript/JavaScript: fluxo SPA (fetch para login, tratamento de redirect_url)
   - Exemplo de validação JWT em Go (pseudo-código ou referência a lib)
   - Exemplo de interceptor que trata 401 e chama refresh (conceitual)

3. **Tabela consolidada de erros**
   - Tabela geral: HTTP → ação
   - Tabelas por endpoint: login, token, refresh (error_code, message, quando ocorre)

4. **Correções finais**
   - Garantir formato JSON de erro em toda a doc (error_code)
   - Revisar diagrama de fluxo (parâmetros tenant_slug, product_slug)
   - Links para ADRs e docs/backend quando relevante

---

## Checklist de Implementação (Documentação)

1. [ ] Revisar/expandir seção JWKS
2. [ ] Revisar seção Validação JWT
3. [ ] Documentar logout (local e futuro)
4. [ ] Adicionar "Ordem de implementação recomendada"
5. [ ] Corrigir exemplo Go (json.Marshal)
6. [ ] Adicionar exemplo TypeScript/JavaScript (SPA)
7. [ ] Adicionar exemplo interceptor (conceitual)
8. [ ] Criar tabela consolidada de erros
9. [ ] Revisão final: diagrama, links, consistência

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docs/code_guidelines/backend.md | Exemplo Go sem concatenação de JSON |
| ADR-006 | Validação JWT completa |
| ADR-010 | Fluxo, logout, rate limiting |

---

## Referências

- [auth-code-flow-integration.md](../../../integration/auth-code-flow-integration.md)
- [ADR-006](../../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [ADR-010](../../../adrs/ADR-010-fluxo-centralizado-de-autenticacao.md)
- [docs/backend/README.md](../../../backend/README.md)
