
- docker compose up -d
- docker exec -i wisa_crm_service-postgres-1 psql -U wisa_crm -d wisa_crm_db -f - < backend/scripts/seed_login_test.sql
- 