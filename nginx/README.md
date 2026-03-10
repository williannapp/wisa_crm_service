# NGINX — Reverse Proxy para Testes

Configuração do NGINX como reverse proxy para execução de testes de integração end-to-end do wisa-crm-service.

## Subdomínios

| Subdomínio | Descrição |
|------------|-----------|
| `auth.wisa.labs.com.br` | Portal de autenticação (frontend) + API (backend) + JWKS |
| `lingerie-maria.wisa.labs.com.br` | Aplicação de teste (test-app) |

## Pré-requisitos

Adicione os subdomínios ao `/etc/hosts` para testes locais:

```
127.0.0.1 auth.wisa.labs.com.br
127.0.0.1 lingerie-maria.wisa.labs.com.br
```

## Uso

### Stack completa (recomendado para testes E2E)

O NGINX está integrado ao `docker-compose.yml` principal. Para subir toda a stack incluindo o NGINX:

```bash
docker compose up -d
```

O NGINX ficará disponível na porta **80**. Acesse:

- http://auth.wisa.labs.com.br — Portal de login
- http://lingerie-maria.wisa.labs.com.br — Test-app

**Nota:** O backend redireciona o callback para `https://` após o login. Para fluxo E2E completo sem TLS local, considere usar [mkcert](https://github.com/FiloSottile/mkcert) para certificados locais.

### Execução isolada (build)

O `nginx/docker-compose.yml` permite buildar a imagem do NGINX sem subir a stack completa:

```bash
cd nginx
docker compose build
```

**Nota:** O `nginx -t` resolve os hosts (backend, frontend, etc.) ao carregar a config. Para validar a sintaxe, use o compose principal (onde os serviços estão na mesma rede).

## Validar configuração

Na raiz do projeto, para validar a sintaxe da config:

```bash
docker compose run --rm nginx nginx -t
```

## Estrutura

```
nginx/
├── conf.d/
│   └── wisa.conf      # Server blocks
├── docker-compose.yml # Compose isolado para validação
├── Dockerfile         # Imagem baseada em nginx:alpine
└── README.md          # Este arquivo
```

## Referências

- [ADR-007 — NGINX como Reverse Proxy](../docs/adrs/ADR-007-nginx-como-reverse-proxy.md)
- [Feature 015 — Config NGINX para testes](../docs/tasks/features/015-nginx-config-tests/TRACKER.md)
