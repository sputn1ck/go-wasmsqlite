-- Add BLOB fields for testing binary data handling
ALTER TABLE users ADD COLUMN avatar BLOB;
ALTER TABLE users ADD COLUMN metadata BLOB;

-- Create a new table specifically for testing BLOB operations
CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    content_type TEXT,
    data BLOB NOT NULL,
    thumbnail BLOB,
    size INTEGER,
    checksum BLOB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_attachments_user_id ON attachments(user_id);
CREATE INDEX IF NOT EXISTS idx_attachments_created_at ON attachments(created_at);