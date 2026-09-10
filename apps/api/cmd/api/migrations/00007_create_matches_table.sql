-- +goose Up
-- Records that a pair matched, so the match is a fact the database owns rather
-- than something re-derived on every match-check message. The unique pair makes
-- notification idempotent: concurrent swipes, Kafka redeliveries and worker
-- restarts all collapse onto a single insert.
CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_a UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_b UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_a, user_b),
    CHECK (user_a < user_b)
);

CREATE INDEX idx_matches_user_a ON matches(user_a);

CREATE INDEX idx_matches_user_b ON matches(user_b);

-- Backfill pairs that already matched, so existing users are not notified again
-- the next time a match-check arrives for them.
INSERT INTO matches (user_a, user_b)
SELECT DISTINCT LEAST(l1.from_user_id, l1.to_user_id), GREATEST(l1.from_user_id, l1.to_user_id)
FROM likes l1
JOIN likes l2 ON l1.from_user_id = l2.to_user_id AND l1.to_user_id = l2.from_user_id
WHERE l1.is_like = TRUE AND l2.is_like = TRUE
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS matches;
