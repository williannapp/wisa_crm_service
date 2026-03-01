# ADR-001 — Go (Golang) como Linguagem de Backend

**Status:** Aceito  
**Data:** 2026-03-01  
**Autores:** Equipe de Arquitetura  
**Componente:** wisa-crm-service (backend)

---

## Contexto

O `wisa-crm-service` é um sistema centralizador de autenticação que atuará como Identity Provider (IdP) para múltiplos sistemas clientes em um modelo SaaS multi-tenant. Suas responsabilidades incluem:

- Autenticação de usuários com alta frequência de requisições
- Emissão e assinatura de tokens JWT com criptografia assimétrica
- Validação de assinaturas de clientes (billing)
- Exposição de endpoints REST para sistemas clientes
- Operação ininterrupta em ambiente VPS com recursos limitados

O sistema precisará suportar múltiplos tenants simultaneamente, com picos de requisições no início do dia útil (login simultâneo de usuários), exigindo alta concorrência com baixo consumo de recursos.

A escolha da linguagem de backend impacta diretamente:

- Performance e uso de CPU/memória
- Capacidade de lidar com concorrência
- Segurança por design
- Facilidade de deploy e operação em VPS
- Longevidade e manutenibilidade do código

---

## Decisão

**Go (Golang) é a linguagem escolhida para o backend do `wisa-crm-service`.**

A versão mínima adotada será **Go 1.22+**, aproveitando as melhorias de roteamento HTTP nativo e o enhanced loop variable scoping que elimina bugs clássicos de goroutines.

---

## Justificativa

### 1. Concorrência nativa por design

Go foi projetado com concorrência como primitiva de linguagem, não como adição posterior. O modelo de goroutines e channels do Go permite:

- Criação de milhares de goroutines com custo mínimo (~2KB de stack inicial vs ~1MB de uma thread OS)
- Controle explícito de comunicação entre goroutines via channels
- Primitivas de sincronização idiomáticas (`sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Once`)

Em um cenário de pico de autenticação simultânea de múltiplos tenants, o Go consegue processar centenas de requisições concorrentes em hardware modesto sem a complexidade do modelo async/await (Node.js) ou o overhead de threads pesadas (Java).

**Exemplo prático:** Um endpoint de login que consulta o banco, valida senha com bcrypt (CPU-bound) e emite JWT pode ser processado de forma completamente paralela por goroutines, sem bloqueio de event loop.

### 2. Performance e uso de recursos

Go compila para binário nativo sem dependência de runtime externo (JVM, interpretador Python, etc.). Isso resulta em:

- **Startup time sub-segundo** — crítico para ambientes containerizados e para reinicializações rápidas em VPS
- **Baixo footprint de memória** — um serviço Go de autenticação típico consome 20–60MB de RAM em idle, comparado a 200–500MB de um serviço equivalente em Java/Spring
- **CPU eficiente** — o garbage collector do Go (concurrent, tri-color mark-and-sweep) é otimizado para baixa latência, com pausas típicas abaixo de 1ms

Em um ambiente VPS com recursos compartilhados ou limitados, essa eficiência é determinante.

### 3. Segurança por tipagem estática e compilação

Go elimina classes inteiras de vulnerabilidades comuns em linguagens dinamicamente tipadas:

- **Buffer overflows** são gerenciados pelo runtime (bounds checking automático)
- **Null pointer dereferences** são detectáveis em compile-time parcialmente e em runtime com panic controlável
- **Type confusion** é impossível por design do sistema de tipos
- **Race conditions** são detectáveis com a flag `-race` do compilador durante testes

A ausência de herança e a composição explícita reduzem a complexidade e a superfície de ataque do código.

### 4. Biblioteca padrão robusta para casos de uso de auth

A biblioteca padrão do Go cobre diretamente os casos de uso do `wisa-crm-service`:

- `crypto/rsa`, `crypto/ecdsa` — operações criptográficas para JWT assimétrico
- `net/http` — servidor HTTP production-ready sem frameworks externos obrigatórios
- `crypto/bcrypt` — hashing seguro de senhas (via `golang.org/x/crypto`)
- `crypto/rand` — geração de números aleatórios criptograficamente seguros
- `encoding/json` — serialização/deserialização de claims JWT

Dependências externas são minimizadas, reduzindo a superfície de ataque via supply chain.

### 5. Deploy simplificado em VPS

Go produz um **único binário estático** (com CGO_ENABLED=0) que:

- Não requer instalação de runtime, interpretador ou JVM na VPS
- Pode ser copiado e executado diretamente
- É trivial de versionar e reverter
- Consome mínimo de espaço em disco

Isso simplifica drasticamente o processo de deploy, manutenção e recuperação de falhas em ambiente de VPS Linux.

### 6. Observabilidade e tooling maduro

O ecossistema Go oferece:

- **pprof** nativo para profiling de CPU e memória em produção
- **expvar** para métricas expostas via HTTP
- Integração nativa com **Prometheus** via `promhttp`
- **race detector** integrado ao toolchain para identificar race conditions em testes

---

## Consequências

### Positivas

- Alta performance com baixo consumo de recursos na VPS
- Concorrência segura e eficiente para múltiplos tenants simultâneos
- Binário único facilita deploy, rollback e operação
- Tipagem estática reduz bugs em produção
- Startup rápido permite reinicializações e deployments sem impacto perceptível

### Negativas

- Curva de aprendizado para desenvolvedores vindos de linguagens OO tradicionais (ausência de generics completos em versões antigas, sem herança)
- Verbosidade no tratamento de erros (`if err != nil` explícito) pode aumentar o volume de código
- Ecossistema de ORMs e frameworks menos maduro que Java/Spring ou PHP/Laravel (mitigado pela escolha do GORM)
- Reflexão limitada comparada a linguagens dinâmicas pode dificultar certos padrões de metaprogramação

---

## Riscos


| Risco                                                 | Probabilidade | Impacto | Severidade |
| ----------------------------------------------------- | ------------- | ------- | ---------- |
| Race condition em código multi-goroutine              | Média         | Alto    | Alta       |
| Goroutine leak por channel não fechado                | Média         | Médio   | Média      |
| Consumo excessivo de CPU em operações bcrypt em pico  | Média         | Médio   | Média      |
| Vulnerabilidade em dependência externa (supply chain) | Baixa         | Alto    | Média      |
| Memory leak por referências circulares não detectadas | Baixa         | Médio   | Baixa      |


---

## Mitigações

### Race conditions

- Executar testes com flag `-race` obrigatória na CI: `go test -race ./...`
- Code review com foco em padrões de acesso concorrente a mapas e slices
- Preferir `sync.Map` para mapas acessados concorrentemente, ou `sync.RWMutex` explícito
- Usar channels para comunicação entre goroutines ao invés de memória compartilhada quando possível

### Goroutine leaks

- Usar `context.Context` com timeout/cancelamento em todas as goroutines de I/O
- Adotar ferramentas de análise como `goleak` nos testes de integração
- Monitorar `runtime.NumGoroutine()` via métricas Prometheus com alerta para crescimento anormal

### CPU em bcrypt

- Configurar custo do bcrypt (cost factor) adequado para o hardware da VPS — recomendado cost 12-14
- Implementar rate limiting no endpoint de login para prevenir uso de bcrypt como vetor de DoS
- Considerar worker pool dedicado para operações bcrypt com limite máximo de goroutines simultâneas

### Supply chain

- Usar `go mod verify` na CI para verificar integridade de dependências
- Manter `go.sum` versionado no repositório
- Revisar e atualizar dependências mensalmente com `go get -u` controlado
- Preferir dependências com baixo número de transitivas

### Memory leaks

- Implementar profiling periódico com pprof em staging
- Configurar alerts de uso de memória no monitoramento da VPS

---

## Alternativas Consideradas

### Node.js (TypeScript)

- **Prós:** Ecossistema imenso, mesmo paradigma do frontend Angular, async/await nativo
- **Contras:** Event loop single-threaded limita CPU-bound (bcrypt), gerenciamento de memória menos previsível, runtime Node.js obrigatório na VPS, vulnerabilidades de supply chain frequentes no npm, performance inferior em operações criptográficas

### Java (Spring Boot)

- **Prós:** Ecossistema enterprise maduro, Spring Security para auth, JPA/Hibernate
- **Contras:** Alto consumo de memória (300MB+ em idle), startup lento (5–30s), JVM obrigatória na VPS, overhead desnecessário para um serviço focado

### Python (FastAPI)

- **Prós:** Desenvolvimento rápido, async nativo com FastAPI
- **Contras:** GIL (Global Interpreter Lock) limita paralelismo real, performance inferior, necessita runtime e virtualenv, tipagem dinâmica aumenta risco de bugs em produção

### Rust

- **Prós:** Performance máxima, segurança de memória garantida em compile-time, zero-cost abstractions
- **Contras:** Curva de aprendizado extremamente alta (borrow checker), ecossistema de auth/JWT menos maduro, produtividade de desenvolvimento significativamente menor

**Go representa o melhor equilíbrio entre performance, segurança, produtividade e operabilidade para este caso de uso específico.**

---

## Referências

- [The Go Programming Language Specification](https://go.dev/ref/spec)
- [Go Memory Model](https://go.dev/ref/mem)
- [Concurrency Patterns in Go](https://go.dev/blog/pipelines)
- [Go Race Detector](https://go.dev/doc/articles/race_detector)

