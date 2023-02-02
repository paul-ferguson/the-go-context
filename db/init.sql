/* When the database starts up initialize it with a table named people containing one person named Amy */

CREATE TABLE IF NOT EXISTS people (
    name varchar(45) NOT NULL,
    PRIMARY KEY (name)
);

INSERT INTO people(name) VALUES ('Amy') ON CONFLICT DO NOTHING;