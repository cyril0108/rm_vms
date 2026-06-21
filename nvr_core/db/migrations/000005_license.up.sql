CREATE TABLE IF NOT EXISTS licenses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- THE SOURCE OF TRUTH: The exact signed string uploaded by the user
    raw_token TEXT NOT NULL UNIQUE, 

    -- CACHED CLAIMS FOR THE UI (Do not trust these for backend enforcement)
    iss TEXT NOT NULL,
    aud TEXT NOT NULL,
    kind TEXT NOT NULL,          -- e.g., 'trial', 'perpetual', 'subscription'
    machine_id TEXT NOT NULL,    
    max_devices INTEGER NOT NULL DEFAULT 0,

    -- UNIX Epoch timestamps for easy UI countdowns
    issued_at INTEGER NOT NULL,  
    expires_at INTEGER NOT NULL, 

    -- Standard audit field
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for quickly finding active/expired licenses
CREATE INDEX idx_licenses_expires ON licenses(expires_at);