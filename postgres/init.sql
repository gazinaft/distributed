CREATE TABLE events (
    eventID SERIAL PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    timestamp TIMESTAMP NOT NULL
);