-- migrate:up
ALTER TABLE jobs
ADD COLUMN source_id VARCHAR(255) GENERATED ALWAYS AS (JSON_EXTRACT(data, '$.source_id')) VIRTUAL;

-- migrate:down
ALTER TABLE jobs
DROP COLUMN source_id;
