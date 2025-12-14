DROP INDEX IF EXISTS idx_user_progress_lookup;
ALTER TABLE user_progress DROP CONSTRAINT IF EXISTS user_progress_user_manga_unique;