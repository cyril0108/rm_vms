CREATE TABLE bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_id INTEGER NOT NULL,
    user_id   INTEGER NOT NULL,
    start_time INTEGER NOT NULL, -- Unix timestamp (suggest milliseconds for NVR accuracy)
    duration  INTEGER NOT NULL, -- Duration in milliseconds or seconds
    message   TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Enforce relationships (assuming you have cameras and users tables)
    FOREIGN KEY(camera_id) REFERENCES cameras(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Crucial for timeline queries: "Get all bookmarks for Camera X between Time Y and Time Z"
CREATE INDEX idx_bookmarks_camera_time ON bookmarks(camera_id, start_time);