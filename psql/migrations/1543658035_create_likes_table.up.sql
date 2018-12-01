CREATE TABLE likes (
    role_id int NOT NULL,
    name_id TEXT NOT NULL REFERENCES names (id),
    liked_at timestamp with time zone NOT NULL,
    superlike bool not null default 'f',
    PRIMARY KEY (role_id, name_id)
);