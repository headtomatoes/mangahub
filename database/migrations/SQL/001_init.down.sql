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