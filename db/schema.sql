CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE jobs (
    id INTEGER PRIMARY KEY,
    absolute_url VARCHAR(767) NOT NULL UNIQUE,
    data JSON
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20251220202955');
