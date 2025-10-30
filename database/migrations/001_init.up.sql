CREATE EXTENSION IF NOT EXISTS pgcrypto;


-- ========================================
-- DROP TABLES (for reset)
-- ========================================
DROP TABLE IF EXISTS ratings CASCADE;
DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS chat_messages CASCADE;
DROP TABLE IF EXISTS user_progress CASCADE;
DROP TABLE IF EXISTS manga_genres CASCADE;
DROP TABLE IF EXISTS genres CASCADE;
DROP TABLE IF EXISTS manga CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- ========================================
-- USERS
-- ========================================
CREATE TABLE users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
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
    id BIGSERIAL PRIMARY KEY,
    slug TEXT UNIQUE,
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
    id BIGSERIAL PRIMARY KEY,
    manga_id BIGINT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    UNIQUE (manga_id, genre_id)
);

-- ========================================
-- USER PROGRESS
-- ========================================
CREATE TABLE user_progress (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manga_id BIGINT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    current_chapter INTEGER DEFAULT 0,
    status TEXT CHECK(status IN ('reading', 'completed', 'plan_to_read', 'dropped')),
    rating INTEGER CHECK(rating >= 1 AND rating <= 10),
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- CHAT MESSAGES
-- ========================================
CREATE TABLE chat_messages (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    username TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- USER SESSIONS
-- ========================================
CREATE TABLE user_sessions (
    id BIGSERIAL PRIMARY KEY,
    token TEXT UNIQUE NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- REFRESH TOKENS
-- ========================================
CREATE TABLE refresh_tokens (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT UNIQUE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- COMMENTS
-- ========================================
CREATE TABLE comments (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manga_id BIGINT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ========================================
-- RATINGS
-- ========================================
CREATE TABLE ratings (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manga_id BIGINT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    rating INTEGER CHECK(rating >= 1 AND rating <= 10),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, manga_id)
);
