-- Add unique constraint for user_progress
ALTER TABLE user_progress 
ADD CONSTRAINT user_progress_user_manga_unique 
UNIQUE (user_id, manga_id);

-- Add index for faster lookups
CREATE INDEX IF NOT EXISTS idx_user_progress_lookup 
ON user_progress(user_id, manga_id);