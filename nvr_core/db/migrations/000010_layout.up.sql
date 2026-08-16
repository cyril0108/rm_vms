-- Create the layouts table
CREATE TABLE IF NOT EXISTS layouts (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    mode TEXT NOT NULL,       -- e.g., 'grid', 'carousel', 'focus'
    payload TEXT NOT NULL,    -- JSON string: e.g., '{"rows": 3, "cols": 3}'
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_layouts_user_id ON layouts(user_id);

-- Create the layout items table
CREATE TABLE IF NOT EXISTS layout_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    layout_id INTEGER NOT NULL,
    type TEXT NOT NULL,       -- e.g., 'camera', 'roi', 'webpage'
    payload TEXT NOT NULL,    -- JSON string: e.g., '{"camera_id": 1, "x": 0.2...}'

    -- Foreign key to ensure data integrity
    FOREIGN KEY (layout_id) REFERENCES layouts(id) ON DELETE CASCADE
);

-- Index for fast lookups when fetching a layout's items
CREATE INDEX idx_layout_items_layout_id ON layout_items(layout_id);