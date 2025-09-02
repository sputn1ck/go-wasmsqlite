-- Drop the attachments table
DROP TABLE IF EXISTS attachments;

-- Note: SQLite doesn't support dropping columns directly
-- The avatar and metadata columns in users table would need table recreation
-- For simplicity, we'll just document this limitation
SELECT 'Note: avatar and metadata columns cannot be dropped from users table without recreating it';