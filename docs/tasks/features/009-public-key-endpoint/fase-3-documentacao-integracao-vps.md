# Fase 3 — Documentação: Integração Clientes e Configurações VPS/NGINX

## Objetivo

Documentar para as aplicações clientes como obter a chave pública para validação de JWTs e garantir que as configurações necessárias na VPS (NGINX) estejam documentadas para que o endpoint funcione conforme ADRs.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O documento `docs/integration/auth-code-flow-integration.md` já menciona na seção "Pré-requisitos" que o cliente precisa da "Chave pública do auth para validar a assinatura RS256 do JWT (endpoint JWKS ou arquivo)". Porém, não há instruções detalhadas sobre a URL do endpoint, formato da resposta, como usar o kid, ou como fazer cache. É necessário adicionar uma seção específica sobre o JWKS.

### Ação 3.1

Adicionar ou expandir seção em `docs/integration/auth-code-flow-integration.md`:

**Título:** "Obtenção da Chave Pública (JWKS)"

**Conteúdo:**
- URL do endpoint: `GET {AUTH_SERVER_URL}/.well-known/jwks.json`
- Exemplo: `https://auth.wisa-crm.com/.well-known/jwks.json`
- Resposta: JSON com `keys` array; cada chave tem `kid`, `kty`, `use`, `alg`, `n`, `e`
- O cliente deve usar o `kid` do header do JWT para selecionar a chave correta no JWKS
- Cache recomendado: 24 horas (Cache-Control: max-age=86400) — o cliente pode cachear localmente
- Em rotación de chaves, o JWKS pode conter múltiplas chaves; tokens antigos (kid antigo) continuam válidos até expirar

### Observação 3.1

O `docs/context.md` e o fluxo de validação do JWT (docs/context.md linha 101) dizem que o cliente deve "Validar a assinatura criptográfica usando a chave pública do wisa-crm-service". O documento de integração deve explicar como obter essa chave via JWKS.

### Pensamento 2

O `docs/backend/README.md` lista os endpoints. Adicionar:

- **GET /.well-known/jwks.json** — Retorna chaves públicas em formato JWKS para validação de assinatura RS256 dos JWTs. Não exige autenticação. Cache 24h.

### Ação 3.2

Adicionar entrada na seção de endpoints (ou criar seção "Endpoints de Descoberta") em docs/backend/README.md.

### Observação 3.2

O health e o JWKS são endpoints que não fazem parte da API de auth stricto sensu — são utilitários/descoberta.

### Pensamento 3

O requisito do usuário especifica: "Caso alguma configuração adicional seja necessária na VPS para que o código funcione conforme descrito nas ADRs, adicionar a documentação em: docs/vps/configurations.md"

A ADR-007 já documenta o `location /.well-known/jwks.json` no NGINX. O arquivo `docs/backend/vps-configurations.md` existe e contém configs de PostgreSQL, env, migrations, PgBouncer. O usuário pediu `docs/vps/configurations.md` — path diferente. Criar o diretório `docs/vps/` e o arquivo `configurations.md` com as configs de NGINX específicas para o JWKS e outros itens que possam faltar. Ou adicionar ao `docs/backend/vps-configurations.md` uma seção de NGINX. A ADR-007 é extensa e contém a config completa. O vps-configurations.md em docs/backend foca em passos operacionais (comandos, paths). O NGINX é parte da infraestrutura — faz sentido ter em docs/vps ou em docs/backend. Seguir literalmente o pedido do usuário: criar `docs/vps/configurations.md`.

### Ação 3.3

Criar `docs/vps/configurations.md` com:
- Referência à ADR-007 para a config completa do NGINX
- Seção específica: "Endpoint JWKS (Chave Pública)" com o bloco `location /.well-known/jwks.json` que deve estar no site config
- Passos para adicionar o location se não existir
- Rate limiting: zona jwks_zone (60 req/min por IP, burst 10) conforme ADR-007
- Verificação: curl https://auth.dominio.com/.well-known/jwks.json deve retornar 200

### Observação 3.3

O `docs/backend/vps-configurations.md` pode fazer referência ao `docs/vps/configurations.md` para detalhes de NGINX, ou vice-versa. Para evitar duplicação, o `docs/vps/configurations.md` pode ser o documento central de configs de infraestrutura (NGINX, firewall, etc.), enquanto `docs/backend/vps-configurations.md` foca em backend (DB, env, migrations). Criar docs/vps/configurations.md com a seção NGINX para o JWKS.

### Pensamento 4

A ADR-007 menciona a zona `jwks_zone` na tabela de rate limiting, mas a "Configuração Completa de Referência" no final do ADR não inclui o location do jwks. O docs/vps/configurations.md deve ter o bloco completo para copiar e colar.

### Ação 3.4

Incluir no docs/vps/configurations.md:

```markdown
## NGINX — Endpoint JWKS (Chave Pública)

O endpoint `/.well-known/jwks.json` distribui a chave pública para validação de JWTs. 
Deve estar configurado no server block HTTPS.

### location block

Adicionar dentro do `server { listen 443 ssl http2; ... }`:

```nginx
# JWKS — rate limiting conforme ADR-007 (60 req/min por IP)
limit_req_zone $binary_remote_addr zone=jwks_zone:10m rate=60r/m;

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

A zona `jwks_zone` deve ser declarada no contexto http (fora do server block).
```

### Observação 3.4

A ADR-007 declara as zonas no início (limit_req_zone). O location está dentro do server. A documentação deve deixar claro onde cada bloco vai.

---

## Checklist de Implementação

- [ ] 1. Adicionar/expandir seção "Obtenção da Chave Pública (JWKS)" em docs/integration/auth-code-flow-integration.md
- [ ] 2. Documentar URL, formato, uso do kid, cache de 24h
- [ ] 3. Adicionar GET /.well-known/jwks.json em docs/backend/README.md
- [ ] 4. Criar docs/vps/configurations.md (ou docs/vps/ se não existir)
- [ ] 5. Incluir seção NGINX para JWKS com location block e rate limiting
- [ ] 6. Referenciar ADR-007 para config completa

---

## Referências

- ADR-006 — JWKS, distribuição de chave pública
- ADR-007 — NGINX, location /.well-known/jwks.json, jwks_zone
- docs/integration/auth-code-flow-integration.md
- docs/backend/README.md
- docs/backend/vps-configurations.md
