# Feature: Create a NGINX configuration to execute tests

## Seu papel:

Você é um Arquiteto de Software Senior, especialista em:

- Desenvolvimento Backend em Go (Golang)
- Desenvolvimento Frontend em Angular (v20+)
- Sistemas distribuídos
- Segurança de aplicações
- Arquiteturas SaaS multi-tenant
- Clean Architecture
- Infraestrutura em VPS Linux com NGINX

Você possui forte domínio de:
- Segurança de aplicações web
- Escalabilidade horizontal e vertical
- Controle de concorrência
- Hardening de servidores
- Boas práticas de JWT e autenticação baseada em chave pública/privada
- Modelagem de banco PostgreSQL em ambientes multi-tenants

## Antes de receber sua tarefa, faça:

- **PRIMEIRO:** Leia o documento @docs/context.md para entender o objetivo do projeto;
- Em seguida, você deve ler o arquivo @docs/tasks/TRACKER.md e entender quais features já foram implementadas;
- Ler o documento @docs/tasks/TRACKER.md, e entender como funciona o gerenciamento de implementações de novas features.
- Ler os documentos de guidelines existentes em @docs/code_guidelines.
- Ler os documentos de adrs existentes em @docs/adrs


## Entendido o que é este projeto, você deverá **pensar e planejar** detalhadamente a implementação dos seguintes items:

1. Crie uma folder separada para a config de NGINX
    - Na pasta separada crie um docker.yaml para rodar o NGINX
    - Leve em consideração essa config de roteamento:

        ```json

        # Configuração para o subdomínio AUTH
        server {
            listen 80;   # ou 443 ssl para HTTPS
            server_name auth.wisa.labs.com.br;

            # Frontend — SPA Angular
            location / {
                proxy_pass http://0.0.0.0:4200;   # no Docker: nome do serviço
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
            }

            # API — Backend Go
            location /api/ {
                proxy_pass http://0.0.0.0:8080;
                proxy_set_header Host $host;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
                proxy_set_header X-Forwarded-Proto $scheme;
            }

            # JWKS (chave pública)
            location /.well-known/ {
                proxy_pass http://0.0.0.0:8080;
                proxy_set_header Host $host;
            }
        }


        server {
            listen 80;
            server_name lingerie-maria.wisa.labs.com.br;

            location /gestao-pocket {
                proxy_pass http://0.0.0.0:4201;
            }

        }

        ```
    - Revise as configs de roteamento e faça alterações `SE JULGAR NECESSÁRIO`. Se alguma porta estiver errada ou se faltar alguma regra.
    - Não altere nenhum código dentro de `backend` e `frontend`
    - Adicione ao docker-compose

- IMPORTANTE
    - Não altere nenhum código dentro de `backend` e `frontend`


# Seu processamento:

1. Você deverá **OBRIGATORIAMENTE** para cada um dos itens solicitados acima:
    - Criar um planejando *PASSO-A-PASSO* sobre como fazer a implementação deste item. Seu pensamento deve seguir o estilo ReAct, alternando entre "Pensamento" (seu raciocínio) e "Ação" (Uma etapa ou verificação concreta que você realizaria no projeto para atingir o objetivo desejado). Após cada ação, escreva uma "observação" para registrar sua linha de raciocínio, checando se tais modificações estão no caminho correto para atingir o objetivo deste item, se as mesmas seguem os padrões deste projeto, se não gerariam falhas de segurança crítica etc. Após obter uma conclusão de implementação para este item, crie conforme orientações no documento @docs/tasks/README.md uma feature para este item;

    **IMPORTANTE:** Se uma feature for considerada grande, a divida em etapas, separando-as em pequenas tasks, sempre conforme orientado na documentação já estudada. Neste caso, tome como exemplo, a documentação da última feature existente, onde cada fase foi criado um documento separado.

    **IMPORTANTE:** Não implemente nenhum código na estrutura. Seu papel é criar **SOMENTE** o planejamento.

    - Repetir os passos até terminar todos os itens solicitados para implementação.

2. Garanta que o plano de implementação atenda os requisitos presentes no @docs/code_guidelines
3. Garanta que o plano de implementação atenda os requisitos presentes no @docs/adrs
