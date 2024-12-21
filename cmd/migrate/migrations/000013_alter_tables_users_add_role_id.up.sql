
INSERT INTO roles (name, description, level)
VALUES
    ('user', 'User role', 1);

INSERT INTO roles (name, description, level)
VALUES
    ('admin', 'Admin role', 2);
ALTER TABLE IF EXISTS users ADD COLUMN role_id INT REFERENCES roles (id) DEFAULT 1;

UPDATE users
SET role_id = (
    SELECT id
    FROM roles
    WHERE name = 'user'
);

ALTER TABLE users
ALTER COLUMN role_id DROP DEFAULT;

ALTER TABLE users
ALTER COLUMN role_id SET NOT NULL;
