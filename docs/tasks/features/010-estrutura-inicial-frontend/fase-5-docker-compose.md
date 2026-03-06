# Fase 5 — Adicionar Frontend no Docker Compose

## Objetivo

Adicionar o serviço `frontend` no `docker-compose.yml` para permitir execução integrada do frontend Angular com o backend, Redis e PostgreSQL em ambiente de desenvolvimento.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O `docker-compose.yml` atual possui três serviços: `redis`, `postgres`, `backend`. O frontend será um quarto serviço. O frontend Angular, após build, é uma SPA estática servida por NGINX e não depende de PostgreSQL ou Redis diretamente. Ele se comunica com o backend via HTTP (API).

### Ação 5.1

O serviço frontend não precisa de `depends_on` para postgres ou redis — a comunicação é via HTTP com o backend. Porém, para uma experiência de desenvolvimento integrada, pode-se configurar `depends_on: backend` para garantir que o backend esteja disponível quando o usuário acessar o frontend (o frontend fará chamadas à API do backend).

### Observação 5.1

Em desenvolvimento, o frontend pode precisar da URL do backend (ex.: `http://backend:8080` ou `http://localhost:8080` dependendo de como o browser acessa). O browser do usuário faz as requisições, não o container do frontend — portanto, a URL da API deve ser acessível **do cliente** (navegador). Se o usuário acessa o frontend em `localhost:4200`, as requisições à API partirão do browser; `localhost:8080` funcionará se o backend estiver mapeado na porta 8080 do host. Não há necessidade de `depends_on` para o frontend em termos de ordem de startup — o frontend pode subir independentemente; `depends_on` é útil para garantir que o backend esteja "healthy" antes de considerar o stack pronto, mas isso é opcional.

---

### Pensamento 2

Estrutura do serviço frontend no docker-compose:

```yaml
frontend:
  build:
    context: ./frontend
    dockerfile: Dockerfile
  ports:
    - "4200:80"
  depends_on:
    - backend
```

A porta 80 do container (NGINX) é mapeada para 4200 no host, mantendo consistência com a porta padrão do `ng serve`.

### Ação 5.2

Definir o serviço `frontend` com:
- `build`: context `./frontend`, dockerfile `Dockerfile`
- `ports`: `4200:80` (host:container)
- `depends_on`: `backend` (opcional, para ordem de startup)

### Observação 5.2

O frontend não precisa de variáveis de ambiente para a estrutura inicial, pois a API URL pode ser hardcoded ou configurada em um environment file posteriormente. Para desenvolvimento, a API estará em `http://localhost:8080` quando acessada do browser do usuário.

---

### Pensamento 3

Em produção (VPS), o NGINX do host (ADR-007) serve os arquivos estáticos e faz proxy para o backend. No Docker Compose de desenvolvimento, temos dois cenários:

1. **Desenvolvimento local com containers:** frontend em container, backend em container — o usuário acessa `localhost:4200` (frontend) e `localhost:8080` (API). O frontend precisa saber que a API está em `http://localhost:8080` — isso é configurado no código do Angular (environment.ts ou similar).

2. **Simulação de produção:** poderia ter um NGINX no compose que faz proxy de `/` para o frontend e `/api` para o backend — mas isso seria mais complexo e não é requisito para a estrutura inicial.

### Ação 5.3

Para a estrutura inicial, o docker-compose apenas adiciona o serviço frontend. A configuração da URL da API no frontend será feita em features futuras (tela de login, interceptors HTTP). Nesta fase, o frontend sobe e exibe a aplicação Angular padrão (placeholder).

### Observação 5.3

Nenhuma variável de ambiente obrigatória para o serviço frontend nesta fase.

---

### Pensamento 4

Rede Docker: por padrão, todos os serviços estão na mesma rede e podem se comunicar pelo nome do serviço (ex.: `backend`, `postgres`). O frontend, servindo HTML/JS estático, não faz chamadas de rede de dentro do container — o browser do usuário faz as chamadas. Portanto, a rede padrão é suficiente.

### Ação 5.4

Não é necessário configurar rede customizada. O padrão do docker-compose já permite que os serviços coexistam.

### Observação 5.4

Se no futuro houver um proxy reverso (NGINX) no compose que roteie tráfego, a configuração de rede poderia ser ajustada. Fora do escopo desta fase.

---

### Pensamento 5

Ordem de serviços no arquivo: manter organização lógica. Sugestão: postgres, redis, backend, frontend (ordem de dependência).

### Ação 5.5

Inserir o bloco `frontend` após o bloco `backend` no `docker-compose.yml`. A estrutura final terá:
- postgres
- redis
- backend (depends_on: postgres, redis)
- frontend (depends_on: backend)

### Observação 5.5

`depends_on` para frontend é opcional — o frontend pode subir mesmo se o backend falhar. Porém, para UX de desenvolvimento, faz sentido que o backend esteja pronto quando o desenvolvedor acessar o frontend. Usar `depends_on: backend` sem `condition: service_healthy` (o backend pode não ter healthcheck no compose atual — verificar). O backend no compose atual não tem healthcheck explícito no serviço; o postgres e redis têm. Para `depends_on: backend`, o Docker aguarda o backend estar "running", não necessariamente "healthy". Isso é suficiente.

---

### Pensamento 6

Volumes: o frontend em produção é estático (build já feito na imagem). Para desenvolvimento com hot reload, uma abordagem seria montar o código fonte e usar um container com `ng serve` — mas isso seria um Dockerfile diferente (dev). A estrutura inicial usa o build de produção no container. Se no futuro for desejado um modo "dev" com volume mount, pode-se adicionar um profile ou um docker-compose.override.yml.

### Ação 5.6

Não adicionar volumes para a estrutura inicial. O frontend usa a imagem buildada. Volumes para hot reload podem ser feature futura.

### Observação 5.6

Consistente com o objetivo de "estrutura inicial" — deploy pronto para uso, não necessariamente desenvolvimento com live reload no Docker.

---

### Checklist de Implementação

1. [ ] Abrir `docker-compose.yml` na raiz do projeto
2. [ ] Adicionar serviço `frontend` após o serviço `backend`
3. [ ] Configurar `build.context: ./frontend` e `build.dockerfile: Dockerfile`
4. [ ] Configurar `ports: "4200:80"`
5. [ ] Adicionar `depends_on: [backend]` (opcional)
6. [ ] Validar com `docker compose build frontend` e `docker compose up`
7. [ ] Verificar que o frontend está acessível em `http://localhost:4200`
8. [ ] Documentar no README ou na documentação do projeto como subir o stack completo

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| docker-compose | Serviço frontend adicionado e funcional |
| Integração | Stack completo (postgres, redis, backend, frontend) |
| Porta | 4200 para frontend (consistente com ng serve) |
| ADR-007 | Frontend servido por NGINX no container (prévia do fluxo de produção) |

---

## Referências

- [Docker Compose file reference](https://docs.docker.com/compose/compose-file/)
- [ADR-007 — NGINX como reverse proxy](../../../adrs/ADR-007-nginx-como-reverse-proxy.md)
