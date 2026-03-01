# ADR-007 — NGINX como Reverse Proxy e API Gateway

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (infraestrutura de rede)

---

## Contexto

O `wisa-crm-service` é exposto publicamente na internet como ponto de autenticação centralizado para múltiplos sistemas clientes. Toda requisição de login, renovação de token e consulta de chave pública passa por este serviço antes de atingir os sistemas do cliente.

Isso cria um perfil de risco específico:

- O endpoint de login é um **alvo natural para ataques de força bruta e credential stuffing**
- A superfície de ataque inclui DDoS, SSRF, directory traversal e HTTP smuggling
- O serviço precisa suportar **HTTPS obrigatório** com TLS configurado corretamente
- O roteamento entre o frontend Angular (arquivos estáticos) e o backend Go (API REST) precisa ser gerenciado
- Múltiplos domínios podem precisar de tratamento diferenciado (ex: `auth.wisa-crm.com` para o portal de login, `api.wisa-crm.com` para a API)

O componente de rede que fica entre a internet e a aplicação é a **primeira linha de defesa** e precisa ser configurado com rigor.

---

## Decisão

**NGINX será utilizado como reverse proxy, terminação TLS, servidor de arquivos estáticos e primeira camada de segurança HTTP do `wisa-crm-service`.**

A configuração do NGINX implementará:
- Terminação TLS com certificados Let's Encrypt (Certbot)
- Headers de segurança HTTP obrigatórios
- Rate limiting por IP para endpoints de autenticação
- Proxy reverso para o backend Go (porta interna não exposta ao público)
- Serving de arquivos estáticos do Angular com cache otimizado

---

## Justificativa

### 1. Terminação TLS fora da aplicação

Gerenciar TLS no NGINX ao invés de no backend Go oferece:

- **Renovação automática de certificados** via Certbot/Let's Encrypt sem reiniciar o backend
- **Configuração centralizada de cipher suites** — uma mudança no NGINX afeta todo o tráfego
- **OCSP Stapling** para validação de revogação de certificados sem latência adicional
- **HTTP/2** habilitado transparentemente para o frontend Angular, melhorando performance de carregamento

A aplicação Go escuta em `127.0.0.1:8080` (loopback apenas), nunca exposta diretamente.

### 2. Rate limiting como primeira linha de defesa

O NGINX implementa rate limiting no nível de rede, **antes de qualquer código da aplicação ser executado**:

```nginx
# Zona de rate limiting: 1 requisição/segundo por IP no endpoint de login
limit_req_zone $binary_remote_addr zone=login_zone:10m rate=5r/m;

location /api/v1/auth/login {
    limit_req zone=login_zone burst=3 nodelay;
    limit_req_status 429;
    # ...
}
```

Isso significa que um ataque de força bruta exige muito mais recursos do atacante do que do sistema, e não desperdiça goroutines ou conexões de banco para requisições que seriam rejeitadas de qualquer forma.

**Zonas de rate limiting propostas:**

| Zona | Endpoint | Limite | Burst |
|------|----------|--------|-------|
| `login_zone` | `/api/v1/auth/login` | 5 req/min por IP | 3 |
| `refresh_zone` | `/api/v1/auth/refresh` | 30 req/min por IP | 5 |
| `jwks_zone` | `/.well-known/jwks.json` | 60 req/min por IP | 10 |
| `api_zone` | `/api/v1/*` | 100 req/min por IP | 20 |

### 3. Headers de segurança HTTP

O NGINX adiciona headers de segurança em todas as respostas, independentemente do que o backend Go retorne:

```nginx
# Headers de segurança obrigatórios
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
add_header X-Frame-Options "DENY" always;
add_header X-Content-Type-Options "nosniff" always;
add_header X-XSS-Protection "1; mode=block" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "camera=(), microphone=(), geolocation=()" always;
add_header Content-Security-Policy "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none';" always;
```

**HSTS com `preload`** garante que navegadores nunca tentem HTTP, mesmo na primeira visita (após inclusão na lista HSTS preload).

### 4. Configuração TLS hardening

```nginx
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305;
ssl_prefer_server_ciphers off;  # Deixar o cliente escolher (TLS 1.3 recomendado)

ssl_session_cache shared:SSL:10m;
ssl_session_timeout 1d;
ssl_session_tickets off;         # Desabilitar session tickets para forward secrecy

ssl_stapling on;                 # OCSP Stapling
ssl_stapling_verify on;
resolver 8.8.8.8 1.1.1.1 valid=300s;
```

TLS 1.0 e 1.1 desabilitados (vulneráveis a POODLE, BEAST, etc.). Apenas TLS 1.2 e 1.3 são aceitos.

### 5. Separação de roteamento: estático vs API

```nginx
server {
    listen 443 ssl http2;
    server_name auth.wisa-crm.com;

    # Arquivos estáticos do Angular (SPA)
    location / {
        root /var/www/wisa-crm/browser;
        try_files $uri $uri/ /index.html;  # SPA routing
        
        # Cache para assets com hash no nome (Angular production build)
        location ~* \.(js|css|woff2|png|ico)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # API Backend Go
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_read_timeout 30s;
        proxy_connect_timeout 5s;
        proxy_send_timeout 30s;
    }
    
    # JWKS endpoint — cache agressivo no cliente
    location /.well-known/jwks.json {
        proxy_pass http://127.0.0.1:8080;
        add_header Cache-Control "public, max-age=86400";  # Cache 24h
    }
}

# Redirect HTTP → HTTPS
server {
    listen 80;
    server_name auth.wisa-crm.com;
    return 301 https://$server_name$request_uri;
}
```

### 6. Proteção contra ataques comuns via NGINX

```nginx
# Limitar tamanho de request body (previne upload de payload gigante)
client_max_body_size 1m;

# Timeout para prevenir slow HTTP attacks (Slowloris)
client_body_timeout 10s;
client_header_timeout 10s;
keepalive_timeout 65s;
send_timeout 10s;

# Ocultar versão do NGINX
server_tokens off;

# Bloquear métodos HTTP não utilizados
if ($request_method !~ ^(GET|POST|OPTIONS)$) {
    return 405;
}

# Bloquear User-Agents suspeitos
if ($http_user_agent ~* (nikto|sqlmap|nmap|masscan|dirbuster)) {
    return 403;
}
```

---

## Consequências

### Positivas
- TLS terminado e configurado de forma centralizada e segura
- Rate limiting protege o endpoint de login contra força bruta antes de consumir recursos da aplicação
- Headers de segurança HTTP garantidos em todas as respostas, mesmo em casos de bug no backend
- Backend Go não é diretamente exposto à internet (escuta apenas em loopback)
- Separação clara de responsabilidades: NGINX cuida de rede/HTTP, Go cuida de lógica de negócio

### Negativas
- NGINX é um único ponto de falha na frente da aplicação — processo de alta disponibilidade requer atenção
- Configuração incorreta do NGINX pode introduzir vulnerabilidades de segurança HTTP
- Logs do NGINX e do backend Go precisam ser correlacionados para debugging (adicionar `X-Request-ID`)
- Atualizações de segurança do NGINX precisam ser aplicadas com agilidade

---

## Riscos

| Risco | Probabilidade | Impacto | Severidade |
|-------|--------------|---------|-----------|
| Certificado TLS expirado causando indisponibilidade | Média | Crítico | Alta |
| Misconfiguration de CORS expondo endpoints da API | Média | Alto | Alta |
| HTTP Request Smuggling entre NGINX e backend Go | Baixa | Alto | Média |
| Bypass de rate limiting via IP rotation (botnets) | Alta | Médio | Média |
| Versão NGINX vulnerável sem patch | Média | Alto | Alta |
| Buffer overflow em módulos NGINX customizados | Baixa | Alto | Média |

---

## Mitigações

### Certificado TLS expirado
- Configurar Certbot com renovação automática: `certbot renew --pre-hook "systemctl stop nginx" --post-hook "systemctl start nginx"` ou, melhor, com reload: `--deploy-hook "nginx -s reload"`
- Monitorar validade do certificado com alerta 30 dias antes da expiração (Prometheus + `ssl_certificate_expiry`)
- Testar renovação automática mensalmente

### CORS
- Configurar CORS explicitamente, **nunca** usar `Access-Control-Allow-Origin: *` na API de autenticação
- Manter lista branca de origens permitidas:
  ```nginx
  set $cors_origin "";
  if ($http_origin ~* "^https://(cliente1|cliente2)\.seusistema\.com$") {
      set $cors_origin $http_origin;
  }
  add_header Access-Control-Allow-Origin $cors_origin always;
  ```

### HTTP Request Smuggling
- Usar `proxy_http_version 1.1` e `proxy_set_header Connection ""`
- Configurar `Content-Length` estrito no backend Go
- Desabilitar `chunked_transfer_encoding` se não necessário

### Bypass de rate limiting
- Combinar rate limiting no NGINX com rate limiting baseado em conta no backend Go (por email tentado)
- Implementar CAPTCHA após N falhas consecutivas de login
- Monitorar padrões de login suspeitos e bloquear ASNs de botnets conhecidos via `fail2ban`

### Versão NGINX vulnerável
- Configurar atualizações automáticas de segurança (unattended-upgrades) para o pacote NGINX
- Subscrever ao feed de segurança do NGINX para CVEs
- Usar NGINX mainline ou stable com suporte ativo

---

## Configuração Completa de Referência

```nginx
# /etc/nginx/sites-available/wisa-crm

# Zones de rate limiting
limit_req_zone $binary_remote_addr zone=login_zone:10m rate=5r/m;
limit_req_zone $binary_remote_addr zone=api_zone:10m rate=100r/m;
limit_conn_zone $binary_remote_addr zone=conn_zone:10m;

server {
    listen 443 ssl http2;
    server_name auth.wisa-crm.com;
    
    # TLS
    ssl_certificate /etc/letsencrypt/live/auth.wisa-crm.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/auth.wisa-crm.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:...;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;
    ssl_stapling on;
    ssl_stapling_verify on;
    
    # Limites gerais
    client_max_body_size 1m;
    client_body_timeout 10s;
    client_header_timeout 10s;
    server_tokens off;
    
    # Conexões simultâneas por IP
    limit_conn conn_zone 20;
    
    # Headers de segurança
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; frame-ancestors 'none';" always;
    
    # SPA Angular
    location / {
        root /var/www/wisa-crm/browser;
        try_files $uri $uri/ /index.html;
    }
    
    # API com rate limiting
    location /api/v1/auth/login {
        limit_req zone=login_zone burst=3 nodelay;
        limit_req_status 429;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto https;
        proxy_read_timeout 30s;
    }
    
    location /api/ {
        limit_req zone=api_zone burst=20 nodelay;
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-Proto https;
        proxy_read_timeout 30s;
    }
}

server {
    listen 80;
    server_name auth.wisa-crm.com;
    return 301 https://$server_name$request_uri;
}
```

---

## Alternativas Consideradas

### Caddy
- **Prós:** HTTPS automático por padrão (sem Certbot separado), configuração mais simples (Caddyfile), HTTP/3 nativo
- **Contras:** Menos adotado em produção em ambientes enterprise, documentação de casos extremos menos abrangente, menor flexibilidade de módulos que o NGINX

### Traefik
- **Prós:** Excelente para ambientes containerizados (Docker/Kubernetes), auto-discovery de serviços
- **Contras:** Overhead desnecessário para uma VPS single-server, configuração mais complexa sem containers, menor performance bruta que NGINX para serving de arquivos estáticos

### HAProxy
- **Prós:** Excelente como load balancer, alta performance para TCP/HTTP proxy
- **Contras:** Não serve arquivos estáticos, configuração de TLS menos intuitiva, sem módulos de segurança HTTP nativos comparáveis ao NGINX

**NGINX é a escolha mais madura, performática e amplamente documentada para o perfil deste sistema.**

---

## Referências

- [NGINX Security Controls](https://docs.nginx.com/nginx/admin-guide/security-controls/)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)
- [NGINX Rate Limiting](https://www.nginx.com/blog/rate-limiting-nginx/)
- [Certbot Documentation](https://certbot.eff.org/docs/)
- [OWASP NGINX Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/NGINX_Security_Cheat_Sheet.html)
