CREATE TABLE
    "system_users" (
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
        "role_id" INT NULL,
        "role_code" VARCHAR(255) NULL
    );

ALTER TABLE "system_users" ADD PRIMARY KEY ("id");

ALTER TABLE "system_users" ADD CONSTRAINT FK_USER FOREIGN KEY (role_id) REFERENCES ROLES (ID);

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Insert a default admin user
INSERT INTO "system_users" (
    "firstname",
    "lastname",
    "middlename",
    "email",
    "password",
    "blocked",
    "first_time_login",
    "activated",
    "role_id"
)
VALUES (
    'Admin',                
    'User',                 
    NULL,                   
    'admin@bursary.com',   
    ENCODE(DIGEST('Karanja@123', 'sha256'), 'hex')::bytea,
    FALSE,                  
    FALSE,                  
    TRUE,                   
    1                      
);

UPDATE "system_users"
SET
    role_id = (
        SELECT
            id
        FROM
            roles
        WHERE
            name = 'admin'
    );


CREATE TABLE
    "bursaries" (
        "id" bigserial NOT NULL PRIMARY KEY,
        "bursary_name" VARCHAR(255) NOT NULL,
        "description" TEXT NULL,
        "end_date" DATE NOT NULL,
        "amount_allocated" NUMERIC(15, 2) NULL,
        "amount_per_student" NUMERIC(15, 2) NULL,
        "allocation_type" VARCHAR(50) NOT NULL CHECK ("allocation_type" IN ('fixed', 'variable')),
        "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
    );

CREATE TABLE
    "applications" (
        "id" bigserial NOT NULL,
        "bursary_id" BIGINT NOT NULL REFERENCES "bursaries"("id"),
        "student_id" BIGINT NOT NULL REFERENCES "students"("id"),
        "stage" VARCHAR(50) NOT NULL CHECK ("stage" IN ('submitted', 'ward', 'county', 'finance', 'ministry', 'disbursed')),
        "remarks" TEXT NULL,
        "soft_delete" BOOLEAN NULL DEFAULT FALSE,
        "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
        PRIMARY KEY ("id")
    );