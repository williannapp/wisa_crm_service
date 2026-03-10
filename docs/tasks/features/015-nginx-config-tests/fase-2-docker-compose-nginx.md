# Fase 2 — Dockerfile NGINX e Integração ao Docker Compose do Projeto

## Objetivo

Criar um **Dockerfile** para o NGINX na pasta `nginx/`, incorporando a configuração de roteamento definida na Fase 1, e adicionar o serviço NGINX ao **docker-compose.yml** principal do projeto. Dessa forma, o NGINX será construído como imagem customizada e integrado à stack existente.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O projeto utiliza `build` + `Dockerfile` para todos os serviços customizados (postgres, backend, frontend, test-app). Para consistência, o NGINX deve seguir o mesmo padrão: um Dockerfile que gera uma imagem com a configuração embutida, em vez de montar volumes em tempo de execução.

### Ação 2.1

Criar `nginx/Dockerfile` baseado em `nginx:alpine`:
- Remover o `default.conf` da imagem para evitar conflito com os server blocks customizados
- Copiar `conf.d/wisa.conf` para `/etc/nginx/conf.d/`
- Manter a imagem enxuta (single-stage, sem alterações desnecessárias)

### Observação 2.1

A imagem `nginx:alpine` inclui `/etc/nginx/conf.d/default.conf`. Removê-lo garante que apenas nossos server blocks (`auth.wisa.labs.com.br` e `lingerie-maria.wisa.labs.com.br`) estejam ativos. A configuração fica versionada na imagem, facilitando deploy e reprodução.

---

### Pensamento 2

Estrutura final da pasta `nginx/` após Fase 1 + Fase 2:

```
nginx/
├── conf.d/
│   └── wisa.conf
├── Dockerfile
└── README.md          # Instruções de uso e pré-requisitos
```

O `Dockerfile`:

```dockerfile
FROM nginx:alpine
RUN rm /etc/nginx/conf.d/default.conf
COPY conf.d/wisa.conf /etc/nginx/conf.d/wisa.conf
```

### Ação 2.2

Criar o `nginx/Dockerfile` conforme acima. O NGINX inclui automaticamente todos os `.conf` em `/etc/nginx/conf.d/`.

### Observação 2.2

Usar `COPY conf.d/` permite adicionar mais arquivos de config no futuro sem alterar o Dockerfile. Alternativa: `COPY conf.d/*.conf /etc/nginx/conf.d/` para garantir que apenas `.conf` sejam copiados.

---

### Pensamento 3

O serviço NGINX deve ser adicionado ao **docker-compose.yml** na raiz do projeto, seguindo o padrão dos outros serviços (build com context e dockerfile, depends_on, porta exposta).

### Ação 2.3

Adicionar o serviço `nginx` ao `docker-compose.yml` principal:

```yaml
nginx:
  build:
    context: ./nginx
    dockerfile: Dockerfile
  ports:
    - "80:80"
  depends_on:
    - backend
    - frontend
    - test-app-backend
    - test-app-frontend
```

### Observação 2.3

O NGINX fará proxy para os outros serviços; o `depends_on` garante ordem de inicialização. A porta 80 expõe o NGINX como ponto de entrada único para testes com subdomínios (via /etc/hosts).

---

### Pensamento 4

Criar `nginx/README.md` com:
- Descrição do propósito (config NGINX para testes de integração)
- Estrutura do Dockerfile e da config
- Como validar a config: `docker compose run --rm nginx nginx -t`
- Pré-requisito: /etc/hosts com auth.wisa.labs.com.br e lingerie-maria.wisa.labs.com.br
- Como subir a stack com NGINX: `docker compose up -d`
- Conformidade com ADR-007

### Ação 2.4

Redigir `nginx/README.md` conforme o planejamento.

---

## Resumo das Ações da Fase 2

| # | Ação | Resultado esperado |
|---|------|-------------------|
| 2.1 | Criar nginx/Dockerfile | Imagem baseada em nginx:alpine, config wisa.conf embutida |
| 2.2 | Definir conteúdo do Dockerfile | Remover default.conf, COPY wisa.conf |
| 2.3 | Adicionar serviço nginx ao docker-compose.yml | Build do contexto nginx/, porta 80, depends_on |
| 2.4 | Criar nginx/README.md | Instruções de uso e referências |

---

## Checklist de Implementação

- [ ] Criar `nginx/Dockerfile` (FROM nginx:alpine, remover default, COPY wisa.conf)
- [ ] Criar `nginx/conf.d/wisa.conf` (conteúdo da Fase 1)
- [ ] Adicionar serviço `nginx` ao `docker-compose.yml` na raiz
- [ ] Configurar build (context: ./nginx, dockerfile: Dockerfile)
- [ ] Configurar depends_on para backend, frontend, test-app-backend, test-app-frontend
- [ ] Expor porta 80
- [ ] Criar `nginx/README.md` com instruções
- [ ] Validar que `docker compose build nginx` e `docker compose up -d nginx` funcionam

---

## Riscos e Mitigações

| Risco | Mitigação |
|-------|-----------|
| Porta 80 já em uso | Documentar uso de porta alternativa (ex.: 8000:80) |
| Config inválida quebra o build | Usar `nginx -t` após build: `docker compose run --rm nginx nginx -t` |

---

## Conformidade

- **ADR-007:** Config NGINX para reverse proxy
- **Code guidelines:** N/A (infraestrutura)
