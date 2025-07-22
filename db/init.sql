CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label TEXT UNIQUE NOT NULL,
    hidden INTEGER NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label TEXT UNIQUE NOT NULL,
    group_id INTEGER REFERENCES groups(id) ON DELETE SET NULL,
    last_report INTEGER,
    hidden INTEGER NOT NULL,
    secret TEXT UNIQUE NOT NULl,

    arch TEXT,
    operating_system TEXT,
    cpu TEXT,
    version TEXT
) STRICT;

CREATE TABLE IF NOT EXISTS server_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,

    cpu REAL NOT NULL,
    memory_percent REAL NOT NULL,
    memory_used INTEGER NOT NULL,
    memory_total INTEGER NOT NULL,
    disk_percent REAL NOT NULL,
    disk_used INTEGER NOT NULL,
    disk_total INTEGER NOT NULL,
    network_out_rate INTEGER NOT NULL,
    network_in_rate INTEGER NOT NULL,
    swap_percent REAL NOT NULL,
    swap_used INTEGER NOT NULL,
    swap_total INTEGER NOT NULL,
    uptime INTEGER NOT NULL,
    load_1 REAL NOT NULL,
    load_5 REAL NOT NULL,
    load_15 REAL NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS server_metrics_10m (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,

    cpu REAL NOT NULL,
    memory_percent REAL NOT NULL,
    memory_used INTEGER NOT NULL,
    memory_total INTEGER NOT NULL,
    disk_percent REAL NOT NULL,
    disk_used INTEGER NOT NULL,
    disk_total INTEGER NOT NULL,
    network_out_rate INTEGER NOT NULL,
    network_in_rate INTEGER NOT NULL,
    swap_percent REAL NOT NULL,
    swap_used INTEGER NOT NULL,
    swap_total INTEGER NOT NULL,
    uptime INTEGER NOT NULL,
    load_1 REAL NOT NULL,
    load_5 REAL NOT NULL,
    load_15 REAL NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    created_at INTEGER NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT UNIQUE NOT NULL,
    value TEXT NOT NULL
) STRICT;

CREATE TABLE IF NOT EXISTS incidents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    started_at INTEGER NOT NULL,
    ended_at INTEGER,
    state TEXT NOT NULL
) STRICT;


CREATE TABLE IF NOT EXISTS services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label TEXT NOT NULL UNIQUE,
    type TEXT NOT NULL,
    host TEXT NOT NULL
) STRICT;


CREATE TABLE IF NOT EXISTS service_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at INTEGER NOT NULL,
    `timestamp` INTEGER NOT NULL,
    `from` INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    `to` INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    min INTEGER,
    max INTEGER,
    loss REAL NOT NULL,
    avg INTEGER,
    median INTEGER
);

CREATE TABLE IF NOT EXISTS server_services (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    server_id INTEGER NOT NULL REFERENCES servers(id) ON DELETE CASCADE,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    UNIQUE(server_id, service_id)
);

INSERT OR IGNORE INTO settings (key, value) VALUES
    ('public_url', 'http://localhost:3005'),
    ('site_name', 'pbb');