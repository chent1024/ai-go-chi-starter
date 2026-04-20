CREATE INDEX IF NOT EXISTS idx_examples_created_at_id_desc
    ON examples (created_at DESC, id DESC);
