-- Users table
CREATE TABLE users (
    id TEXT PRIMARY KEY,           -- UUID
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Manga table  
CREATE TABLE manga (
    id TEXT PRIMARY KEY,           -- slug-based ID
    title TEXT NOT NULL,
    author TEXT,
    genres TEXT,                   -- JSON array as text
    status TEXT CHECK(status IN ('ongoing', 'completed', 'hiatus')),
    total_chapters INTEGER,
    description TEXT,
    cover_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User progress (many-to-many relationship)
CREATE TABLE user_progress (
    user_id TEXT,
    manga_id TEXT,
    current_chapter INTEGER DEFAULT 0,
    status TEXT CHECK(status IN ('reading', 'completed', 'plan_to_read', 'dropped')),
    rating INTEGER CHECK(rating >= 1 AND rating <= 10),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, manga_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);

-- Chat messages
CREATE TABLE chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    username TEXT NOT NULL,        -- denormalized for performance
    message TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User sessions (for JWT blacklisting)
CREATE TABLE user_sessions (
    token_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
