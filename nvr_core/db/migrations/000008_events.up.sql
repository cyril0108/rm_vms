CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type string,
    message TEXT,
    create_datetime INTEGER NOT NULL
);
