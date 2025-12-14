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
