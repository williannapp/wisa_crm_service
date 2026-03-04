# Fase 4 — Configuração de Variáveis de Ambiente

## Objetivo

Criar o suporte à leitura de variáveis de ambiente necessárias para rodar o projeto. O requisito original menciona ".venv" — em contexto Go, isso se traduz em `.env` (arquivo de variáveis) e `.env.example` (template versionado).

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

Em projetos Python, `.venv` é o diretório do ambiente virtual. Em Go, não há conceito de venv. O requisito "armazenar as variáveis de ambiente" indica a necessidade de um mecanismo para configurar a aplicação via variáveis de ambiente. As opções em Go são:

1. **Variáveis de ambiente do sistema** — `os.Getenv()`
2. **Arquivo `.env`** — carregado por bibliotecas como `godotenv`
3. **Arquivo de configuração** (YAML, JSON) — mais complexo para fase inicial

A abordagem padrão em desenvolvimento local é usar `.env` carregado por `godotenv`, já previsto na Fase 2.

### Ação 4.1

Definir que o projeto usará:
- **`.env`** — arquivo local (gitignored) com valores reais
- **`.env.example`** — template versionado com nomes das variáveis e valores de exemplo (sem secrets)

### Observação 4.1

O `.env.example` serve como documentação e onboarding. Novos desenvolvedores copiam para `.env` e preenchem os valores locais.

---

### Pensamento 2

Sobre "Criar um `.venv`": interpretar como "criar o suporte/configuração para variáveis de ambiente". Uma interpretação alternativa seria criar literalmente um diretório `.venv` como convenção para guardar arquivos de ambiente — mas isso não é padrão em Go. A convenção é `.env` na raiz do backend ou na raiz do projeto.

### Ação 4.2

Posicionar os arquivos:
- `backend/.env.example` — template das variáveis do backend
- `backend/.env` — arquivo local (criado pelo dev, gitignored)

Ou, se as variáveis forem compartilhadas (ex: docker-compose na raiz):
- `wisa-crm-service/.env.example` na raiz
- `wisa-crm-service/.env` na raiz (gitignored)

Para esta fase inicial, o backend é o único componente. Portanto: **`backend/.env.example`** e instrução para criar **`backend/.env`** localmente.

### Observação 4.2

Consistente com estrutura monolítica do backend.

---

### Pensamento 3

Quais variáveis são necessárias para a estrutura inicial (health endpoint, servidor HTTP)?

| Variável | Exemplo | Obrigatória | Descrição |
|----------|---------|-------------|-----------|
| `PORT` | `8080` | Não (default 8080) | Porta do servidor HTTP |
| `APP_ENV` | `development` | Não | development, staging, production |

Para fases futuras (banco, JWT): `DATABASE_URL`, `JWT_PRIVATE_KEY_PATH`, `JWT_PUBLIC_KEY_PATH`, etc. Na Fase 4, incluir apenas as variáveis necessárias para rodar o servidor e o health check.

### Ação 4.3

Conteúdo de `backend/.env.example`:
```
# Server
PORT=8080
APP_ENV=development

# Database (para fases futuras - opcional agora)
# DATABASE_URL=postgres://user:pass@localhost:5432/wisa_crm_db?sslmode=disable
```

### Observação 4.3

Comentários ajudam o desenvolvedor a saber quando cada variável será utilizada.

---

### Pensamento 4

O carregamento do `.env` deve ocorrer no `main.go`, antes de qualquer inicialização que dependa de configuração. A biblioteca `godotenv` oferece `godotenv.Load()`. Por padrão, carrega `.env` no diretório de trabalho atual. O executável, quando rodado de `backend/`, terá CWD em `backend/` se iniciado de lá. Para garantir, pode-se usar path absoluto ou relativo ao executável.

### Ação 4.4

No `main.go` (criado na Fase 6), incluir no início:
```go
if err := godotenv.Load(); err != nil {
    log.Print("No .env file found, using system environment")
}
```
O log indica que a ausência de `.env` não é fatal — as variáveis de sistema podem ser usadas (útil em produção via systemd EnvironmentFile).

### Observação 4.4

Em produção (ADR-009), as variáveis vêm do `EnvironmentFile` do systemd. O `.env` é apenas para desenvolvimento local. O `godotenv.Load()` pode falhar silenciosamente em produção, e a aplicação usará `os.Getenv()`.

---

### Pensamento 5

Criar ou não um pacote de configuração? Para a fase inicial, a leitura direta com `os.Getenv("PORT")` no `main.go` é suficiente. Um pacote `internal/config` ou `pkg/config` pode ser criado em fase posterior quando houver validação, defaults e mais variáveis.

### Ação 4.5

Manter simplicidade: ler `os.Getenv("PORT")` diretamente no `main.go`. Se vazio, usar default `8080`.

### Observação 4.5

YAGNI (You Aren't Gonna Need It). Evita over-engineering na fase inicial.

---

### Checklist de Implementação

1. [ ] Criar `backend/.env.example` com variáveis PORT e APP_ENV (e placeholders comentados para DATABASE_URL etc.)
2. [ ] Garantir que `backend/.env` está no `.gitignore` (Fase 3)
3. [ ] Adicionar no `main.go` a chamada `godotenv.Load()` no startup
4. [ ] Usar `os.Getenv("PORT")` com fallback para "8080"
5. [ ] Documentar no README ou em `docs/` que o desenvolvedor deve copiar `.env.example` para `.env` e ajustar valores
6. [ ] Opcional: criar script ou instrução `cp backend/.env.example backend/.env`

---

## Conformidade

| Requisito | Atendimento |
|-----------|-------------|
| ADR-009 | Variáveis sensíveis não em .env.example; .env gitignored |
| Segurança | Nenhum secret no template |
| Code Guidelines | Simplicidade mantida |

---

## Nota sobre ".venv"

O termo `.venv` no requisito foi interpretado como "mecanismo para variáveis de ambiente". Em Go, a convenção é `.env` + `.env.example`. Se no futuro o projeto integrar ferramentas Python que usem `.venv`, o diretório `.venv/` já estará no `.gitignore` (Fase 3).

---

## Referências

- [godotenv](https://github.com/joho/godotenv)
- [ADR-009 — Variáveis via EnvironmentFile](../../../adrs/ADR-009-infraestrutura-vps-linux.md)
