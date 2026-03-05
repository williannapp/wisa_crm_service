-- Seed para teste do endpoint de login
-- Executar: psql $DATABASE_URL -f scripts/seed_login_test.sql

-- Tenant
INSERT INTO wisa_crm_db.tenants (id, slug, name, tax_id, type, status)
VALUES (
  'a0000001-0001-0001-0001-000000000001'::uuid,
  'cliente1',
  'Cliente 1',
  '12345678901234',
  'company',
  'active'
)
ON CONFLICT (slug) DO NOTHING;

-- Product
INSERT INTO wisa_crm_db.products (id, slug, name, status)
VALUES (
  'a0000002-0002-0002-0002-000000000002'::uuid,
  'crm',
  'CRM',
  'active'
)
ON CONFLICT (slug) DO NOTHING;

-- Subscription
INSERT INTO wisa_crm_db.subscriptions (tenant_id, product_id, type, status, start_date, end_date)
SELECT 
  'a0000001-0001-0001-0001-000000000001'::uuid,
  'a0000002-0002-0002-0002-000000000002'::uuid,
  'payment',
  'active',
  CURRENT_DATE,
  CURRENT_DATE + INTERVAL '1 year'
WHERE NOT EXISTS (
  SELECT 1 FROM wisa_crm_db.subscriptions 
  WHERE tenant_id = 'a0000001-0001-0001-0001-000000000001'::uuid 
  AND product_id = 'a0000002-0002-0002-0002-000000000002'::uuid
);

-- User (senha: senha123)
INSERT INTO wisa_crm_db.users (id, tenant_id, name, email, password_hash, status)
VALUES (
  'a0000003-0003-0003-0003-000000000003'::uuid,
  'a0000001-0001-0001-0001-000000000001'::uuid,
  'Usuario Teste',
  'usuario@empresa.com',
  '$2b$12$BfAfPNAeBCrB9gbdjFn9IOhsTzy66BtwVzdsjrQ5XoqA.//2UHIJO',
  'active'
)
ON CONFLICT (tenant_id, email) DO NOTHING;

-- UserProductAccess
INSERT INTO wisa_crm_db.user_product_access (user_id, product_id, tenant_id, access_profile)
SELECT 
  'a0000003-0003-0003-0003-000000000003'::uuid,
  'a0000002-0002-0002-0002-000000000002'::uuid,
  'a0000001-0001-0001-0001-000000000001'::uuid,
  'admin'
WHERE NOT EXISTS (
  SELECT 1 FROM wisa_crm_db.user_product_access 
  WHERE user_id = 'a0000003-0003-0003-0003-000000000003'::uuid 
  AND product_id = 'a0000002-0002-0002-0002-000000000002'::uuid
);
