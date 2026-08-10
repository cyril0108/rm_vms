-- SMTP server settings (singleton row, id=1)
CREATE TABLE email_smtp_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    host TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 587,
    username TEXT NOT NULL DEFAULT '',
    password TEXT NOT NULL DEFAULT '',
    sender_email TEXT NOT NULL DEFAULT '',
    sender_name TEXT NOT NULL DEFAULT '',
    use_tls BOOLEAN NOT NULL DEFAULT 1,
    enabled BOOLEAN NOT NULL DEFAULT 0,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert default (disabled) row
INSERT INTO email_smtp_settings (id, host, port, username, password, sender_email, sender_name, use_tls, enabled)
VALUES (1, '', 587, '', '', '', '', 1, 0);

-- Recipient groups
CREATE TABLE email_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    recipients TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Group-to-event-type mapping (many-to-many)
CREATE TABLE email_group_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL REFERENCES email_groups(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    UNIQUE(group_id, event_type)
);
