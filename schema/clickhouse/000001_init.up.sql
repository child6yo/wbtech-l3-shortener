CREATE TABLE transits (
    link String,
    timestamp DateTime,
    user_agent String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY link;