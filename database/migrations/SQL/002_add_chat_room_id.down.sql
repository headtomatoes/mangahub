-- Remove indexes
DROP INDEX IF EXISTS idx_chat_messages_room_id;
DROP INDEX IF EXISTS idx_chat_messages_user_id;

-- Remove foreign key constraint
ALTER TABLE chat_messages DROP CONSTRAINT IF EXISTS fk_chat_messages_room_id;

-- Remove columns
ALTER TABLE chat_messages DROP COLUMN IF EXISTS room_id;
ALTER TABLE chat_messages DROP COLUMN IF EXISTS user_name;
