-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS cpu_dict (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(512) NOT NULL,
  price_per_minute DECIMAL(10, 5) NOT NULL DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS cpu_dict_name_key ON cpu_dict (LOWER(TRIM(name)));

CREATE TABLE IF NOT EXISTS pricing_config (
  id SERIAL PRIMARY KEY,
  base_per_minute DECIMAL(10, 5) NOT NULL DEFAULT 0.01,
  ram_per_gb_per_minute DECIMAL(10, 5) NOT NULL DEFAULT 0.002,
  storage_hdd_per_gb_per_minute DECIMAL(10, 5) NOT NULL DEFAULT 0.0001,
  storage_ssd_per_gb_per_minute DECIMAL(10, 5) NOT NULL DEFAULT 0.0003,
  internet_per_mbit_per_minute DECIMAL(10, 5) NOT NULL DEFAULT 0.00005
);

INSERT INTO pricing_config (base_per_minute, ram_per_gb_per_minute, storage_hdd_per_gb_per_minute, storage_ssd_per_gb_per_minute, internet_per_mbit_per_minute)
SELECT 0.01, 0.002, 0.0001, 0.0003, 0.00005
FROM (SELECT 1) AS one
WHERE NOT EXISTS (SELECT 1 FROM pricing_config LIMIT 1);

INSERT INTO cpu_dict (name, price_per_minute)
SELECT v.name, v.price_per_minute FROM (VALUES
  ('Apple M1 Pro', 0.08),
  ('Apple M2', 0.07),
  ('Apple M2 Pro', 0.10),
  ('Intel Core i5-12400', 0.04),
  ('Intel Core i7-12700', 0.06),
  ('Intel Core i9-12900K', 0.12),
  ('AMD Ryzen 5 5600X', 0.04),
  ('AMD Ryzen 7 5800X', 0.06),
  ('AMD Ryzen 9 5900X', 0.09)
) AS v(name, price_per_minute)
WHERE NOT EXISTS (SELECT 1 FROM cpu_dict c WHERE LOWER(TRIM(c.name)) = LOWER(TRIM(v.name)));

-- Ensure gpu_dict has price column (may already exist from sessionhandlerservice)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'gpu_dict' AND column_name = 'price') THEN
    ALTER TABLE gpu_dict ADD COLUMN price DECIMAL(10, 5) DEFAULT 0;
  END IF;
END $$;

-- Seed example GPUs with prices (BYN per minute); skip if name already exists
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'NVIDIA GeForce RTX 3060', 12288, 0.05, 150 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'NVIDIA GeForce RTX 3060');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'NVIDIA GeForce RTX 3070', 8192, 0.07, 200 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'NVIDIA GeForce RTX 3070');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'NVIDIA GeForce RTX 3080', 10240, 0.10, 280 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'NVIDIA GeForce RTX 3080');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'NVIDIA GeForce RTX 3090', 24576, 0.18, 350 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'NVIDIA GeForce RTX 3090');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'NVIDIA GeForce RTX 4060', 8192, 0.06, 180 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'NVIDIA GeForce RTX 4060');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'NVIDIA GeForce RTX 4070', 12288, 0.09, 250 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'NVIDIA GeForce RTX 4070');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'NVIDIA GeForce RTX 4080', 16384, 0.14, 320 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'NVIDIA GeForce RTX 4080');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'NVIDIA GeForce RTX 4090', 24576, 0.22, 400 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'NVIDIA GeForce RTX 4090');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'AMD Radeon RX 6700 XT', 12288, 0.05, 140 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'AMD Radeon RX 6700 XT');
INSERT INTO gpu_dict (name, total_vram, price, avg_dlperf) SELECT 'AMD Radeon RX 6800 XT', 16384, 0.08, 220 WHERE NOT EXISTS (SELECT 1 FROM gpu_dict WHERE TRIM(name) = 'AMD Radeon RX 6800 XT');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pricing_config;
DROP TABLE IF EXISTS cpu_dict;
-- Do not drop gpu_dict or alter it - may be shared
-- +goose StatementEnd
