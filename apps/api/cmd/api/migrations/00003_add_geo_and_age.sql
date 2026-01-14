-- +goose Up
ALTER TABLE
    profiles
ADD
    COLUMN birth_date DATE;

ALTER TABLE
    profiles
ADD
    COLUMN latitude DOUBLE PRECISION;

ALTER TABLE
    profiles
ADD
    COLUMN longitude DOUBLE PRECISION;

-- +goose Down
ALTER TABLE
    profiles DROP COLUMN birth_date;

ALTER TABLE
    profiles DROP COLUMN latitude;

ALTER TABLE
    profiles DROP COLUMN longitude;