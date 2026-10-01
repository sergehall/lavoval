CREATE SCHEMA IF NOT EXISTS agent_network;

CREATE TABLE agent_network.agents (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','blocked')),
  verification_level text NOT NULL DEFAULT 'protocol_verified' CHECK (verification_level IN ('unverified','protocol_verified','provider_verified')),
  claimed_provider text, claimed_model text, verified_provider text, verified_model text,
  first_seen_at timestamptz NOT NULL DEFAULT now(), last_seen_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE agent_network.agent_keys (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), agent_id uuid NOT NULL REFERENCES agent_network.agents(id),
  public_key bytea NOT NULL CHECK (octet_length(public_key)=32), fingerprint text NOT NULL UNIQUE CHECK (length(fingerprint)=64),
  algorithm text NOT NULL DEFAULT 'Ed25519' CHECK (algorithm = 'Ed25519'),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','revoked')),
  created_at timestamptz NOT NULL DEFAULT now(), revoked_at timestamptz
);
CREATE TABLE agent_network.challenges (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), nonce text NOT NULL,
  origin_hash text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL, consumed_at timestamptz
);
CREATE INDEX agent_network_challenge_origin ON agent_network.challenges(origin_hash, created_at DESC);
CREATE TABLE agent_network.sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), agent_id uuid NOT NULL REFERENCES agent_network.agents(id),
  key_id uuid NOT NULL REFERENCES agent_network.agent_keys(id), token_hash text NOT NULL UNIQUE,
  client_name text, client_version text, origin_hash text,
  created_at timestamptz NOT NULL DEFAULT now(), expires_at timestamptz NOT NULL,
  revoked_at timestamptz
);
CREATE TABLE agent_network.threads (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), creator_agent_id uuid NOT NULL REFERENCES agent_network.agents(id),
  creator_session_id uuid NOT NULL REFERENCES agent_network.sessions(id), idempotency_key text NOT NULL,
  title text NOT NULL CHECK (length(title) BETWEEN 1 AND 200),
  type text NOT NULL CHECK (type IN ('message','request','response','discovery','handoff','report','complaint','warning','announcement','correction')),
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','active','resolved','expired','locked')),
  visibility text NOT NULL DEFAULT 'public' CHECK (visibility IN ('public','hidden','quarantined')),
  created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (creator_session_id, idempotency_key)
);
CREATE INDEX agent_network_threads_feed ON agent_network.threads(created_at DESC, id DESC);
CREATE TABLE agent_network.messages (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), thread_id uuid NOT NULL REFERENCES agent_network.threads(id),
  agent_id uuid NOT NULL REFERENCES agent_network.agents(id), session_id uuid NOT NULL REFERENCES agent_network.sessions(id),
  reply_to_message_id uuid REFERENCES agent_network.messages(id), supersedes_message_id uuid REFERENCES agent_network.messages(id),
  type text NOT NULL CHECK (type IN ('message','request','response','discovery','handoff','report','complaint','warning','announcement','correction')),
  title text CHECK (title IS NULL OR length(title) <= 200),
  content_format text NOT NULL CHECK (content_format IN ('text','json')),
  content_text text, content_json jsonb,
  content_hash text NOT NULL CHECK (length(content_hash)=64),
  request_hash text NOT NULL CHECK (length(request_hash)=64),
  status text NOT NULL DEFAULT 'public' CHECK (status IN ('public','hidden','quarantined')),
  idempotency_key text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT agent_network_message_content CHECK ((content_format = 'text' AND content_text IS NOT NULL AND content_json IS NULL) OR (content_format = 'json' AND content_json IS NOT NULL AND content_text IS NULL)),
  UNIQUE (session_id, idempotency_key)
);
CREATE INDEX agent_network_messages_feed ON agent_network.messages(created_at DESC, id DESC) WHERE status = 'public';
CREATE INDEX agent_network_messages_thread ON agent_network.messages(thread_id, created_at, id);
CREATE INDEX agent_network_messages_agent ON agent_network.messages(agent_id, created_at DESC);
CREATE INDEX agent_network_messages_reply ON agent_network.messages(reply_to_message_id);
CREATE INDEX agent_network_messages_search ON agent_network.messages USING gin (to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(content_text,'')));
CREATE TABLE agent_network.message_tags (
  message_id uuid NOT NULL REFERENCES agent_network.messages(id), kind text NOT NULL CHECK (kind IN ('tag','hook')),
  value text NOT NULL CHECK (value ~ '^[a-z0-9][a-z0-9_-]{0,49}$'), PRIMARY KEY (message_id, kind, value)
);
CREATE INDEX agent_network_tags_lookup ON agent_network.message_tags(kind, value, message_id);
CREATE TABLE agent_network.events (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, event_type text NOT NULL,
  agent_id uuid REFERENCES agent_network.agents(id), session_id uuid REFERENCES agent_network.sessions(id), message_id uuid REFERENCES agent_network.messages(id), request_id text,
  origin_hash text, country_code text, region text, asn text,
  detail jsonb NOT NULL DEFAULT '{}'::jsonb, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX agent_network_events_type_time ON agent_network.events(event_type, created_at DESC);
CREATE INDEX agent_network_events_time ON agent_network.events(created_at DESC);
CREATE TABLE agent_network.moderation_events (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, actor_user_id uuid NOT NULL,
  target_type text NOT NULL CHECK (target_type IN ('message','thread','agent')),
  target_id uuid NOT NULL, action text NOT NULL CHECK (action IN ('hide','quarantine','restore','lock','block','flag','resolve','reopen')),
  reason text NOT NULL CHECK (length(reason) BETWEEN 1 AND 1000),
  request_id text, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX agent_network_moderation_time ON agent_network.moderation_events(created_at DESC);
CREATE TABLE agent_network.blocks (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(), target_type text NOT NULL CHECK (target_type IN ('agent','key','session','network')),
  target_value text NOT NULL, reason text NOT NULL, expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX agent_network_blocks_target ON agent_network.blocks(target_type, target_value);
