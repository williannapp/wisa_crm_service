# Feature 011 — Tela de Login — Task Tracker

> Consulte o [TRACKER central](../../TRACKER.md) para visão geral.

## Legenda

- `[ ]` Pendente | `[~]` Em andamento | `[x]` Concluída | `[-]` Cancelada

## Status da Feature

| Fase/Task | Descrição | Progresso | Status |
|-----------|-----------|-----------|--------|
| Fase 1 | Design tokens e estilos base | 1/1 | Concluída |
| Fase 2 | Estrutura e rota da página de login | 1/1 | Concluída |
| Fase 3 | Layout da página (container, gradient, overlay) | 1/1 | Concluída |
| Fase 4 | Card e formulário (estrutura visual) | 1/1 | Concluída |
| Fase 5 | Componentes visuais, responsividade e acessibilidade | 1/1 | Concluída |

## Arquivos de Tasks

- [fase-1-design-tokens-estilos.md](./fase-1-design-tokens-estilos.md)
- [fase-2-estrutura-rota-login.md](./fase-2-estrutura-rota-login.md)
- [fase-3-layout-pagina.md](./fase-3-layout-pagina.md)
- [fase-4-card-formulario.md](./fase-4-card-formulario.md)
- [fase-5-componentes-polish.md](./fase-5-componentes-polish.md)

## Resumo das Tasks

- [x] Fase 1 — Definir design tokens (cores, fontes, variáveis CSS) baseados no protótipo Login-Wisa
- [x] Fase 2 — Criar componente LoginPage em `features/auth/login`, configurar rota lazy-loaded `/login`
- [x] Fase 3 — Implementar layout da página: container full-screen, gradient background, overlay de textura
- [x] Fase 4 — Implementar card com formulário: campos usuário/senha, botão, links (apenas design, sem lógica)
- [x] Fase 5 — Adicionar ícones, toggle senha, responsividade, acessibilidade e data-testid

## Escopo

- **Objetivo:** Criar a tela de login do sistema com foco exclusivo no **design**.
- **Fora do escopo:** Funcionalidade de autenticação (será implementada em features futuras).
- **Referência de design:** Protótipo em `Login-Wisa/client/src/pages/login.tsx`.
