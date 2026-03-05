# Fase 1 — Infraestrutura Redis

## Objetivo

Adicionar Redis ao projeto para armazenamento de authorization codes temporários. Inclui configuração Docker, variáveis de ambiente e cliente Redis em Go, em conformidade com ADR-010 e code guidelines.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O ADR-010 define que o fluxo de autenticação usa Redis para armazenar authorization codes com TTL curto. O Redis é uma dependência de infraestrutura obrigatória. O projeto já utiliza Docker Compose (postgres + backend); adicionar Redis segue o mesmo padrão.

A conexão com Redis deve ser configurável via variável de ambiente (ex.: `REDIS_URL`) para permitir diferentes ambientes (dev local, staging, produção). O cliente Go recomendado é `github.com/redis/go-redis/v9` — biblioteca madura, com suporte a context e connection pool.

### Ação 1.1

Adicionar serviço Redis ao `docker-compose.yml`:

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  healthcheck:
    test: ["CMD", "redis-cli", "ping"]
    interval: 5s
    timeout: 3s
    retries: 5
```

### Observação 1.1

A imagem `redis:7-alpine` é leve e adequada para desenvolvimento. Para produção, considerar Redis com persistência (RDB/AOF) e configurações de segurança. O healthcheck garante que o backend só suba após o Redis estar disponível.

---

### Pensamento 2

O backend precisa depender do Redis para que o serviço inicie apenas quando o Redis estiver pronto. Isso evita erros de conexão no startup. O `depends_on` com `condition: service_healthy` deve ser utilizado.

### Ação 1.2

Adicionar dependência do Redis no serviço backend do docker-compose:

```yaml
backend:
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
  environment:
    # ... existentes ...
    REDIS_URL: redis://redis:6379/0
```

### Observação 1.2

A URL `redis://redis:6379/0` usa o nome do serviço Docker (`redis`) como hostname. O `/0` indica o database 0 (padrão do Redis). Para múltiplas instâncias ou namespaces, databases diferentes (1, 2, …) podem ser usados.

---

### Pensamento 3

O `.env.example` do backend deve documentar a variável `REDIS_URL` para que desenvolvedores saibam como configurar. O formato padrão é `redis://host:port/db`. Para dev local fora do Docker, `redis://localhost:6379/0`.

### Ação 1.3

Adicionar `REDIS_URL` ao `backend/.env.example`:

```env
# Redis (authorization codes)
# Formato: redis://host:port/db ou redis://:password@host:port/db
REDIS_URL=redis://localhost:6379/0
```

### Observação 1.3

O Redis padrão não exige senha; em produção, usar `redis://:senha@host:6379/0`. Nunca commitar senhas no repositório.

---

### Pensamento 4

O cliente Redis em Go deve ser inicializado no `main.go` ou em um módulo de infraestrutura. Seguindo Clean Architecture (ADR-005), a camada de infraestrutura é responsável por abstrações de I/O. O cliente Redis é um detalhe de implementação — a interface que o Use Case utilizará (AuthCodeStore) será definida na Fase 2.

Para esta fase, o foco é: (1) criar um pacote que inicializa o cliente Redis a partir de REDIS_URL; (2) garantir que o main.go carregue essa configuração; (3) expor uma função `NewRedisClient(url string) (*redis.Client, error)` ou similar.

A biblioteca `go-redis` oferece `redis.NewClient(opt)` com `redis.Options`. O parsing de URL pode ser feito com `redis.ParseURL(url)`.

### Ação 1.4

Criar pacote em `internal/infrastructure/cache/` ou `internal/infrastructure/redis/`:

- Arquivo `redis_client.go` (ou `client.go`):
  - Função `NewRedisClient(ctx context.Context, redisURL string) (*redis.Client, error)`
  - Usar `redis.ParseURL(redisURL)` para obter `*redis.Options`
  - Chamar `redis.NewClient(opt)`
  - Opcional: `client.Ping(ctx)` para validar conexão no startup
  - Retornar o cliente para injeção em outros componentes

### Observação 1.4

O `redis.Client` é thread-safe e mantém um pool de conexões internamente. Não é necessário criar múltiplos clientes. O contexto no `NewRedisClient` pode ser usado para o Ping inicial; o cliente em si não precisa de context no construtor.

---

### Pensamento 5

O `main.go` deve carregar `REDIS_URL` do ambiente e inicializar o Redis. Se a variável não estiver definida em desenvolvimento, pode-se usar default `redis://localhost:6379/0` ou exigir a variável e falhar no startup. Para consistência com DATABASE_URL, exigir a variável é mais seguro.

### Ação 1.5

No `main.go` (ou onde as dependências são construídas):
- Carregar `REDIS_URL` de `os.Getenv("REDIS_URL")`
- Se vazio em produção, falhar com log fatal
- Em desenvolvimento, usar default `redis://localhost:6379/0` se não definido
- Chamar `cache.NewRedisClient(ctx, redisURL)` e armazenar o cliente para uso nas Fases 2–4

### Observação 1.5

O cliente Redis deve ser passado para o componente que implementará o AuthCodeStore na Fase 2. Por ora, apenas garantir que o cliente seja criado e que o backend inicie sem erro quando o Redis estiver disponível.

---

### Pensamento 6

O `go.mod` precisa incluir a dependência do Redis. Verificar se `github.com/redis/go-redis/v9` já está presente; caso contrário, executar `go get github.com/redis/go-redis/v9`.

### Ação 1.6

Adicionar dependência ao projeto:

```bash
go get github.com/redis/go-redis/v9
```

### Observação 1.6

A versão v9 é estável e amplamente utilizada. O package fornece `redis.Client` para operações síncronas e suporte a pipelines, pub/sub, etc.

---

### Pensamento 7 (Verificação de Segurança)

- **Credenciais:** REDIS_URL pode conter senha; nunca logar a URL completa. Em logs, mascarar: `redis://****@host:6379/0`.
- **Rede:** Em produção, Redis deve estar em rede interna (não expor 6379 publicamente). O Docker Compose em dev expõe a porta para debugging; em prod, remover ou restringir.
- **Falha de conexão:** Se Redis estiver indisponível, o login falhará ao armazenar o code. O sistema deve tratar esse erro adequadamente (retornar 503 ou mensagem genérica) — isso será abordado na Fase 2.

### Observação 1.7

As mitigações de ADR-010 para indisponibilidade do Redis (monitoramento, fallback) são responsabilidades operacionais; o código deve falhar de forma controlada e logar o erro.

---

## Checklist de Implementação

- [ ] 1. Adicionar serviço `redis` ao `docker-compose.yml` com healthcheck
- [ ] 2. Adicionar `depends_on: redis` e `REDIS_URL` ao serviço backend
- [ ] 3. Documentar `REDIS_URL` no `.env.example`
- [ ] 4. Criar `internal/infrastructure/cache/redis_client.go` (ou equivalente) com `NewRedisClient`
- [ ] 5. Adicionar `go get github.com/redis/go-redis/v9`
- [ ] 6. Inicializar cliente Redis no `main.go` e propagar para wiring
- [ ] 7. Validar: `docker-compose up` — backend inicia após Redis; endpoint /health responde

---

## Conclusão

A Fase 1 estabelece a base para o uso do Redis no wisa-crm-service. A implementação segue os padrões do projeto (Docker, .env, Clean Architecture). Nenhuma lógica de negócio é introduzida — apenas infraestrutura.
