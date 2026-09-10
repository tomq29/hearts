-- +goose Up
-- These columns were nullable with no default, so "this person has no photos"
-- was stored as NULL rather than an empty list. Go reads NULL as a nil slice,
-- which marshals to JSON null, and clients that index into the array crash.
-- An empty list is what the absence of photos actually means.
UPDATE profiles SET photos = '{}' WHERE photos IS NULL;
UPDATE profiles SET self_described_flaws = '{}' WHERE self_described_flaws IS NULL;
UPDATE profiles SET self_described_strengths = '{}' WHERE self_described_strengths IS NULL;

ALTER TABLE profiles ALTER COLUMN photos SET DEFAULT '{}';
ALTER TABLE profiles ALTER COLUMN self_described_flaws SET DEFAULT '{}';
ALTER TABLE profiles ALTER COLUMN self_described_strengths SET DEFAULT '{}';

ALTER TABLE profiles ALTER COLUMN photos SET NOT NULL;
ALTER TABLE profiles ALTER COLUMN self_described_flaws SET NOT NULL;
ALTER TABLE profiles ALTER COLUMN self_described_strengths SET NOT NULL;

-- +goose Down
ALTER TABLE profiles ALTER COLUMN photos DROP NOT NULL;
ALTER TABLE profiles ALTER COLUMN self_described_flaws DROP NOT NULL;
ALTER TABLE profiles ALTER COLUMN self_described_strengths DROP NOT NULL;

ALTER TABLE profiles ALTER COLUMN photos DROP DEFAULT;
ALTER TABLE profiles ALTER COLUMN self_described_flaws DROP DEFAULT;
ALTER TABLE profiles ALTER COLUMN self_described_strengths DROP DEFAULT;
