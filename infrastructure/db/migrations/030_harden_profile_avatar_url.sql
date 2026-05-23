-- Harden profile avatars against untrusted URL schemes, credentialed URLs,
-- and unapproved image origins. Invalid existing values are cleared so the
-- stricter constraint can be applied safely.

UPDATE lavoval_profiles
SET avatar_url = NULL
WHERE avatar_url IS NOT NULL
  AND NOT (
    length(avatar_url) <= 2048
    AND avatar_url ~* '^https://(avatars\.githubusercontent\.com|secure\.gravatar\.com|www\.gravatar\.com|lh3\.googleusercontent\.com)([/?#]|$)'
  );

ALTER TABLE lavoval_profiles
  DROP CONSTRAINT IF EXISTS profiles_avatar_url_format,
  ADD CONSTRAINT profiles_avatar_url_format
    CHECK (
      avatar_url IS NULL OR (
        length(avatar_url) <= 2048
        AND avatar_url ~* '^https://(avatars\.githubusercontent\.com|secure\.gravatar\.com|www\.gravatar\.com|lh3\.googleusercontent\.com)([/?#]|$)'
      )
    );
