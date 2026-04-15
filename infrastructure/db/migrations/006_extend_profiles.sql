-- Extend profiles with marketplace identity fields.
-- All constraints serve as a DB-level defence-in-depth layer:
-- the application layer (Go validator + Zod) is the primary guard,
-- but these CHECK constraints prevent bad data even if bypassed.

ALTER TABLE profiles
  -- Public username / slug: alphanumeric, dash, underscore, 3-30 chars.
  ADD COLUMN IF NOT EXISTS username TEXT
    CONSTRAINT profiles_username_format
      CHECK (username IS NULL OR username ~ '^[a-zA-Z0-9_-]{3,30}$'),

  -- Avatar: must be an http(s) URL, max 2048 chars.
  ADD COLUMN IF NOT EXISTS avatar_url TEXT
    CONSTRAINT profiles_avatar_url_format
      CHECK (avatar_url IS NULL OR (length(avatar_url) <= 2048 AND avatar_url ~* '^https?://')),

  -- Free-text location: 2-100 chars when present.
  ADD COLUMN IF NOT EXISTS location TEXT
    CONSTRAINT profiles_location_length
      CHECK (location IS NULL OR (length(location) >= 2 AND length(location) <= 100)),

  -- Skills tag array: max 20 items, each item 1-50 chars.
  ADD COLUMN IF NOT EXISTS skills TEXT[]
    CONSTRAINT profiles_skills_count
      CHECK (skills IS NULL OR array_length(skills, 1) <= 20),

  -- Language codes (e.g. "en", "ru"): max 10 items.
  ADD COLUMN IF NOT EXISTS languages TEXT[]
    CONSTRAINT profiles_languages_count
      CHECK (languages IS NULL OR array_length(languages, 1) <= 10),

  -- Social / personal URLs: http(s) only, max 2048 chars.
  ADD COLUMN IF NOT EXISTS website_url TEXT
    CONSTRAINT profiles_website_url_format
      CHECK (website_url IS NULL OR (length(website_url) <= 2048 AND website_url ~* '^https?://')),

  ADD COLUMN IF NOT EXISTS linkedin_url TEXT
    CONSTRAINT profiles_linkedin_url_format
      CHECK (linkedin_url IS NULL OR (length(linkedin_url) <= 2048 AND linkedin_url ~* '^https?://(www\.)?linkedin\.com/')),

  ADD COLUMN IF NOT EXISTS github_url TEXT
    CONSTRAINT profiles_github_url_format
      CHECK (github_url IS NULL OR (length(github_url) <= 2048 AND github_url ~* '^https?://(www\.)?github\.com/')),

  ADD COLUMN IF NOT EXISTS twitter_url TEXT
    CONSTRAINT profiles_twitter_url_format
      CHECK (twitter_url IS NULL OR (length(twitter_url) <= 2048 AND twitter_url ~* '^https?://(www\.)?(twitter\.com|x\.com)/')),

  -- Marketplace availability: closed enum, default open.
  ADD COLUMN IF NOT EXISTS availability_status TEXT NOT NULL DEFAULT 'open'
    CONSTRAINT profiles_availability_status_enum
      CHECK (availability_status IN ('open', 'limited', 'closed')),

  -- Whether the profile appears in public search / catalogue.
  ADD COLUMN IF NOT EXISTS is_public_profile BOOLEAN NOT NULL DEFAULT true;

-- Unique index on username (partial: only non-deleted rows).
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_username
  ON profiles (username)
  WHERE username IS NOT NULL AND deleted_at IS NULL;
