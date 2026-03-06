# Feature 009 — Public Key Endpoint (JWKS) — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | JWKS Provider: extrair chave pública RSA e formatar em JWKS (suporte a múltiplas chaves) | 1/1 | Concluída |
| Fase 2 | Endpoint GET /.well-known/jwks.json: handler, rota pública, Cache-Control | 1/1 | Concluída |
| Fase 3 | Documentação: integração clientes e configurações VPS/NGINX | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-jwks-provider-extrair-chave-publica.md](./fase-1-jwks-provider-extrair-chave-publica.md)
- [fase-2-endpoint-publico-jwks.md](./fase-2-endpoint-publico-jwks.md)
- [fase-3-documentacao-integracao-vps.md](./fase-3-documentacao-integracao-vps.md)

## Resumo das Tasks

- [x] Fase 1 — Criar JWKSProvider/JWKSService que extrai chave pública do par RSA e formata em RFC 7517 (kty, use, alg, kid, n, e); suportar múltiplas chaves
- [x] Fase 2 — Handler GET /.well-known/jwks.json sem autenticação; Content-Type application/json; Cache-Control: public, max-age=86400; kid alinhado com JWT
- [x] Fase 3 — Documentar em docs/integration como clientes obtêm a chave; configurar NGINX em docs/vps/configurations.md se necessário

## Dependências entre Fases

| Fase | Dependências |
|------|--------------|
| Fase 1 | — |
| Fase 2 | Fase 1 |
| Fase 3 | Fase 2 |

## Ordem Sugerida de Implementação

1. Fase 1 — JWKS Provider (extração e formatação)
2. Fase 2 — Endpoint público e integração no router
3. Fase 3 — Documentação

## Notas Importantes

- **Path:** `/.well-known/jwks.json` (padrão RFC 7517, conforme ADR-006 e ADR-007)
- **HTTPS:** Garantido pela terminação TLS no NGINX (backend escuta em loopback HTTP)
- **kid:** Deve corresponder exatamente ao `kid` no header dos JWTs emitidos (JWT_KEY_ID)
- **Múltiplas chaves:** Estrutura `{"keys": [...]}` permite rotação sem downtime

## Referências

- [docs/context.md](../../context.md)
- [ADR-006 — JWT com Assinatura Assimétrica](../../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [ADR-007 — NGINX como Reverse Proxy](../../adrs/ADR-007-nginx-como-reverse-proxy.md)
- [docs/code_guidelines/backend.md](../../code_guidelines/backend.md)
- [RFC 7517 — JSON Web Key (JWK)](https://tools.ietf.org/html/rfc7517)
