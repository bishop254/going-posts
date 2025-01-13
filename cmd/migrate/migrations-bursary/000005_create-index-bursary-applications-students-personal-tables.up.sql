-- Enable the pg_trgm extension for text-based searches
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Indexes for bursaries table
CREATE INDEX idx_bursaries_bursary_name ON bursaries USING gin (bursary_name gin_trgm_ops); -- Full-text search on bursary name
CREATE INDEX idx_bursaries_description ON bursaries USING gin (description gin_trgm_ops); -- Full-text search on bursary name
CREATE INDEX idx_bursaries_allocation_type ON bursaries USING btree (allocation_type); -- Filtering by allocation type
CREATE INDEX idx_bursaries_end_date ON bursaries USING btree (end_date); -- Filtering by date range

-- Indexes for applications table
CREATE INDEX idx_applications_bursary_id ON applications USING btree (bursary_id); -- Join with bursaries
CREATE INDEX idx_applications_student_id ON applications USING btree (student_id); -- Join with students
CREATE INDEX idx_applications_stage ON applications USING btree (stage); -- Filtering by stage
CREATE INDEX idx_applications_created_at ON applications USING btree (created_at); -- Sorting/Filtering by creation date
CREATE INDEX idx_applications_updated_at ON applications USING btree (updated_at); -- Sorting/Filtering by update date

-- Indexes for students table
CREATE INDEX idx_students_firstname ON students USING gin (firstname gin_trgm_ops); -- Full-text search on first name
CREATE INDEX idx_students_lastname ON students USING gin (lastname gin_trgm_ops); -- Full-text search on last name
CREATE INDEX idx_students_email ON students USING btree (email); -- Ensure uniqueness and fast lookup by email
CREATE INDEX idx_students_blocked ON students USING btree (blocked); -- Filtering blocked users
CREATE INDEX idx_students_activated ON students USING btree (activated); -- Filtering by activation status
CREATE INDEX idx_students_created_at ON students USING btree (created_at); -- Sorting/Filtering by creation date
CREATE INDEX idx_students_updated_at ON students USING btree (updated_at); -- Sorting/Filtering by update date

-- Indexes for students_personal table
CREATE INDEX idx_students_personal_birth_county ON students_personal USING gin (birth_county gin_trgm_ops); -- Full-text search on birth county
CREATE INDEX idx_students_personal_birth_sub_county ON students_personal USING gin (birth_sub_county gin_trgm_ops); -- Full-text search on birth sub-county
CREATE INDEX idx_students_personal_residence ON students_personal USING gin (residence gin_trgm_ops); -- Full-text search on residence
CREATE INDEX idx_students_personal_id_number ON students_personal USING btree (id_number); -- Lookup by ID number
CREATE INDEX idx_students_personal_phone ON students_personal USING btree (phone); -- Lookup by phone number
CREATE INDEX idx_students_personal_student_id ON students_personal USING btree (student_id); -- Ensure uniqueness and join with students
