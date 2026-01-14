-- +goose Up
ALTER TABLE
    profiles
ADD
    COLUMN gender VARCHAR(50);

ALTER TABLE
    profiles
ADD
    COLUMN height INT;

-- +goose Down
ALTER TABLE
    profiles DROP COLUMN gender;

ALTER TABLE
    profiles DROP COLUMN height;