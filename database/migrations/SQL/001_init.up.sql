CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ========================================
-- CREATE CUSTOM TYPES
-- ========================================
CREATE TYPE user_role AS ENUM ('user', 'admin');

-- ========================================
-- DROP TABLES (for reset)
-- ========================================
DROP TABLE IF EXISTS chapters CASCADE;
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
DROP TABLE IF EXISTS user_library CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;

-- ========================================
-- USERS
-- ========================================
CREATE TABLE users (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    role user_role NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

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
    average_rating DECIMAL(3,2),
    cover_url TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_manga_title ON manga(title);
CREATE INDEX IF NOT EXISTS idx_manga_author ON manga(author);
CREATE INDEX IF NOT EXISTS idx_manga_status ON manga(status);
CREATE INDEX IF NOT EXISTS idx_manga_average_rating ON manga(average_rating DESC);

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
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manga_id BIGINT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    current_chapter INTEGER DEFAULT 0,
    status TEXT CHECK(status IN ('reading', 'completed', 'plan_to_read', 'dropped')),
    page INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, manga_id)
);

-- ========================================
-- CHAT MESSAGES
-- ========================================
CREATE TABLE chat_messages (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at DESC);

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
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_revoked ON refresh_tokens(revoked);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

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
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_manga_id ON comments(manga_id);

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
CREATE INDEX IF NOT EXISTS idx_ratings_manga_id ON ratings(manga_id);

-- ========================================
-- USER LIBRARY
-- ========================================
CREATE TABLE user_library (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    manga_id BIGINT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, manga_id)
);
CREATE INDEX IF NOT EXISTS idx_user_library_user_id ON user_library(user_id);
CREATE INDEX IF NOT EXISTS idx_user_library_manga_id ON user_library(manga_id);

-- ========================================
-- NOTIFICATIONS
-- ========================================
CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    manga_id BIGINT REFERENCES manga(id) ON DELETE CASCADE,
    title TEXT,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_read ON notifications(read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

-- ========================================
-- CHAPTERS
-- ========================================
CREATE TABLE chapters (
    id BIGSERIAL PRIMARY KEY,
    manga_id BIGINT NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    chapter_number DECIMAL(10,2) NOT NULL,
    title TEXT,
    release_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (manga_id, chapter_number)
);
CREATE INDEX IF NOT EXISTS idx_chapters_manga_id ON chapters(manga_id, chapter_number DESC);

-- ========================================
-- MANGADEX SYNC SCHEMA EXTENSIONS
-- ========================================

-- Extend manga table with MangaDex sync tracking
ALTER TABLE manga ADD COLUMN IF NOT EXISTS mangadex_id UUID UNIQUE;
ALTER TABLE manga ADD COLUMN IF NOT EXISTS last_synced_at TIMESTAMPTZ;
ALTER TABLE manga ADD COLUMN IF NOT EXISTS last_chapter_check TIMESTAMPTZ;
ALTER TABLE manga ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_manga_mangadex_id ON manga(mangadex_id);
CREATE INDEX IF NOT EXISTS idx_manga_last_synced ON manga(last_synced_at);
CREATE INDEX IF NOT EXISTS idx_manga_last_chapter_check ON manga(last_chapter_check);

-- Update existing chapters table to support MangaDex sync
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS mangadex_chapter_id UUID UNIQUE;
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS volume TEXT;
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS pages INTEGER;
-- ALTER TABLE chapters ADD COLUMN IF NOT EXISTS published_at TIMESTAMPTZ;
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP;

-- Rename release_date to published_at if it exists
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='chapters' AND column_name='release_date') THEN
        ALTER TABLE chapters RENAME COLUMN release_date TO published_at;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_chapters_mangadex_id ON chapters(mangadex_chapter_id);
CREATE INDEX IF NOT EXISTS idx_chapters_published_at ON chapters(published_at DESC);

-- Create sync_state table for tracking sync operations
CREATE TABLE IF NOT EXISTS sync_state (
    id SERIAL PRIMARY KEY,
    sync_type TEXT NOT NULL UNIQUE, -- 'initial_sync', 'new_manga_poll', 'chapter_check'
    last_run_at TIMESTAMPTZ,
    last_success_at TIMESTAMPTZ,
    last_cursor TEXT, -- For pagination/filtering (e.g., createdAtSince timestamp)
    status TEXT DEFAULT 'idle', -- 'idle', 'running', 'error'
    error_message TEXT,
    metadata JSONB, -- Store additional state (e.g., manga IDs to check)
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Initialize sync states
INSERT INTO sync_state (sync_type, status) VALUES 
    ('initial_sync', 'idle'),
    ('new_manga_poll', 'idle'),
    ('chapter_check', 'idle')
ON CONFLICT (sync_type) DO NOTHING;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_manga_updated_at BEFORE UPDATE ON manga
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chapters_updated_at BEFORE UPDATE ON chapters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_sync_state_updated_at BEFORE UPDATE ON sync_state
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- ANILIST SYNC SCHEMA EXTENSIONS
-- ========================================

-- Extend manga table with AniList sync tracking
ALTER TABLE manga ADD COLUMN IF NOT EXISTS anilist_id INTEGER UNIQUE;
ALTER TABLE manga ADD COLUMN IF NOT EXISTS anilist_last_synced_at TIMESTAMPTZ;
ALTER TABLE manga ADD COLUMN IF NOT EXISTS anilist_last_chapter_check TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_manga_anilist_id ON manga(anilist_id);
CREATE INDEX IF NOT EXISTS idx_manga_anilist_last_synced ON manga(anilist_last_synced_at);

-- Extend chapters table with AniList tracking
ALTER TABLE chapters ADD COLUMN IF NOT EXISTS anilist_chapter_id TEXT;
CREATE INDEX IF NOT EXISTS idx_chapters_anilist_id ON chapters(anilist_chapter_id);

-- Add AniList sync states
INSERT INTO sync_state (sync_type, status) VALUES 
    ('anilist_initial_sync', 'idle'),
    ('anilist_new_manga_poll', 'idle'),
    ('anilist_chapter_check', 'idle')
ON CONFLICT (sync_type) DO NOTHING;

-- Add comments for clarity
COMMENT ON COLUMN manga.anilist_id IS 'AniList API manga ID';
COMMENT ON COLUMN manga.anilist_last_synced_at IS 'Last time manga was synced from AniList';
COMMENT ON COLUMN manga.anilist_last_chapter_check IS 'Last time chapters were checked from AniList';