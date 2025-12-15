-- Rollback MangaDex sync schema extensions

-- Drop triggers
DROP TRIGGER IF EXISTS update_manga_updated_at ON manga;
DROP TRIGGER IF EXISTS update_chapters_updated_at ON chapters;
DROP TRIGGER IF EXISTS update_sync_state_updated_at ON sync_state;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop sync_state table
DROP TABLE IF EXISTS sync_state;

-- Remove indexes
DROP INDEX IF EXISTS idx_manga_mangadex_id;
DROP INDEX IF EXISTS idx_manga_last_synced;
DROP INDEX IF EXISTS idx_manga_last_chapter_check;
DROP INDEX IF EXISTS idx_chapters_mangadex_id;
DROP INDEX IF EXISTS idx_chapters_published_at;

-- Remove columns from chapters
ALTER TABLE chapters DROP COLUMN IF EXISTS mangadex_chapter_id;
ALTER TABLE chapters DROP COLUMN IF EXISTS volume;
ALTER TABLE chapters DROP COLUMN IF EXISTS pages;
ALTER TABLE chapters DROP COLUMN IF EXISTS updated_at;

-- Rename published_at back to release_date
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='chapters' AND column_name='published_at') THEN
        ALTER TABLE chapters RENAME COLUMN published_at TO release_date;
    END IF;
END $$;

-- Remove columns from manga
ALTER TABLE manga DROP COLUMN IF EXISTS mangadex_id;
ALTER TABLE manga DROP COLUMN IF EXISTS last_synced_at;
ALTER TABLE manga DROP COLUMN IF EXISTS last_chapter_check;
ALTER TABLE manga DROP COLUMN IF EXISTS updated_at;
