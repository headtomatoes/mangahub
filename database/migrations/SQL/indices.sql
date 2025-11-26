-- Add these to improve query performance:
CREATE INDEX IF NOT EXISTS idx_comments_manga_id ON comments(manga_id);
CREATE INDEX IF NOT EXISTS idx_ratings_manga_id ON ratings(manga_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_progress_status ON user_progress(status);