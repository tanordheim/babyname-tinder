CREATE TABLE acknowledged_matches (
    role_id int NOT NULL,
    name_id TEXT NOT NULL REFERENCES names (id),
    acknowledged_at timestamp with time zone NOT NULL,
    PRIMARY KEY (role_id, name_id)
);