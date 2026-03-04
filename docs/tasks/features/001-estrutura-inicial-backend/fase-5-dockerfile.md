# Fase 5 — Criar Dockerfile para o Backend

## Objetivo

Criar um Dockerfile para construir e executar o backend Go em container, permitindo deploy consistente e uso em ambientes de desenvolvimento com Docker.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-009 adota VPS com binário nativo e systemd em produção — sem Docker. Porém, o Dockerfile é útil para:
- Desenvolvimento local com `docker build` e `docker run`
- CI/CD (testes em container)
- Possível uso futuro em orquestração (se a arquitetura evoluir)
- Docker Compose para desenvolvimento com PostgreSQL (planejado em tasks futuras)

### Ação 5.1

O Dockerfile deve produzir uma imagem **mínima e segura**, seguindo boas práticas:
- Build multi-stage (compilar em uma etapa, executar em outra)
- Imagem final baseada em `alpine` ou `scratch` para reduzir superfície de ataque
- Usuário não-root na execução
- Binário estático (CGO_ENABLED=0) para portabilidade

### Observação 5.1

ADR-001 enfatiza binário estático e baixo footprint. Alpine com binário estático é adequado. A imagem `scratch` é ainda menor, mas exige binário totalmente estático e pode dificultar debugging. Alpine é um equilíbrio.

---

### Pensamento 2

A estrutura do projeto tem `backend/` como submódulo. O Dockerfile deve ficar em `backend/Dockerfile` (junto ao código) ou na raiz? Convenção comum: Dockerfile no mesmo diretório do código que construirá. Portanto: **`backend/Dockerfile`**.

### Ação 5.2

Posicionar o Dockerfile em `backend/Dockerfile`. O `WORKDIR` e `COPY` serão relativos a esse contexto. O comando `docker build` deve ser executado a partir de `backend/` ou com `-f backend/Dockerfile` e contexto na raiz.

### Observação 5.2

Se o `go.mod` estiver em `backend/`, o contexto de build pode ser `backend/` e os caminhos serão simples. Se houver `go.work` ou módulos na raiz, pode ser necessário ajustar.

---

### Pensamento 3

Build multi-stage:
- **Stage 1 (builder):** Imagem com Go 1.22+, compila o código
- **Stage 2 (runtime):** Imagem mínima (alpine ou distroless), apenas o binário

### Ação 5.3

Estrutura do Dockerfile:

```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o wisa-crm-service ./cmd/api

# Stage 2: Runtime
FROM alpine:3.19
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/wisa-crm-service .
USER appuser
EXPOSE 8080
CMD ["./wisa-crm-service"]
```

### Observação 5.3

- `-ldflags="-w -s"` reduz o tamanho do binário (strip debug info)
- `adduser -D` cria usuário sem home, sem shell
- `EXPOSE 8080` documenta a porta; o mapeamento real é feito no `docker run -p`
- A porta 8080 deve ser configurável via variável de ambiente (PORT) no .env; o binário já lerá isso

---

### Pensamento 4

O ADR-009 e as guidelines de segurança recomendam: não executar como root, limitar recursos, não expor informações desnecessárias. O Dockerfile já usa usuário não-root. Para produção em VPS sem Docker, o systemd cuida do hardening. Para o container, as boas práticas acima são suficientes para a fase inicial.

### Ação 5.4

Considerar `HEALTHCHECK` no Dockerfile para que o orchestrator (Docker, Docker Compose, k8s) verifique a saúde da aplicação via endpoint `/health`:

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1
```

Alpine mínimo pode não ter `wget`. Alternativa: `curl` ou instalar `curl` no stage runtime. Ou usar `HEALTHCHECK` com comando que apenas verifica se o processo está rodando — menos robusto. A opção mais limpa é adicionar `curl` ou `wget` ao alpine: `RUN apk add --no-cache curl` e usar `curl -f http://localhost:8080/health`.

### Observação 5.4

Para minimalismo, pode-se omitir o HEALTHCHECK nesta fase e adicionar em fase posterior. Ou usar `curl` que é comum em imagens alpine.

---

### Pensamento 5

As variáveis de ambiente (PORT, etc.) serão passadas em runtime via `docker run -e` ou `docker-compose environment`. O container não deve conter arquivo `.env` com secrets. O `.env` é para desenvolvimento local; em container, as variáveis vêm do orchestrator.

### Ação 5.5

Documentar que, ao rodar o container, as variáveis devem ser passadas:
```bash
docker run -p 8080:8080 -e PORT=8080 wisa-crm-service
```
Ou via `--env-file` se o usuário tiver um `.env` local (cuidado para não fazer commit do env file).

### Observação 5.5

Nenhuma modificação no Dockerfile além do já planejado. A documentação na fase de "como rodar" cobrirá isso.

---

### Checklist de Implementação

1. [ ] Criar `backend/Dockerfile` com build multi-stage
2. [ ] Usar `golang:1.22-alpine` no stage de build
3. [ ] Usar `alpine:3.19` (ou similar) no stage de runtime
4. [ ] Configurar `CGO_ENABLED=0` para binário estático
5. [ ] Copiar `go.mod` e `go.sum` primeiro (aproveitar cache de camadas)
6. [ ] Executar `go mod download` antes de `COPY` do código fonte
7. [ ] Criar usuário não-root e executar com `USER appuser`
8. [ ] Expor porta 8080
9. [ ] Definir `CMD` para executar o binário
10. [ ] Opcional: adicionar `HEALTHCHECK` usando `/health`
11. [ ] Validar com `docker build -t wisa-crm-service ./backend` e `docker run -p 8080:8080 wisa-crm-service`

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-001 Go 1.22 | Imagem golang:1.22-alpine |
| ADR-009 (segurança) | Usuário não-root |
| Binário estático | CGO_ENABLED=0 |
| Imagem mínima | Multi-stage, alpine runtime |

---

## Referências

- [ADR-001 — Go como linguagem](../../../adrs/ADR-001-golang-como-linguagem-de-backend.md)
- [ADR-009 — Infraestrutura VPS](../../../adrs/ADR-009-infraestrutura-vps-linux.md)
- [Docker Best Practices for Go](https://docs.docker.com/language/golang/build-images/)
