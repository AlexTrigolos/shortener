ALTER TABLE urls ADD COLUMN IF NOT EXISTS delete_flag BOOLEAN NOT NULL DEFAULT FALSE;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_short_not_deleted ON urls (short) WHERE delete_flag = FALSE;
DROP INDEX IF EXISTS idx_urls_short;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_original_not_deleted ON urls (original) WHERE delete_flag = FALSE;
DROP INDEX IF EXISTS idx_urls_original;

CREATE INDEX IF NOT EXISTS uniq_user_id_not_deleted ON urls (user_id) WHERE delete_flag = FALSE;
DROP INDEX IF EXISTS idx_urls_user_id;
