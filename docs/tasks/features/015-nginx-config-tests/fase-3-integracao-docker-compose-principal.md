# Fase 3 — Integração ao Docker Compose Principal

## Objetivo

Adicionar o serviço NGINX ao `docker-compose.yml` na raiz do projeto, configurando dependências e rede para que o NGINX atue como reverse proxy para todos os serviços (backend, frontend, test-app-backend, test-app-frontend) e permita a execução de testes de integração end-to-end com subdomínios. Também é necessário alterar na aplicação test-app as variáveis `APP_URL` e `FRONTEND_URL` para que o fluxo de callback e redirects funcione corretamente quando acessado via NGINX.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O docker-compose principal já contém: redis, postgres, backend, frontend, test-app-backend, test-app-frontend. O NGINX precisa:
1. Estar na mesma rede que os demais serviços (rede default do compose)
2. Depender dos serviços que fará proxy (backend, frontend, test-app-*)
3. Expor a porta 80 (ou outra) para o host
4. Montar o arquivo de configuração `nginx/conf.d/wisa.conf`

Ao usar `depends_on`, o NGINX só iniciará após os outros estarem up. Para health checks, podemos usar `condition: service_healthy` onde aplicável, mas postgres e redis já têm; backend e frontend podem não ter. Manter `depends_on` simples (sem condition) para backend, frontend, test-app-backend, test-app-frontend.

### Ação 3.1

Adicionar o serviço `nginx` ao `docker-compose.yml` na raiz:

```yaml
nginx:
  image: nginx:alpine
  ports:
    - "80:80"
  volumes:
    - ./nginx/conf.d/wisa.conf:/etc/nginx/conf.d/wisa.conf:ro
  depends_on:
    - backend
    - frontend
    - test-app-backend
    - test-app-frontend
```

**Observação 3.1:** O `depends_on` garante ordem de inicialização; não garante que os serviços estejam prontos para receber requisições. Os backends Go e o frontend Angular em containers podem levar alguns segundos para ficar prontos. O NGINX retornará 502 até que o upstream responda. Para testes manuais, isso é aceitável; para CI, pode ser necessário um script de wait/retry.

---

### Pensamento 2

Conflito de portas: o projeto expõe atualmente:
- 6379 (redis)
- 5432 (postgres)
- 8080 (backend)
- 4200 (frontend)
- 8081 (test-app-backend)
- 4201 (test-app-frontend)

A porta 80 não está em uso. O NGINX na 80 será o ponto de entrada único quando ativo. O usuário acessará:
- `http://auth.wisa.labs.com.br` (porta 80)
- `http://lingerie-maria.wisa.labs.com.br` (porta 80)

As portas 4200, 8080, 4201, 8081 podem permanecer expostas para desenvolvimento sem NGINX, ou podem ser mantidas para debug. O requisito não pede removê-las; adicionar o NGINX na 80 é suficiente.

### Ação 3.2

Manter a porta 80 para o NGINX. Documentar no README do projeto ou na documentação da feature que, para testes com NGINX, o acesso é via porta 80 e os subdomínios configurados em /etc/hosts. As portas 4200, 8080 etc. continuam disponíveis para desenvolvimento direto (sem NGINX).

### Observação 3.2

Em produção/VPS, o NGINX escuta em 80/443 e os backends não são expostos. No ambiente de desenvolvimento Docker, manter ambas as opções (com e sem NGINX) é útil.

---

### Pensamento 3

Estrutura do docker-compose após integração: o serviço nginx usa caminho relativo `./nginx/conf.d/wisa.conf`. Esse caminho é relativo ao diretório do arquivo docker-compose (raiz do projeto). A pasta `nginx/` deve existir na raiz, e `nginx/conf.d/wisa.conf` deve ter sido criado na Fase 1.

### Ação 3.3

Garantir que o path do volume está correto e que o arquivo existe. Na implementação, verificar que `nginx/conf.d/wisa.conf` existe antes de rodar `docker compose up`.

### Observação 3.3

Se o arquivo não existir, o Docker montará um diretório vazio no lugar (comportamento do bind mount), e o NGINX pode falhar ao carregar. A Fase 1 deve ser concluída antes da Fase 3.

---

### Pensamento 4

Atualização do TRACKER principal e documentação. O `docs/tasks/TRACKER.md` deve incluir a feature 015 na tabela "Status Geral". O `docs/vps/configurations.md` ou um novo doc pode referenciar a config NGINX para ambiente Docker de testes, diferenciando da config VPS (que usa paths como /var/www e 127.0.0.1).

### Ação 3.4

Atualizar `docs/tasks/TRACKER.md`:
- Adicionar linha na tabela "Status Geral" para a feature 015
- Adicionar entrada no "Diário de Sessões" quando a implementação for concluída

Opcional: adicionar em `docs/vps/configurations.md` ou em `nginx/README.md` uma seção "Config Docker para testes" mencionando que a config usa nomes de serviço (backend, frontend) em vez de 127.0.0.1, e que em VPS a config seria adaptada.

### Observação 3.4

A documentação existente em docs/vps/configurations.md descreve config para VPS (127.0.0.1, paths em disco). A config da pasta nginx é para Docker (nomes de serviço). Manter a distinção clara evita confusão na hora do deploy em VPS.

---

### Pensamento 5

Variáveis de ambiente da test-app: ela usa `AUTH_SERVER_URL`, `APP_URL`, `FRONTEND_URL`. Quando atrás do NGINX, a `APP_URL` e `FRONTEND_URL` devem refletir a URL pública. Ex.: se o usuário acessa `http://lingerie-maria.wisa.labs.com.br`, então `APP_URL=http://lingerie-maria.wisa.labs.com.br` e `FRONTEND_URL=http://lingerie-maria.wisa.labs.com.br`. O callback do auth redirecionará para essa URL.

O docker-compose atual da test-app usa:
- `APP_URL: http://localhost:4201`
- `FRONTEND_URL: http://localhost:4201`

Para funcionar com NGINX, essas variáveis precisam ser alteradas para o subdomínio. A restrição "Não altere backend e frontend" refere-se ao backend e frontend **principais** do wisa-crm-service; a test-app pode ser ajustada.

### Ação 3.5

**Alterar na aplicação test-app** (serviço `test-app-backend` no docker-compose.yml) as variáveis de ambiente `APP_URL` e `FRONTEND_URL`:

- De: `http://localhost:4201`
- Para: `http://lingerie-maria.wisa.labs.com.br`

Essa alteração garante que o fluxo E2E funcione corretamente com NGINX: o callback do auth server e os redirects apontarão para o subdomínio correto. A alteração é feita no `docker-compose.yml`, na definição do serviço `test-app-backend`.

```yaml
test-app-backend:
  environment:
    APP_URL: http://lingerie-maria.wisa.labs.com.br
    FRONTEND_URL: http://lingerie-maria.wisa.labs.com.br
    # ... demais variáveis
```

### Observação 3.5

O backend principal e o frontend principal não precisam de alteração de variáveis para o NGINX — o auth já responde em /api e /. A test-app é quem precisa saber sua URL pública para o callback. Caso seja necessário rodar a test-app sem NGINX (ex.: desenvolvimento local direto em localhost:4201), pode-se usar profiles no docker-compose ou um arquivo `.env` local para sobrescrever esses valores.

---

## Resumo das Ações da Fase 3

| # | Ação | Resultado esperado |
|---|------|-------------------|
| 3.1 | Adicionar serviço nginx ao docker-compose.yml | NGINX na rede, depends_on, volume ou build |
| 3.2 | Definir porta 80 | NGINX acessível em :80 |
| 3.3 | Validar path do volume / build | Config NGINX disponível |
| 3.4 | Atualizar TRACKER e docs | Feature 015 registrada |
| 3.5 | **Alterar test-app: APP_URL e FRONTEND_URL** | Valores atualizados para `http://lingerie-maria.wisa.labs.com.br` no serviço test-app-backend |

---

## Checklist de Implementação

- [ ] Adicionar serviço `nginx` ao `docker-compose.yml` na raiz
- [ ] Configurar build ou volume para config NGINX
- [ ] Configurar `depends_on` para backend, frontend, test-app-backend, test-app-frontend
- [ ] Expor porta 80
- [ ] **Alterar no `docker-compose.yml` as variáveis `APP_URL` e `FRONTEND_URL` do serviço `test-app-backend`** para `http://lingerie-maria.wisa.labs.com.br`
- [ ] Documentar em nginx/README ou docs a necessidade de /etc/hosts
- [ ] Atualizar docs/tasks/TRACKER.md com a feature 015

---

## Riscos e Mitigações

| Risco | Mitigação |
|-------|-----------|
| Porta 80 em uso por outro processo | Documentar conflito; sugerir parar outro serviço ou alterar porta do NGINX |
| Callback da test-app aponta para URL errada | Documentar e/ou configurar APP_URL/FRONTEND_URL para o subdomínio |
| 502 ao acessar antes dos serviços subirem | Aguardar alguns segundos após `docker compose up`; ou adicionar healthchecks |

---

## Conformidade

- **ADR-007:** NGINX como reverse proxy
- **ADR-009:** Config compatível com ambiente de desenvolvimento
- **Code guidelines:** N/A
- **Restrição:** Não alterar backend/ e frontend/ do projeto principal
