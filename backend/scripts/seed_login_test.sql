-- Seed para teste do endpoint de login
-- Executar: psql $DATABASE_URL -f scripts/seed_login_test.sql
-- UUIDs devem estar no formato: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (8-4-4-4-12 hex)

-- Tenant
INSERT INTO wisa_crm_db.tenants (id, slug, name, tax_id, type, status)
VALUES (
  '08940000-0000-0000-0000-000000000000'::uuid,
  'lingerie-maria',
  'Lingerie Maria',
  '12345678901234',
  'company',
  'active'
)
ON CONFLICT (slug) DO NOTHING;

-- Product
INSERT INTO wisa_crm_db.products (id, slug, name, status)
VALUES (
  '08790000-0000-0000-0000-000000000000'::uuid,
  'gestao-pocket',
  'Gestão Pocket',
  'active'
)
ON CONFLICT (slug) DO NOTHING;

-- Subscription
INSERT INTO wisa_crm_db.subscriptions (tenant_id, product_id, type, status, start_date, end_date)
SELECT 
  '08940000-0000-0000-0000-000000000000'::uuid,
  '08790000-0000-0000-0000-000000000000'::uuid,
  'payment',
  'active',
  CURRENT_DATE,
  CURRENT_DATE + INTERVAL '1 year'
WHERE NOT EXISTS (
  SELECT 1 FROM wisa_crm_db.subscriptions 
  WHERE tenant_id = '08940000-0000-0000-0000-000000000000'::uuid 
  AND product_id = '08790000-0000-0000-0000-000000000000'::uuid
);

-- User (senha: senha123)
INSERT INTO wisa_crm_db.users (id, tenant_id, name, email, password_hash, status)
VALUES (
  '89552410-0000-0000-0000-000000000000'::uuid,
  '08940000-0000-0000-0000-000000000000'::uuid,
  'Willianna',
  'willianna@lingeries.com.br',
  '123456',
  'active'
)
ON CONFLICT (tenant_id, email) DO NOTHING;

-- UserProductAccess
INSERT INTO wisa_crm_db.user_product_access (user_id, product_id, tenant_id, access_profile)
SELECT 
  '89552410-0000-0000-0000-000000000000'::uuid,
  '08790000-0000-0000-0000-000000000000'::uuid,
  '08940000-0000-0000-0000-000000000000'::uuid,
  'admin'
WHERE NOT EXISTS (
  SELECT 1 FROM wisa_crm_db.user_product_access 
  WHERE user_id = '89552410-0000-0000-0000-000000000000'::uuid 
  AND product_id = '08790000-0000-0000-0000-000000000000'::uuid
);
