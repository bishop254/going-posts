-- Reverse the "applications" table
DROP TABLE IF EXISTS "applications" CASCADE;

-- Reverse the "bursaries" table
DROP TABLE IF EXISTS "bursaries" CASCADE;

-- Reverse the "system_users" table

-- Reverse the foreign key constraint on "system_users"
ALTER TABLE "system_users" DROP CONSTRAINT FK_USER;

-- Remove the default admin user
DELETE FROM "system_users"
WHERE "email" = 'admin@bursary.com';

-- Drop the "system_users" table
DROP TABLE IF EXISTS "system_users" CASCADE;
