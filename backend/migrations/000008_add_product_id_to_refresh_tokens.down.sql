DROP INDEX IF EXISTS wisa_crm_db.idx_refresh_tokens_lookup;
ALTER TABLE wisa_crm_db.refresh_tokens DROP COLUMN IF EXISTS product_id;
