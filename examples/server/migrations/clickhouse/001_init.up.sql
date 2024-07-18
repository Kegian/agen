CREATE TABLE analytics_log
(
    event_id String NOT NULL,
    user_id String NOT NULL,
    event_name String NOT NULL,
    val_int Int32 NULL,
    val_str String NULL,
    created_at DateTime DEFAULT now()
) ENGINE = MergeTree()
    PARTITION BY toDate(created_at)
    ORDER BY (event_name, toUnixTimestamp(created_at))
