-- Create table for Permission
CREATE TABLE permissions (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    version INT
);

-- Create table for Group
CREATE TABLE groups (
    id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    version INT
);

-- Create table for Subject
CREATE TABLE subjects (
    id VARCHAR(255),
    group_id INT,
    PRIMARY KEY (id, group_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

-- Create table for Group Permission
CREATE TABLE group_permissions (
    group_id INT,
    permission_id INT,
    PRIMARY KEY (group_id, permission_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

