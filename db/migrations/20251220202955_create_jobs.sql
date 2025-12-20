-- migrate:up
CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY,
    absolute_url VARCHAR(767) NOT NULL UNIQUE,
    data JSON
)

-- migrate:down
DROP TABLE jobs;
