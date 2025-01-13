-- Disable the pg_trgm extension
DROP EXTENSION IF EXISTS pg_trgm;

-- Drop indexes for bursaries table
DROP INDEX IF EXISTS idx_bursaries_bursary_name; -- Full-text search on bursary name
DROP INDEX IF EXISTS idx_bursaries_allocation_type; -- Filtering by allocation type
DROP INDEX IF EXISTS idx_bursaries_end_date; -- Filtering by date range

-- Drop indexes for applications table
DROP INDEX IF EXISTS idx_applications_bursary_id; -- Join with bursaries
DROP INDEX IF EXISTS idx_applications_student_id; -- Join with students
DROP INDEX IF EXISTS idx_applications_stage; -- Filtering by stage
DROP INDEX IF EXISTS idx_applications_created_at; -- Sorting/Filtering by creation date
DROP INDEX IF EXISTS idx_applications_updated_at; -- Sorting/Filtering by update date

-- Drop indexes for students table
DROP INDEX IF EXISTS idx_students_firstname; -- Full-text search on first name
DROP INDEX IF EXISTS idx_students_lastname; -- Full-text search on last name
DROP INDEX IF EXISTS idx_students_email; -- Ensure uniqueness and fast lookup by email
DROP INDEX IF EXISTS idx_students_blocked; -- Filtering blocked users
DROP INDEX IF EXISTS idx_students_activated; -- Filtering by activation status
DROP INDEX IF EXISTS idx_students_created_at; -- Sorting/Filtering by creation date
DROP INDEX IF EXISTS idx_students_updated_at; -- Sorting/Filtering by update date

-- Drop indexes for students_personal table
DROP INDEX IF EXISTS idx_students_personal_birth_county; -- Full-text search on birth county
DROP INDEX IF EXISTS idx_students_personal_birth_sub_county; -- Full-text search on birth sub-county
DROP INDEX IF EXISTS idx_students_personal_residence; -- Full-text search on residence
DROP INDEX IF EXISTS idx_students_personal_id_number; -- Lookup by ID number
DROP INDEX IF EXISTS idx_students_personal_phone; -- Lookup by phone number
DROP INDEX IF EXISTS idx_students_personal_student_id; -- Ensure uniqueness and join with students
