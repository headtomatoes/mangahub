-- Add room_id and user_name columns to chat_messages table
ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS room_id BIGINT;
ALTER TABLE chat_messages ADD COLUMN IF NOT EXISTS user_name TEXT;

-- Add foreign key constraint for room_id
ALTER TABLE chat_messages 
ADD CONSTRAINT fk_chat_messages_room_id 
FOREIGN KEY (room_id) REFERENCES manga(id) ON DELETE CASCADE;

-- Update existing columns to NOT NULL after adding defaults
UPDATE chat_messages SET room_id = 1 WHERE room_id IS NULL;
UPDATE chat_messages SET user_name = 'Unknown' WHERE user_name IS NULL;

ALTER TABLE chat_messages ALTER COLUMN room_id SET NOT NULL;
ALTER TABLE chat_messages ALTER COLUMN user_name SET NOT NULL;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_chat_messages_room_id ON chat_messages(room_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id ON chat_messages(user_id);

-- Add comment for documentation
COMMENT ON COLUMN chat_messages.room_id IS 'References manga ID as chat room ID';
COMMENT ON COLUMN chat_messages.user_name IS 'Cached username for display purposes';
