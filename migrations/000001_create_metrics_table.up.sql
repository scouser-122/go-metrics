CREATE TABLE metrics (
    id VARCHAR(255) NOT NULL,
    type VARCHAR(7) NOT NULL,
    delta BIGINT NULL,
    value DOUBLE PRECISION NULL,
    hash VARCHAR(511) NULL,
    PRIMARY KEY(id, type)
);