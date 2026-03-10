# Fase 1 — Configuração NGINX e Roteamento

## Objetivo

Criar a pasta `nginx/` na raiz do projeto com a estrutura de arquivos de configuração do NGINX e definir o roteamento para os subdomínios `auth.wisa.labs.com.br` e `lingerie-maria.wisa.labs.com.br`, com revisão das portas e regras conforme o ambiente Docker.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O projeto utiliza Docker Compose com os serviços: `backend` (porta 8080), `frontend` (porta 80 internamente, exposta como 4200), `test-app-backend` (8081), `test-app-frontend` (porta 80 internamente, exposta como 4201). Em ambiente Docker, a comunicação entre containers usa **nomes de serviço** na rede interna, não `0.0.0.0` nem portas do host. O NGINX rodará como outro serviço e fará proxy_pass para os serviços por nome.

### Ação 1.1

Mapear os serviços e portas corretas para proxy_pass:

| Destino | Serviço Docker | Porta interna |
|---------|----------------|---------------|
| Frontend Angular (auth) | `frontend` | 80 |
| Backend Go (auth) | `backend` | 8080 |
| Test-app frontend | `test-app-frontend` | 80 |
| Test-app backend | `test-app-backend` | 8081 |

**Observação 1.1:** A porta 4200/4201 é a exposição no host; dentro da rede Docker os serviços escutam em 80 (frontends) e 8080/8081 (backends). O proxy_pass deve usar `http://frontend:80`, `http://backend:8080`, etc.

---

### Pensamento 2

Para o subdomínio `auth.wisa.labs.com.br`, a configuração do usuário especifica:
- `/` → SPA Angular (frontend)
- `/api/` → Backend Go
- `/.well-known/` → Backend Go (endpoint JWKS)

O backend expõe `/.well-known/jwks.json` (conforme main.go e docs). A location `/.well-known/` cobre esse path. Porém, em NGINX a ordem de avaliação importa: locations mais específicas devem vir antes. A ordem recomendada: `/.well-known/` → `/api/` → `/` (fallback para SPA).

### Ação 1.2

Definir o server block para `auth.wisa.labs.com.br`:

```nginx
server {
    listen 80;
    server_name auth.wisa.labs.com.br;

    # JWKS (chave pública) — mais específico primeiro
    location /.well-known/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API — Backend Go
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Frontend — SPA Angular
    location / {
        proxy_pass http://frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Observação 1.2:** Uso de `backend:8080` e `frontend:80` garante resolução via rede Docker quando todos os serviços estiverem no mesmo compose. Headers X-Forwarded-* são essenciais para o backend saber o esquema (HTTP/HTTPS) e IP real do cliente (ADR-007, OWASP).

---

### Pensamento 3

Para `lingerie-maria.wisa.labs.com.br`, a config original só possui `location /gestao-pocket` apontando para a test-app frontend. Porém, a test-app possui backend com endpoints `/api/`, `/login`, `/callback`. O frontend Angular faz requisições relativas (ex.: `/api/hello`). Se o usuário acessa `https://lingerie-maria.wisa.labs.com.br/gestao-pocket`, o fetch para `/api/hello` será para `https://lingerie-maria.wisa.labs.com.br/api/hello` (mesmo origin). Portanto, é necessário rotear também `/api/`, `/login` e `/callback` para o `test-app-backend`.

### Ação 1.3

Definir o server block para `lingerie-maria.wisa.labs.com.br`. A test-app possui frontend (SPA) e backend (API: /api/, /login, /callback). O Angular é buildado com base href `/`; portanto, para funcionar sem alterar o frontend, a aplicação deve ser servida na **raiz** do subdomínio. O requisito original mencionava `location /gestao-pocket`; para ambiente de **testes** prioriza-se configuração que funcione sem mudanças na test-app.

**Opção recomendada (test-app na raiz):**

```nginx
server {
    listen 80;
    server_name lingerie-maria.wisa.labs.com.br;

    # API da test-app
    location /api/ {
        proxy_pass http://test-app-backend:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Fluxo de autenticação
    location /login {
        proxy_pass http://test-app-backend:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /callback {
        proxy_pass http://test-app-backend:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # /:product/callback (ex.: /gestao-pocket/callback)
    location ~ ^/[^/]+/callback$ {
        proxy_pass http://test-app-backend:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # SPA da test-app na raiz
    location / {
        proxy_pass http://test-app-frontend:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

**Opção alternativa (com /gestao-pocket):** Se for necessário servir a test-app em `/gestao-pocket`, a test-app precisará ser buildada com `--base-href /gestao-pocket/` (alteração em test-app/frontend, fora do escopo atual). A location seria `location /gestao-pocket` e `location /gestao-pocket/` com `proxy_pass http://test-app-frontend:80/`.

**Observação 1.3:** Para atingir o objetivo de "executar testes" sem alterar backend/frontend, a opção recomendada serve a test-app na raiz de `lingerie-maria.wisa.labs.com.br`, garantindo que assets (JS, CSS) e rotas do Angular funcionem corretamente.

---

### Pensamento 4

Estrutura de arquivos na pasta `nginx/`:
- `nginx.conf` — arquivo principal (ou usa default do container)
- `conf.d/default.conf` — server blocks (include no nginx.conf)
- Ou: um único arquivo `conf.d/wisa.conf` com os dois server blocks

O NGINX oficial no Docker usa `etc/nginx/nginx.conf` que inclui `conf.d/*.conf`. Podemos colocar nosso config em `conf.d/default.conf` ou `conf.d/wisa.conf`.

### Ação 1.4

Definir estrutura de arquivos:

```
nginx/
├── conf.d/
│   └── wisa.conf      # Server blocks para auth e lingerie-maria
└── nginx.conf         # (opcional) Custom main config; ou usar default do image
```

Para simplicidade, usar apenas `conf.d/wisa.conf` montado no container; o nginx.conf padrão da imagem já inclui `include /etc/nginx/conf.d/*.conf`.

**Observação 1.4:** A imagem oficial `nginx:alpine` já possui a estrutura; basta montar nosso `conf.d/wisa.conf` em `/etc/nginx/conf.d/wisa.conf` ou substituir o default. O arquivo não deve ser nomeado `default.conf` para não conflitar com o exemplo da imagem; usar `wisa.conf` é mais explícito.

---

### Pensamento 6

Resolução de nomes: no Docker Compose, os serviços se encontram pelo nome. O NGINX precisa estar na **mesma rede** que backend, frontend, test-app-backend, test-app-frontend. Ao adicionar o serviço nginx ao docker-compose principal, ele herdará a rede default e poderá resolver `backend`, `frontend`, etc.

Para `/etc/hosts` em testes locais, o usuário precisa adicionar:
```
127.0.0.1 auth.wisa.labs.com.br
127.0.0.1 lingerie-maria.wisa.labs.com.br
```

Documentar isso no README da pasta nginx ou no doc da fase.

### Ação 1.5

Incluir no documento da Fase 1 uma seção "Pré-requisitos para testes locais" com instrução de adicionar os domínios ao `/etc/hosts`.

---

## Resumo das Ações da Fase 1

| # | Ação | Resultado esperado |
|---|------|-------------------|
| 1.1 | Mapear serviços/portas | Tabela de referência backend:8080, frontend:80, test-app-* |
| 1.2 | Server block auth.wisa.labs.com.br | Locations /.well-known, /api, / com proxy aos serviços corretos |
| 1.3 | Server block lingerie-maria.wisa.labs.com.br | Locations para /api, /login, /callback, /:product/callback e / (SPA na raiz) |
| 1.4 | Estrutura nginx/conf.d/wisa.conf | Arquivo de configuração pronto para montagem no container |
| 1.5 | Documentar /etc/hosts | Pré-requisito para testes locais com subdomínios |

---

## Checklist de Implementação

- [ ] Criar pasta `nginx/`
- [ ] Criar pasta `nginx/conf.d/`
- [ ] Criar arquivo `nginx/conf.d/wisa.conf` com os dois server blocks
- [ ] Revisar que todas as portas usam nomes de serviço Docker (não 0.0.0.0)
- [ ] Garantir headers proxy_set_header em todas as locations
- [ ] Documentar em `nginx/README.md` o pré-requisito de /etc/hosts para testes locais

---

## Riscos e Mitigações

| Risco | Mitigação |
|-------|-----------|
| Nomes de serviço não resolvidos | NGINX no mesmo docker-compose que os outros serviços |
| CORS em chamadas cross-origin | Auth e test-app em subdomínios diferentes; verificar se CORS está configurado no backend quando necessário |
| Base path /gestao-pocket quebra assets | Config recomendada serve test-app na raiz do subdomínio |

---

## Conformidade

- **ADR-007:** Proxy reverso com headers X-Forwarded-*, estrutura de locations
- **ADR-009:** Config compatível com ambiente de desenvolvimento local (porta 80)
- **Code guidelines:** N/A (apenas configuração de infraestrutura)
