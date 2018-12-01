CREATE TABLE dislikes (
    role_id int NOT NULL,
    name_id TEXT NOT NULL REFERENCES names (id),
    disliked_first_at timestamp with time zone NOT NULL,
    disliked_last_at timestamp with time zone NOT NULL,
    disliked_times int not null,
    PRIMARY KEY (role_id, name_id)
);