-- We use the pgcrypto extension to generate UUIDs for our primary keys.
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =================================================================
-- Users Table
-- Stores authentication and basic user information.
-- =================================================================
CREATE TABLE users (
    -- 'id' is the primary key, using a universally unique identifier (UUID).
    -- gen_random_uuid() comes from the pgcrypto extension.
    -- UUIDs are great because they are unique across all tables and databases.
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- 'email' must be unique for each user. We use VARCHAR(255) as a reasonable limit.
    email VARCHAR(255) UNIQUE NOT NULL,
    -- 'username' must also be unique.
    username VARCHAR(50) UNIQUE NOT NULL,
    -- 'password_hash' stores the securely hashed password.
    -- We use TEXT because hash lengths can vary, but are generally consistent.
    password_hash TEXT NOT NULL,
    -- 'refresh_token_hash' stores a secure hash of the user's current refresh token.
    -- It is nullable because a user might not have an active session.
    refresh_token_hash TEXT,
    -- 'refresh_token_expires_at' stores the expiry date of the refresh token.
    refresh_token_expires_at TIMESTAMPTZ,
    -- 'created_at' stores when the user account was created.
    -- TIMESTAMPTZ stores the timestamp with a time zone.
    -- 'NOT NULL' means this field cannot be empty.
    -- 'DEFAULT NOW()' automatically sets the current time when a new user is created.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =================================================================
-- Profiles Table
-- Stores the detailed profile information linked to a user.
-- =================================================================
CREATE TABLE profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- 'user_id' is a foreign key that links this profile to a user in the 'users' table.
    -- This creates a one-to-one relationship.
    -- 'ON DELETE CASCADE' means if a user is deleted, their profile will be deleted too.
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    first_name VARCHAR(100) NOT NULL,
    -- 'TEXT' is used for longer strings with no specific length limit, like a bio.
    bio TEXT,
    -- 'TEXT[]' is a PostgreSQL-specific type for an array of strings.
    -- This is perfect for storing a list of photos or flaws/strengths.
    photos TEXT [],
    self_described_flaws TEXT [],
    self_described_strengths TEXT [],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- 'updated_at' will store when the profile was last modified.
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create an index on the user_id for faster lookups.
CREATE INDEX idx_profiles_user_id ON profiles(user_id);

-- =================================================================
-- Likes Table
-- Stores user interactions (likes/dislikes).
-- =================================================================
CREATE TABLE likes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_like BOOLEAN NOT NULL DEFAULT TRUE,
    -- TRUE for Like, FALSE for Pass
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Ensure a user can't like/pass the same person twice
    UNIQUE(from_user_id, to_user_id),
    -- Ensure a user can't like themselves
    CHECK (from_user_id <> to_user_id)
);

CREATE INDEX idx_likes_from_user ON likes(from_user_id);

CREATE INDEX idx_likes_to_user ON likes(to_user_id);

-- =================================================================
-- Reviews Table
-- Stores feedback left by users about each other.
-- =================================================================
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (
        rating >= 1
        AND rating <= 5
    ),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Ensure a user can't review themselves
    CHECK (author_id <> target_id)
);

CREATE INDEX idx_reviews_target_id ON reviews(target_id);

CREATE INDEX idx_reviews_author_id ON reviews(author_id);
-- =================================================================
-- Notifications Table
-- Stores notifications for users (e.g., "You have a new match!").
-- =================================================================
CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL, -- e.g., 'match', 'system'
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
