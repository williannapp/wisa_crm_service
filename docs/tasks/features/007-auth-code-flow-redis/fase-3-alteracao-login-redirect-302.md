# Fase 3 — Alteração do Endpoint de Login (Redirect HTTP 302)

## Objetivo

Modificar o fluxo de login para: (1) gerar um authorization code aleatório com expiração de 40 segundos; (2) armazenar no Redis as informações necessárias para emitir o JWT; (3) responder com HTTP 302 redirecionando para o callback da aplicação cliente com `?code=...&state=...`. O JWT não é mais retornado diretamente — apenas o code na URL do redirect.

---

## Planejamento Passo a Passo (ReAct)

### Pensamento 1

O login atual (Feature 005) retorna JSON com `token` (ou `redirect_url` conforme implementação parcial). A nova especificação exige:
- Após validação de credenciais e assinatura, **não** emitir JWT imediatamente
- Gerar um code opaco e aleatório
- Armazenar no Redis (AuthCodeStore) os dados para gerar o JWT posteriormente
- Montar a redirect_url internamente: `https://{tenant_slug}.{base_domain}/{product_slug}/callback?code={code}&state={state}`
- Responder HTTP 302 com `Location: redirect_url`

O parâmetro `state` deve ser recebido no request de login — o cliente envia `state` ao redirecionar o usuário para o auth, e o backend o repassa na redirect_url para proteção CSRF.

### Ação 1.1

Verificar e, se necessário, adicionar o campo `state` ao `LoginRequest`:

```go
type LoginRequest struct {
    Slug        string `json:"slug" binding:"required"`
    ProductSlug string `json:"product_slug" binding:"required"`
    UserEmail   string `json:"user_email" binding:"required,email"`
    Password    string `json:"password" binding:"required,min=1"`
    State       string `json:"state"` // opcional; cliente deve enviar para CSRF
}
```

### Observação 1.1

O `state` é recomendado pelo OAuth 2.0 para CSRF. Se o cliente não enviar, o backend pode gerar um valor vazio ou omitir na redirect_url. O ADR-010 recomenda que o cliente valide o state no retorno. Para máxima segurança, exigir `state` (binding:"required") ou documentar que é obrigatório para clientes que implementam callback.

---

### Pensamento 2

A redirect_url é montada internamente. A ATA e ADR-010 especificam que a URL é construída a partir de `tenant_slug`, `product_slug` e um domínio base. O formato pode ser:
- `https://{tenant_slug}.{base_domain}/{product_slug}/callback` — subdomínio por tenant
- Ou `https://{base_domain}/{tenant_slug}/{product_slug}/callback` — path-based

O exemplo do usuário: `https://lingerie.wisa-labs.com/gestao-pocket/callback?code=XYZ&state=123`. Isso sugere `https://{tenant_slug}.{domain}/{product_slug}/callback`. A variável `JWT_AUD_BASE_DOMAIN` ou similar já existe para o claim `aud`; pode haver `REDIRECT_BASE_DOMAIN` ou usar a mesma. Verificar convenção do projeto.

### Ação 1.2

Usar configuração existente ou criar variável para o domínio base da redirect:
- Se `JWT_AUD_BASE_DOMAIN` = `wisa-labs.com`, a redirect seria `https://{slug}.wisa-labs.com/{product_slug}/callback`
- Validar que tenant_slug e product_slug vêm do fluxo validado (já verificados no Use Case) — evita injeção de URL maliciosa

### Observação 1.2

A construção deve usar **apenas** os valores validados de `tenant.Slug` e `product.Slug` (obtidos do banco após validação). Nunca usar valores brutos do request para montar a URL — isso previne Open Redirect.

---

### Pensamento 3

Geração do code: usar `crypto/rand` para 32 bytes, codificar em hex (64 caracteres). Alternativa: base64url (mais compacto). O importante é ser imprevisível e único. `encoding/hex` é simples.

```go
b := make([]byte, 32)
if _, err := rand.Read(b); err != nil {
    return err
}
code := hex.EncodeToString(b)
```

### Ação 1.3

Criar função utilitária `generateAuthCode() (string, error)` em `internal/usecase/auth/` ou em `pkg/crypto/`. O Use Case chama essa função antes de armazenar no Redis.

### Observação 1.3

Evitar `math/rand` — não é criptograficamente seguro. Sempre `crypto/rand`.

---

### Pensamento 4

O Use Case `AuthenticateUser` precisa ser alterado:
- **Antes:** Emitir JWT e retornar `LoginOutput{Token: token}`
- **Depois:** Após todas as validações, construir `AuthCodeData` com Subject, Audience, TenantID, UserAccessProfile; gerar code; chamar `authCodeStore.Store(ctx, code, data, 40)`
- Retornar `LoginOutput{RedirectURL: "...", Code: code}` ou apenas `RedirectURL` (o handler monta a URL ou recebe do Use Case)

A redirect_url completa pode ser montada no Use Case (recebendo o base domain e state como input) ou no Handler. Para manter o Use Case focado em regras de negócio, a montagem da URL pode ficar no Use Case desde que receba os parâmetros necessários — ou o Use Case retorna `Code`, `State`, `TenantSlug`, `ProductSlug` e o Handler monta a URL. A segunda opção separa melhor: o Use Case retorna dados; o Handler formata a resposta HTTP.

Decisão: O Use Case retorna `LoginOutput{RedirectURL: string}` já montada. O Use Case recebe `State` no input e tem acesso a `RedirectBaseDomain` (config). Assim, a lógica de montagem fica no Use Case, mas usa apenas dados validados. O Handler apenas faz o redirect 302.

### Ação 1.4

Alterar `LoginInput` para incluir `State string`.

Alterar `LoginOutput` para `RedirectURL string` (remover `Token`).

No `Execute` do Use Case, após validações e antes de retornar:
1. Gerar `code` com `generateAuthCode()`
2. Construir `AuthCodeData{Subject: user.ID, Audience: aud, TenantID: tenant.ID, UserAccessProfile: accessProfile}`
3. `authCodeStore.Store(ctx, code, data, 40)`
4. Montar `redirectURL := fmt.Sprintf("https://%s.%s/%s/callback?code=%s&state=%s", tenant.Slug, redirectBaseDomain, product.Slug, code, url.QueryEscape(input.State))`
5. Retornar `&LoginOutput{RedirectURL: redirectURL}`

### Observação 1.4

`url.QueryEscape(input.State)` evita que caracteres especiais no state quebrem a URL. O Use Case precisa de `redirectBaseDomain` — pode ser a mesma config que `audienceBaseDomain` ou uma variável separada. Verificar nomenclatura existente.

---

### Pensamento 5

O path do callback: o usuário especificou `https://lingerie.wisa-labs.com/gestao-pocket/callback`. Isso indica `/{product_slug}/callback`. Alguns fluxos usam `/callback` na raiz. Seguir o exemplo: `/{product_slug}/callback`.

### Ação 1.5

Formato da redirect_url: `https://{tenant_slug}.{base_domain}/{product_slug}/callback?code={code}&state={state}`

### Observação 1.5

Documentar na Fase 5 para os clientes que a rota esperada é `/{product_slug}/callback` ou ` /callback` conforme convenção de cada aplicação. O auth server apenas monta a URL com o padrão definido; clientes podem ajustar se usarem rota diferente (neste caso, seria necessário parâmetro extra no login para path do callback — por simplicidade, assumir padrão `/{product_slug}/callback`).

---

### Pensamento 6

O Handler deve responder com HTTP 302:
- `c.Redirect(302, out.RedirectURL)` (Gin)
- Ou `c.Header("Location", out.RedirectURL)` e `c.AbortWithStatus(302)`

O Gin oferece `c.Redirect(http.StatusFound, url)` que equivale a 302.

### Ação 1.6

No AuthHandler.Login, em caso de sucesso:
```go
c.Redirect(http.StatusFound, out.RedirectURL)
```

Remover qualquer `c.JSON` com token ou cookie de access token — o token será obtido via `/auth/token` pelo cliente.

### Observação 1.6

Para SPAs que usam fetch/HttpClient: se o frontend do auth fizer login via AJAX, o servidor retornaria 302 e o fetch seguiria o redirect automaticamente, resultando na resposta do cliente (callback). Isso pode não ser o comportamento desejado — o frontend do auth esperaria receber JSON. **Alternativa documentada:** o backend pode aceitar header `Accept: application/json` e, nesse caso, retornar 200 JSON `{ "redirect_url": "..." }` em vez de 302, permitindo que o frontend faça `window.location = response.redirect_url`. O usuário especificou 302 — implementar 302 como padrão. Se necessário, a alternativa JSON pode ser fase futura.

---

### Pensamento 7 (Tratamento de erros)

Se `authCodeStore.Store` falhar (ex.: Redis indisponível), o Use Case deve retornar um erro. O ErrorMapper pode mapear para 503 Service Unavailable ou 500. Criar `domain.ErrAuthCodeStorageUnavailable` ou usar um erro genérico. O importante é não expor detalhes internos ao cliente.

### Ação 1.7

Para erro de Redis no Store: retornar erro que mapeie para 503 (serviço temporariamente indisponível). O cliente pode retentar. Não diferenciar "Redis down" de outros erros internos na resposta.

### Observação 1.7

Logar o erro real internamente para diagnóstico. A resposta ao cliente deve ser genérica.

---

### Pensamento 8 (Wiring e dependências)

O `AuthenticateUserUseCase` precisa de uma nova dependência: `AuthCodeStore`. O construtor deve receber essa dependência. O main.go deve construir `RedisAuthCodeStore` e passá-lo para `NewAuthenticateUserUseCase`.

O Use Case também precisa de `redirectBaseDomain` — pode ser a mesma string que `audienceBaseDomain` ou uma config específica. Verificar se a URL de redirect e o claim `aud` usam o mesmo domínio. Geralmente sim: `lingerie.wisa-labs.com` para redirect e `lingerie.wisa-labs.com` como aud.

### Ação 1.8

Adicionar `authCodeStore service.AuthCodeStore` e `redirectBaseDomain string` ao AuthenticateUserUseCase. O redirectBaseDomain pode ser o mesmo valor de audienceBaseDomain (ambos vêm de `JWT_AUD_BASE_DOMAIN` ou equivalente).

### Observação 1.8

Manter uma única fonte de verdade para o domínio base. Se houver divergência futura entre aud e redirect (ex.: redirect para CDN), criar variável separada.

---

## Checklist de Implementação

- [ ] 1. Adicionar `state` ao LoginRequest (opcional ou obrigatório conforme política)
- [ ] 2. Criar `generateAuthCode()` com crypto/rand
- [ ] 3. Adicionar AuthCodeStore e redirectBaseDomain ao AuthenticateUserUseCase
- [ ] 4. No Execute: após validações, gerar code, construir AuthCodeData, Store no Redis com TTL 40s
- [ ] 5. Montar redirect_url com tenant_slug, product_slug, code, state
- [ ] 6. Alterar LoginOutput para RedirectURL apenas
- [ ] 7. No AuthHandler: responder com c.Redirect(302, out.RedirectURL)
- [ ] 8. Remover retorno de token e cookie de access_token do login
- [ ] 9. Tratar erro de Store (Redis) com mapeamento para 503
- [ ] 10. Atualizar main.go com nova dependência AuthCodeStore

---

## Conclusão

A Fase 3 transforma o login de "retorna JWT" para "retorna redirect com code". O fluxo segue o Authorization Code do OAuth 2.0. A redirect_url é montada internamente com dados validados, evitando Open Redirect. O code tem TTL de 40 segundos e é single-use (consumido na Fase 4).
