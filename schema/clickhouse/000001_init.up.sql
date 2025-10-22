CREATE TABLE transits (
    short String,
    timestamp DateTime,
    user_agent String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY short;