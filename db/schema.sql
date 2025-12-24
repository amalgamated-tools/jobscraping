CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE jobs (
    id INTEGER PRIMARY KEY,
    absolute_url VARCHAR(767) NOT NULL UNIQUE,
    data JSON
, source VARCHAR(255) GENERATED ALWAYS AS (JSON_EXTRACT(data, '$.source')) VIRTUAL, source_id VARCHAR(255) GENERATED ALWAYS AS (JSON_EXTRACT(data, '$.source_id')) VIRTUAL);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20251220202955'),
  ('20251220204205'),
  ('20251220221021');
