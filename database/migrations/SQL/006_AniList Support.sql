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