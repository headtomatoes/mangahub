CREATE EXTENSION IF NOT EXISTS pgcrypto;



-- ========================================
-- USERS
-- ========================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- MANGA
-- ========================================
CREATE TABLE manga (
    id TEXT PRIMARY KEY,  -- slug-based ID
    title TEXT NOT NULL,
    author TEXT,
    status TEXT CHECK(status IN ('ongoing', 'completed', 'hiatus')),
    total_chapters INTEGER,
    description TEXT,
    cover_url TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_manga_title ON manga(title);
CREATE INDEX idx_manga_author ON manga(author);
CREATE INDEX idx_manga_status ON manga(status);

-- ========================================
-- GENRES
-- ========================================
CREATE TABLE genres (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- ========================================
-- MANGA ↔ GENRES (Many-to-Many)
-- ========================================
CREATE TABLE manga_genres (
    manga_id TEXT NOT NULL,
    genre_id INTEGER NOT NULL,
    PRIMARY KEY (manga_id, genre_id),
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE,
    FOREIGN KEY (genre_id) REFERENCES genres(id) ON DELETE CASCADE
);

-- ========================================
-- USER PROGRESS
-- ========================================
CREATE TABLE user_progress (
    user_id UUID NOT NULL,
    manga_id TEXT NOT NULL,
    current_chapter INTEGER DEFAULT 0,
    status TEXT CHECK(status IN ('reading', 'completed', 'plan_to_read', 'dropped')),
    rating INTEGER CHECK(rating >= 1 AND rating <= 10),
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, manga_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);

-- ========================================
-- CHAT MESSAGES
-- ========================================
CREATE TABLE chat_messages (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    username TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- ========================================
-- USER SESSIONS
-- ========================================
CREATE TABLE user_sessions (
    token_id TEXT PRIMARY KEY,
    user_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- ========================================
-- COMMENTS
-- ========================================
CREATE TABLE comments (
    comment_id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    manga_id TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);

-- ========================================
-- RATINGS
-- ========================================
CREATE TABLE ratings (
    user_id UUID NOT NULL,
    manga_id TEXT NOT NULL,
    rating INTEGER CHECK(rating >= 1 AND rating <= 10),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, manga_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (manga_id) REFERENCES manga(id) ON DELETE CASCADE
);
