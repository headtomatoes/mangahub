-- Drop all indexes in order
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_manga_title;
DROP INDEX IF EXISTS idx_manga_author;
DROP INDEX IF EXISTS idx_manga_status;
DROP INDEX IF EXISTS idx_manga_average_rating;
DROP INDEX IF EXISTS idx_chat_messages_created_at;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_refresh_tokens_revoked;
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_comments_manga_id;
DROP INDEX IF EXISTS idx_ratings_manga_id;
DROP INDEX IF EXISTS idx_user_library_user_id;
DROP INDEX IF EXISTS idx_user_library_manga_id;
DROP INDEX IF EXISTS idx_notifications_user_id;
DROP INDEX IF EXISTS idx_notifications_read;
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_chapters_manga_id;

-- Drop MangaDex sync indexes
DROP INDEX IF EXISTS idx_manga_mangadex_id;
DROP INDEX IF EXISTS idx_manga_last_synced;
DROP INDEX IF EXISTS idx_manga_last_chapter_check;
DROP INDEX IF EXISTS idx_chapters_mangadex_id;
DROP INDEX IF EXISTS idx_chapters_published_at;

-- Drop AniList sync indexes
DROP INDEX IF EXISTS idx_manga_anilist_id;
DROP INDEX IF EXISTS idx_manga_anilist_last_synced;
DROP INDEX IF EXISTS idx_chapters_anilist_id;

-- Drop triggers
DROP TRIGGER IF EXISTS update_manga_updated_at ON manga;
DROP TRIGGER IF EXISTS update_chapters_updated_at ON chapters;
DROP TRIGGER IF EXISTS update_sync_state_updated_at ON sync_state;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop sync_state table BEFORE other tables
DROP TABLE IF EXISTS sync_state;

-- Drop all tables
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

-- Drop custom types
DROP TYPE IF EXISTS user_role;

-- Rollback MangaDex sync schema extensions

-- Rename published_at back to release_date
DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name='chapters' AND column_name='published_at') THEN
        ALTER TABLE chapters RENAME COLUMN published_at TO release_date;
    END IF;
END $$;

-- Remove columns from chapters
ALTER TABLE chapters DROP COLUMN IF EXISTS mangadex_chapter_id;
ALTER TABLE chapters DROP COLUMN IF EXISTS volume;
ALTER TABLE chapters DROP COLUMN IF EXISTS pages;
ALTER TABLE chapters DROP COLUMN IF EXISTS updated_at;

-- Remove columns from manga
ALTER TABLE manga DROP COLUMN IF EXISTS mangadex_id;
ALTER TABLE manga DROP COLUMN IF EXISTS last_synced_at;
ALTER TABLE manga DROP COLUMN IF EXISTS last_chapter_check;
ALTER TABLE manga DROP COLUMN IF EXISTS updated_at;
