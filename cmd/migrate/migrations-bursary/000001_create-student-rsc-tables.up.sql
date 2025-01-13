CREATE TABLE
    "students" (
        "id" bigserial NOT NULL,
        "firstname" VARCHAR(255) NOT NULL,
        "lastname" VARCHAR(255) NOT NULL,
        "middlename" VARCHAR(255) NULL,
        "email" VARCHAR(255) NOT NULL,
        "password" bytea NOT NULL,
        "blocked" BOOLEAN NOT NULL,
        "first_time_login" BOOLEAN NOT NULL,
        "activated" BOOLEAN NOT NULL,
        "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
        "role_id" INT NULL DEFAULT 0
    );

ALTER TABLE "students" ADD PRIMARY KEY ("id");

ALTER TABLE "students" ADD CONSTRAINT "students_email_unique" UNIQUE ("email");

CREATE TABLE
    "students_personal" (
        "id" bigserial NOT NULL,
        "dob" DATE NOT NULL,
        "gender" VARCHAR(255) NOT NULL,
        "citizenship" VARCHAR(255) NOT NULL,
        "religion" VARCHAR(255) NOT NULL,
        "parental_status" VARCHAR(255) NOT NULL,
        "birth_cert_no" VARCHAR(255) NULL,
        "birth_town" VARCHAR(255) NULL,
        "birth_county" VARCHAR(255) NOT NULL,
        "birth_sub_county" VARCHAR(255) NOT NULL,
        "ward" VARCHAR(255) NOT NULL,
        "voters_card_no" VARCHAR(255) NULL,
        "residence" VARCHAR(255) NOT NULL,
        "id_number" BIGINT NULL,
        "phone" BIGINT NOT NULL,
        "kra_pin_no" VARCHAR(255) NULL,
        "passport_no" VARCHAR(255) NULL,
        "special_need" INT NOT NULL DEFAULT 0,
        "special_needs_type" VARCHAR(255) NULL,
        "student_id" BIGINT NOT NULL UNIQUE
    );

ALTER TABLE "students_personal" ADD PRIMARY KEY ("id");

CREATE TABLE
    "students_institution" (
        "id" bigserial NOT NULL,
        "inst_name" VARCHAR(255) NOT NULL,
        "inst_type" VARCHAR(255) NOT NULL,
        "category" VARCHAR(255) NULL,
        "telephone" BIGINT NOT NULL,
        "email" VARCHAR(255) NULL,
        "address" VARCHAR(255) NOT NULL,
        "inst_county" VARCHAR(255) NOT NULL,
        "inst_sub_county" VARCHAR(255) NOT NULL,
        "inst_ward" VARCHAR(255) NULL,
        "principal_name" VARCHAR(255) NOT NULL,
        "year_joined" BIGINT NOT NULL,
        "curr_class_level" VARCHAR(255) NOT NULL,
        "adm_no" VARCHAR(255) NOT NULL,
        "student_id" BIGINT NOT NULL UNIQUE,
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW (),
        "bank_name" VARCHAR(255) NOT NULL,
        "bank_branch" VARCHAR(255) NOT NULL,
        "bank_acc_name" VARCHAR(255) NOT NULL,
        "bank_acc_no" BIGINT NOT NULL
    );

ALTER TABLE "students_institution" ADD PRIMARY KEY ("id");

CREATE TABLE
    "students_guardian" (
        "id" bigserial NOT NULL,
        "title" VARCHAR(255) NOT NULL,
        "firstname" VARCHAR(255) NOT NULL,
        "lastname" VARCHAR(255) NOT NULL,
        "middlename" VARCHAR(255) NULL,
        "phone" BIGINT NOT NULL,
        "phone_alternate" BIGINT NULL,
        "email" VARCHAR(255) NULL,
        "id_number" BIGINT NOT NULL,
        "kra_pin_no" VARCHAR(255) NULL,
        "passport_no" VARCHAR(255) NULL,
        "alien_no" VARCHAR(255) NULL,
        "occupation" VARCHAR(255) NULL,
        "work_location" VARCHAR(255) NULL,
        "work_phone" BIGINT NULL,
        "relationship" VARCHAR(255) NOT NULL,
        "address" VARCHAR(255) NULL,
        "residence" VARCHAR(255) NOT NULL,
        "town" VARCHAR(255) NOT NULL,
        "county" VARCHAR(255) NOT NULL,
        "sub_county" VARCHAR(255) NOT NULL,
        "ward" VARCHAR(255) NULL,
        "voters_card_no" VARCHAR(255) NULL,
        "polling_station" VARCHAR(255) NULL,
        "student_id" BIGINT NOT NULL,
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
    );

ALTER TABLE "students_guardian" ADD PRIMARY KEY ("id");

CREATE TABLE
    "students_siblings" (
        "id" bigserial NOT NULL,
        "name" VARCHAR(255) NOT NULL,
        "inst_name" VARCHAR(255) NOT NULL,
        "year" BIGINT NOT NULL,
        "fees" BIGINT NOT NULL,
        "paid" BOOLEAN NOT NULL,
        "balance" BIGINT NOT NULL,
        "student_id" BIGINT NOT NULL,
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
    );

ALTER TABLE "students_siblings" ADD PRIMARY KEY ("id");

CREATE TABLE
    "students_sponsor" (
        "id" bigserial NOT NULL,
        "name" VARCHAR(255) NOT NULL,
        "sponsorship_type" VARCHAR(255) NOT NULL,
        "sponsorship_nature" VARCHAR(255) NOT NULL,
        "phone" BIGINT NOT NULL,
        "email" VARCHAR(255) NULL,
        "address" VARCHAR(255) NULL,
        "contact_person_name" VARCHAR(255) NULL,
        "contact_person_phone" BIGINT NULL,
        "student_id" BIGINT NOT NULL UNIQUE,
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
    );

ALTER TABLE "students_sponsor" ADD PRIMARY KEY ("id");

CREATE TABLE
    "students_emergency" (
        "id" bigserial NOT NULL,
        "firstname" VARCHAR(255) NOT NULL,
        "middlename" VARCHAR(255) NULL,
        "lastname" VARCHAR(255) NOT NULL,
        "phone" BIGINT NOT NULL,
        "email" VARCHAR(255) NULL,
        "id_number" BIGINT NOT NULL,
        "occupation" VARCHAR(255) NULL,
        "relationship" VARCHAR(255) NOT NULL,
        "residence" VARCHAR(255) NOT NULL,
        "town" VARCHAR(255) NULL,
        "work_place" VARCHAR(255) NULL,
        "work_phone" BIGINT NULL,
        "provided_by" VARCHAR(255) NULL,
        "student_id" BIGINT NOT NULL UNIQUE,
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
    );

ALTER TABLE "students_emergency" ADD PRIMARY KEY ("id");

ALTER TABLE "students_siblings" ADD CONSTRAINT "students_siblings_student_id_foreign" FOREIGN KEY ("student_id") REFERENCES "students" ("id");

ALTER TABLE "students_emergency" ADD CONSTRAINT "students_emergency_student_id_foreign" FOREIGN KEY ("student_id") REFERENCES "students" ("id");

ALTER TABLE "students_personal" ADD CONSTRAINT "students_personal_student_id_foreign" FOREIGN KEY ("student_id") REFERENCES "students" ("id");

ALTER TABLE "students_institution" ADD CONSTRAINT "students_institution_student_id_foreign" FOREIGN KEY ("student_id") REFERENCES "students" ("id");

ALTER TABLE "students_sponsor" ADD CONSTRAINT "students_sponsor_student_id_foreign" FOREIGN KEY ("student_id") REFERENCES "students" ("id");

ALTER TABLE "students_guardian" ADD CONSTRAINT "students_guardian_student_id_foreign" FOREIGN KEY ("student_id") REFERENCES "students" ("id");