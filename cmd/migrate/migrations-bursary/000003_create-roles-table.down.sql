-- Reverse the NOT NULL constraint on the role_id column
ALTER TABLE STUDENTS
ALTER COLUMN role_id DROP NOT NULL;

-- Reverse the default value drop operation on the role_id column
ALTER TABLE STUDENTS
ALTER COLUMN role_id SET DEFAULT NULL;

-- Reverse the update operation
UPDATE STUDENTS
SET role_id = NULL;

-- Remove the foreign key constraint
ALTER TABLE STUDENTS DROP CONSTRAINT FK_USER;

-- Delete inserted rows
DELETE FROM roles WHERE name = 'ward';
DELETE FROM roles WHERE name = 'county';
DELETE FROM roles WHERE name = 'finance-assistant';
DELETE FROM roles WHERE name = 'finance';
DELETE FROM roles WHERE name = 'ministry';
DELETE FROM roles WHERE name = 'admin';
DELETE FROM roles WHERE name = 'student';

-- Drop the ROLES table
DROP TABLE IF EXISTS ROLES;
