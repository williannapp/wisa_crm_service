# Configurações VPS — Infraestrutura

Este documento complementa as ADRs com configurações operacionais para a VPS onde o wisa-crm-service é implantado. Para configs específicas do backend (PostgreSQL, variáveis de ambiente), ver [docs/backend/vps-configurations.md](../backend/vps-configurations.md).

---

## 1. NGINX — Endpoint JWKS (Chave Pública)

O endpoint `/.well-known/jwks.json` distribui a chave pública RSA para que os sistemas clientes validem a assinatura dos JWTs emitidos. Deve estar configurado no server block HTTPS do NGINX.

**Referência:** [ADR-007 — NGINX como Reverse Proxy](../adrs/ADR-007-nginx-como-reverse-proxy.md)

### 1.1 Zona de Rate Limiting

Declarar no contexto `http` do NGINX (fora do bloco `server`):

```nginx
# Zona para endpoint JWKS — 60 requisições/minuto por IP, burst 10
limit_req_zone $binary_remote_addr zone=jwks_zone:10m rate=60r/m;
```

### 1.2 Location Block

Adicionar dentro do `server { listen 443 ssl http2; ... }`:

```nginx
location /.well-known/jwks.json {
    limit_req zone=jwks_zone burst=10 nodelay;
    limit_req_status 429;
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
    add_header Cache-Control "public, max-age=86400";
}
```

### 1.3 Verificação

Após aplicar a configuração e recarregar o NGINX (`nginx -t && systemctl reload nginx`):

```bash
curl -s https://auth.seu-dominio.com/.well-known/jwks.json | jq .
```

Deve retornar HTTP 200 com JSON no formato:

```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "alg": "RS256",
      "kid": "key-2026-v1",
      "n": "...",
      "e": "AQAB"
    }
  ]
}
```

---

## 2. Config NGINX para ambiente Docker (testes)

Para testes de integração E2E em Docker, a pasta `nginx/` na raiz do projeto contém configuração específica que usa **nomes de serviço** (backend, frontend, test-app-backend, test-app-frontend) em vez de `127.0.0.1`. Em VPS, a config seria adaptada com `127.0.0.1` e paths em disco. Ver [nginx/README.md](../../nginx/README.md).

---

## 3. Referências

- [ADR-006 — JWT com Assinatura Assimétrica](../adrs/ADR-006-jwt-com-assinatura-assimetrica.md)
- [ADR-007 — NGINX como Reverse Proxy](../adrs/ADR-007-nginx-como-reverse-proxy.md)
- [ADR-009 — Infraestrutura VPS Linux](../adrs/ADR-009-infraestrutura-vps-linux.md)
- [docs/backend/vps-configurations.md](../backend/vps-configurations.md)
- [nginx/README.md](../../nginx/README.md) — Config para Docker (testes locais)
