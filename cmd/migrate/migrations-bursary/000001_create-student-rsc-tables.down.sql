-- Drop foreign key constraints
ALTER TABLE "students_siblings" DROP CONSTRAINT "students_siblings_student_id_foreign";
ALTER TABLE "students_emergency" DROP CONSTRAINT "students_emergency_student_id_foreign";
ALTER TABLE "students_personal" DROP CONSTRAINT "students_personal_student_id_foreign";
ALTER TABLE "students_institution" DROP CONSTRAINT "students_institution_student_id_foreign";
ALTER TABLE "students_sponsor" DROP CONSTRAINT "students_sponsor_student_id_foreign";
ALTER TABLE "students_guardian" DROP CONSTRAINT "students_guardian_student_id_foreign";

-- Drop tables
DROP TABLE IF EXISTS "students_emergency";
DROP TABLE IF EXISTS "students_sponsor";
DROP TABLE IF EXISTS "students_siblings";
DROP TABLE IF EXISTS "students_guardian";
DROP TABLE IF EXISTS "students_institution";
DROP TABLE IF EXISTS "students_personal";
DROP TABLE IF EXISTS "students";
