-- Migration inicial - baseline
-- Extensão para geração de UUIDs (nativo no PG13+ como gen_random_uuid())
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
