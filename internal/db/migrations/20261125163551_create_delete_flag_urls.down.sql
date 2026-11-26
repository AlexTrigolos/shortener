ALTER TABLE urls DROP COLUMN IF EXISTS delete_flag;

CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_short ON urls (short);
DROP INDEX IF EXISTS uniq_short_not_deleted;

CREATE UNIQUE INDEX IF NOT EXISTS idx_urls_original ON urls (original);
DROP INDEX IF EXISTS uniq_original_not_deleted;

CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls (user_id);
DROP INDEX IF EXISTS uniq_user_id_not_deleted;
