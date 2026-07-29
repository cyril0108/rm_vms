CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    message TEXT,
    payload TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Guarantee that the payload is always valid JSON (or null)
    CHECK (payload IS NULL OR json_valid(payload))
);