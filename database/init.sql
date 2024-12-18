CREATE DATABASE logdb;

\c logdb

CREATE TABLE logs (
    id SERIAL PRIMARY KEY,
    log_time TIMESTAMP NOT NULL,
    log_level VARCHAR(10) NOT NULL,
    message TEXT NOT NULL
);
