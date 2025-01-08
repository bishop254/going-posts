CREATE TABLE
    IF NOT EXISTS ROLES (
        ID BIGSERIAL PRIMARY KEY,
        NAME VARCHAR(255) NOT NULL UNIQUE,
        DESCRIPTION TEXT,
        LEVEL INT NOT NULL DEFAULT 0
    );

INSERT INTO
    roles (name, description, level)
VALUES
    ('student', 'Student role', 1);

INSERT INTO
    roles (name, description, level)
VALUES
    ('admin', 'Super Admin role', 100);

INSERT INTO
    roles (name, description, level)
VALUES
    ('ministry', 'Ministry role', 6);

INSERT INTO
    roles (name, description, level)
VALUES
    ('finance', 'Finance role', 5);

INSERT INTO
    roles (name, description, level)
VALUES
    ('finance-assistant', 'Finance Assistant role', 4);

INSERT INTO
    roles (name, description, level)
VALUES
    ('county', 'County role', 3);

INSERT INTO
    roles (name, description, level)
VALUES
    ('ward', 'Ward role', 2);

ALTER TABLE STUDENTS ADD CONSTRAINT FK_USER FOREIGN KEY (role_id) REFERENCES ROLES (ID);

UPDATE STUDENTS
SET
    role_id = (
        SELECT
            id
        FROM
            roles
        WHERE
            name = 'student'
    );

ALTER TABLE STUDENTS
ALTER COLUMN role_id
DROP DEFAULT;

ALTER TABLE STUDENTS
ALTER COLUMN role_id
SET
    NOT NULL;