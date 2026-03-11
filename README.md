
- docker compose up -d
- docker exec -i wisa_crm_service-postgres-1 psql -U wisa_crm -d wisa_crm_db -f - < backend/scripts/seed_login_test.sql


Rebuild Backend:

docker compose up -d backend


Acessar os logs no docker compose:

- docker compose logs -f backend

- docker compose logs -f test-app-backend


CREDENCIAIS

Email: willianna@lingeries.com.br
Senha: senha123
tenant_slug: lingerie-maria
product_slug: gestao-pocket


RODAR AS MIGRATIONS:

docker run --rm --network wisa_crm_service_default \
  -v "$(pwd)/backend/migrations:/migrations" \
  migrate/migrate \
  -path=/migrations \
  -database "postgres://wisa_crm:wisa_crm_secret@postgres:5432/wisa_crm_db?sslmode=disable" \
  up


docker compose exec -T postgres psql -U wisa_crm -d wisa_crm_db -f - < backend/scripts/seed_login_test.sql



https://lingerie-maria.wisa.labs.com.br/gestao-pocket/callback?code=af678a00301ecdfd1be7b328bb18875d6d9fd9287c65752398d1632c85a83117&state=e508105c7d58b47f6497f9e283f944a404a043c77ad2a17f01e845aae2069f6d