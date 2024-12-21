-- Reverse: ALTER TABLE users to drop NOT NULL constraint
ALTER TABLE users IF EXISTS ALTER COLUMN role_id DROP NOT NULL;

-- Reverse: ALTER TABLE users to add DEFAULT
ALTER TABLE users IF EXISTS ALTER COLUMN role_id SET DEFAULT 1;

-- Reverse: Update `role_id` in users table to NULL (if allowed) or reset to previous value
UPDATE users
SET role_id = NULL; -- Replace with previous default value if applicable

-- Reverse: Delete the inserted roles
DELETE FROM roles
WHERE name IN ('user', 'admin');

-- Reverse: Drop the `role_id` column from the users table
ALTER TABLE users IF EXISTS DROP COLUMN role_id;
