# Fase 4 — Criar Dockerfile para o Frontend

## Objetivo

Criar um Dockerfile para construir e executar o frontend Angular em container, permitindo deploy consistente e uso em ambientes de desenvolvimento com Docker. O frontend será servido como arquivos estáticos via NGINX, conforme ADR-007.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-007 define que o NGINX serve os arquivos estáticos do Angular em produção (SPA). Em um ambiente Docker Compose para desenvolvimento, o frontend pode ser servido de duas formas:

A) Container com NGINX servindo os arquivos estáticos (build do Angular)
B) Container com `ng serve` para desenvolvimento com hot reload

Para a estrutura inicial e alinhamento com o ADR-007, a abordagem de **produção** é preferível: build do Angular e serving via NGINX. Para desenvolvimento local com hot reload, o desenvolvedor pode rodar `ng serve` fora do Docker.

### Ação 4.1

O Dockerfile deve produzir uma imagem que:
1. **Stage 1 (builder):** Compila o projeto Angular com `ng build --configuration production`
2. **Stage 2 (runtime):** Serve os arquivos estáticos com NGINX

### Observação 4.1

Build multi-stage reduz o tamanho final da imagem — a imagem de runtime não contém Node.js, apenas NGINX e os arquivos estáticos. Isso diminui a superfície de ataque e acelera o deploy.

---

### Pensamento 2

Posicionamento do Dockerfile: seguir a convenção do backend (`backend/docker/backend/Dockerfile`). Para o frontend, opções:
- `frontend/Dockerfile` — na raiz do frontend (simples)
- `frontend/docker/Dockerfile` — em subpasta (consistente com backend)

O backend usa `backend/docker/backend/Dockerfile` e `backend/docker/database/Dockerfile`. Para o frontend, um único Dockerfile em `frontend/Dockerfile` é suficiente e é a convenção mais comum para projetos Angular.

### Ação 4.2

Posicionar o Dockerfile em `frontend/Dockerfile`. O contexto de build será `frontend/` (ou a raiz com `-f frontend/Dockerfile`). O `COPY` e `WORKDIR` serão relativos ao contexto.

### Observação 4.2

Para `docker build -f frontend/Dockerfile .`, o contexto é a raiz. Os caminhos no Dockerfile devem considerar isso: `COPY frontend/package*.json ./` se o contexto for raiz. Se o contexto for `frontend/`, usa-se `COPY package*.json ./`. Recomendação: contexto `frontend/` com `docker build -t wisa-crm-frontend ./frontend`.

---

### Pensamento 3

Estrutura do Dockerfile multi-stage:

**Stage 1 — Builder:**
- Base: `node:20-alpine` (ou `node:20-bookworm-slim` para compatibilidade)
- `WORKDIR /app`
- `COPY package*.json .`
- `RUN npm ci`
- `COPY . .`
- `RUN npm run build` (ou `ng build --configuration production`)

O output do build Angular (por padrão) fica em `dist/frontend/browser/` (ou `dist/<project-name>/browser/`). Verificar o `angular.json` para o path exato do `outputPath`.

**Stage 2 — Runtime:**
- Base: `nginx:alpine`
- Copiar apenas os arquivos estáticos do stage builder para `/usr/share/nginx/html`
- Configurar NGINX para SPA (try_files para fallback em index.html)
- Expor porta 80

### Ação 4.3

Estrutura proposta:

```dockerfile
# Stage 1: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: Runtime
FROM nginx:alpine
COPY --from=builder /app/dist/frontend/browser /usr/share/nginx/html
# Copiar configuração customizada do NGINX para SPA routing
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Observação 4.3

O path `dist/frontend/browser` assume que o projeto Angular se chama "frontend" e o outputPath no `angular.json` é `dist/frontend/browser`. O Angular 17+ usa `browser` como subpasta. Ajustar conforme o `angular.json` real. Para `ng new frontend`, o output padrão pode ser `dist/frontend/` ou `dist/<project-name>/`. Verificar documentação do Angular para a estrutura exata.

---

### Pensamento 4

Configuração NGINX para SPA: o Angular usa client-side routing. Todas as rotas devem servir `index.html` para que o Router do Angular trate a navegação. A configuração padrão do NGINX não faz isso — é necessário `try_files $uri $uri/ /index.html`.

### Ação 4.4

Criar arquivo `frontend/nginx.conf`:

```nginx
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;
    location / {
        try_files $uri $uri/ /index.html;
    }
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### Observação 4.4

O NGINX Alpine usa `default.conf` em `/etc/nginx/conf.d/`. O arquivo padrão pode ser sobrescrito. O COPY no Dockerfile deve colocar o `nginx.conf` em `/etc/nginx/conf.d/default.conf` para substituir a config padrão. O path no COPY é relativo ao contexto (frontend/).

---

### Pensamento 5

Segurança (ADR-002, ADR-009): a imagem deve seguir boas práticas:
- Não executar como root (NGINX Alpine por padrão já usa usuário nginx)
- Imagem mínima (alpine)
- Sem secrets no build (variáveis de ambiente para API URL podem ser injetadas em runtime — para SPA, a URL da API pode ser compilada no build ou lida de config; para estrutura inicial, pode-se deixar placeholder)

### Ação 4.5

Para a estrutura inicial, não incluir substituição de variáveis em runtime nos arquivos estáticos. O frontend será configurado com a URL da API em builds futuros (ex.: environment.prod.ts). O Dockerfile desta fase apenas builda e serve.

### Observação 4.5

Se no futuro for necessário configurar a API URL em runtime (sem rebuild), pode-se usar um script de entrypoint que substitui placeholders no index.html ou usar um arquivo config.json carregado em runtime. Fora do escopo desta fase.

---

### Pensamento 6

O backend Dockerfile fica em `backend/docker/backend/Dockerfile`. O docker-compose referencia:
```yaml
backend:
  build:
    context: ./backend
    dockerfile: docker/backend/Dockerfile
```
Para o frontend, com Dockerfile em `frontend/Dockerfile`:
```yaml
frontend:
  build:
    context: ./frontend
    dockerfile: Dockerfile
```

### Ação 4.6

Validar que a imagem builda corretamente:
```bash
docker build -t wisa-crm-frontend ./frontend
docker run -p 4200:80 wisa-crm-frontend
```
A aplicação deve ser acessível em `http://localhost:4200`. A porta 80 do container pode ser mapeada para 4200 para manter consistência com `ng serve` (porta 4200) ou para 80 se preferir.

### Observação 4.6

O docker-compose (Fase 5) definirá a porta exposta. Comum usar 4200 para frontend em dev para evitar conflito com outros serviços.

---

### Checklist de Implementação

1. [ ] Criar `frontend/Dockerfile` com build multi-stage
2. [ ] Stage 1: usar `node:20-alpine`, `npm ci`, `npm run build`
3. [ ] Stage 2: usar `nginx:alpine`
4. [ ] Copiar output do build para `/usr/share/nginx/html`
5. [ ] Criar `frontend/nginx.conf` com `try_files` para SPA routing
6. [ ] Garantir path correto do output (verificar `angular.json` outputPath)
7. [ ] Expor porta 80 no container
8. [ ] Validar com `docker build` e `docker run`
9. [ ] Documentar no README ou na feature como executar

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-007 NGINX | Arquivos estáticos servidos via NGINX |
| ADR-002 Angular | Build com `ng build --configuration production` |
| Segurança | Imagem mínima (alpine), multi-stage |
| SPA routing | try_files para fallback index.html |

---

## Referências

- [ADR-007 — NGINX como reverse proxy](../../../adrs/ADR-007-nginx-como-reverse-proxy.md)
- [ADR-002 — Angular como framework](../../../adrs/ADR-002-angular-como-framework-frontend.md)
- [Dockerizing Angular App](https://angular.dev/tools/cli/deploy)
- [NGINX serving static files](https://nginx.org/en/docs/http/ngx_http_core_module.html#try_files)
