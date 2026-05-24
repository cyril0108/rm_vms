PRAGMA foreign_keys = ON;

-- ======================================
-- Cameras
-- ======================================
CREATE TABLE IF NOT EXISTS cameras (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT UNIQUE NOT NULL,             -- Deterministic UUIDv5 for cross-system sync
    name TEXT NOT NULL,

    -- ONVIF Specific State
    manufacturer TEXT,
    model TEXT,
    serial_number TEXT NOT NULL, -- NOT UNIQUE, handled via index/code

    -- Networking & Discovery
    ip_address TEXT NOT NULL,
    http_port INTEGER DEFAULT 80,
    type TEXT NOT NULL,

    mac_address TEXT,

    -- Authentication
    username TEXT,
    password_enc TEXT, 

    -- RTSP Data Plane
    stream_url TEXT NOT NULL,
    sub_stream_url TEXT,

    -- ONVIF Control Tokens
    onvif_profile_token TEXT,
    sub_stream_profile_token TEXT,
    supports_ptz INTEGER DEFAULT 0,

    -- Storage & State
    retention_gb_limit INTEGER,
    is_active INTEGER DEFAULT 1,

    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
)STRICT;

-- Prevents adding the exact same camera hardware twice, but allows 
-- two cheap cameras with blank serials if they have different IPs.
CREATE UNIQUE INDEX IF NOT EXISTS idx_cam_dedup ON cameras(serial_number, ip_address);


-- ======================================
-- Segment Recording Table
-- ======================================
CREATE TABLE IF NOT EXISTS segments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    camera_id  INTEGER NOT NULL,
    start_time INTEGER NOT NULL,   -- Unix timestamp (milliseconds)
    end_time   INTEGER NOT NULL,   -- Unix timestamp (milliseconds)
    file_path  TEXT NOT NULL,      -- e.g., '/storage/cam_01/main/2026-03-20/19/44_00.mp4'
    profile    TEXT NOT NULL,      -- e.g., 'main', 'sub'
    snapshot_path TEXT NOT NULL DEFAULT '',   -- e.g., '/storage/cam_01/snapshot/2026-03-20/19/44_00.jpg'
    size_bytes INTEGER NOT NULL,   -- Tracked for watermark calculations

    FOREIGN KEY (camera_id) REFERENCES cameras(id) ON DELETE CASCADE
);

-- Optimizes: SELECT * FROM segments WHERE camera_id = ? AND profile = ? AND start_time >= ? AND end_time <= ?
CREATE INDEX IF NOT EXISTS idx_segments_timeline ON segments(camera_id, profile, start_time, end_time);

-- Optimizes: SELECT * FROM segments ORDER BY start_time ASC LIMIT ?
CREATE INDEX IF NOT EXISTS idx_segments_pruning ON segments(start_time);



-- ======================================
-- Roles & Permissions
-- ======================================

-- ROLES TABLE
-- Represents the broad job function (e.g., 'admin', 'supervisor', 'guard', 'guest')
CREATE TABLE roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (id, name, description)
VALUES (1, "擁有者", "admin");

UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM roles) WHERE name = 'roles';


-- PERMISSIONS TABLE
-- The granular actions allowed in the system. 
-- Best practice is using colon-separated domain notation (e.g., 'camera:add', 'timeline:view')
CREATE TABLE permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT UNIQUE NOT NULL,
    description TEXT
);

INSERT INTO permissions (id, code, description)
VALUES
(1, "system", "常駐系統權限，admin only"),
(2, "user_manage", "使用者管理"),
(3, "user_no_self_manage", "無法修改自身角色"),
(4, "layout_all", "佈局操作"),
(5, "view_all_device", "裝置檢視權限"),
(6, "camera_configure", "攝影機設定"),
(7, "camera_ptz", "PTZ 控制"),
(8, "camera_playback", "錄影回放 / 搜尋"),
(9, "camera_live", "即時影像監看"),
(10, "recording_export", "錄影匯出");

UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM permissions) WHERE name = 'permissions';


-- ROLE_PERMISSIONS (Mapping Table)
-- Dynamically binds many permissions to a role.
-- If you want to change what a 'guard' can do, you add/remove rows here.
CREATE TABLE role_permissions (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
) WITHOUT ROWID;


-- ======================================
-- USERS TABLE
-- ======================================
-- The actual user accounts. Every user belongs to exactly ONE role.
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL, -- Never store plaintext! Use bcrypt/argon2 in Go
    role_id INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT 1, -- Allows instantly disabling a user without deleting their logs
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE RESTRICT
);

INSERT INTO users (username, password, role_id)
VALUES ("admin", "", 1);

UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM users) WHERE name = 'users';


-- USER_CAMERA_ACCESS (Resource-Level Mapping)
-- NVR Specific: Binds a user to specific cameras. 
-- If an admin role implicitly sees all cameras, you handle that in Go logic.
-- But for lower roles, if a row doesn't exist here, the camera is invisible to them.
CREATE TABLE user_camera_access (
    user_id INTEGER NOT NULL,
    camera_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, camera_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    FOREIGN KEY (camera_id) REFERENCES cameras (id) ON DELETE CASCADE
) WITHOUT ROWID;

-- USER_PERMISSIONS (Direct Grants / Exceptions)
-- Allows assigning specific actions directly to a user.
CREATE TABLE user_permissions (
    user_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, permission_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
) WITHOUT ROWID;



-- Part 1: Permissions from the User's Role
-- SELECT p.code 
-- FROM users u
-- JOIN role_permissions rp ON u.role_id = rp.role_id
-- JOIN permissions p ON rp.permission_id = p.id
-- WHERE u.id = ? AND u.is_active = 1

-- UNION

-- -- Part 2: Permissions granted directly to the User
-- SELECT p.code 
-- FROM users u
-- JOIN user_permissions up ON u.id = up.user_id
-- JOIN permissions p ON up.permission_id = p.id
-- WHERE u.id = ? AND u.is_active = 1;
