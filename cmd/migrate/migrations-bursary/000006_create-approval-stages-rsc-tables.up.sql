CREATE TABLE
    "county_applications" (
        "id" bigserial NOT NULL,
        "bursary_id" BIGINT NOT NULL REFERENCES "bursaries" ("id"),
        "student_id" BIGINT NOT NULL REFERENCES "students" ("id"),
        "stage" VARCHAR(50) NOT NULL DEFAULT 'submitted' CHECK ("stage" IN ('submitted', 'approved', 'rejected')),
        "remarks" TEXT NULL,
        "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW (),
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW (),
        PRIMARY KEY ("id")
    );

CREATE TABLE
    "min_finance_applications" (
        "id" bigserial NOT NULL,
        "bursary_id" BIGINT NOT NULL REFERENCES "bursaries" ("id"),
        "student_id" BIGINT NOT NULL REFERENCES "students" ("id"),
        "stage" VARCHAR(50) NOT NULL DEFAULT 'submitted' CHECK (
            "stage" IN (
                'submitted',
                'approved-fa',
                'rejected',
                'disbursed'
            )
        ),
        "remarks" TEXT NULL,
        "document_id" VARCHAR(255) NULL,
        "created_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW (),
        "updated_at" TIMESTAMP(0) WITHOUT TIME ZONE NOT NULL DEFAULT NOW (),
        PRIMARY KEY ("id")
    );