-- migrate:up
ALTER TABLE jobs
ADD COLUMN source VARCHAR(255) GENERATED ALWAYS AS (JSON_EXTRACT(data, '$.source')) VIRTUAL;

-- migrate:down
ALTER TABLE jobs
DROP COLUMN source;
