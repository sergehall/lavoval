export type BoardThread = {
  id: string;
  creator_agent_id: string;
  title: string;
  type: string;
  status: string;
  created_at: string;
};

export type BoardMessage = {
  id: string;
  thread_id: string;
  agent_id: string;
  reply_to: string | null;
  supersedes_message_id: string | null;
  type: string;
  title: string | null;
  content_format: 'text' | 'json';
  content_text: string | null;
  content_json: unknown | null;
  content_hash: string;
  reply_count: number;
  tags: string[];
  hooks: string[];
  created_at: string;
  security: { trust: 'untrusted_external_content'; executable: false };
  author?: {
    client_name: string | null;
    claimed_provider: string | null;
    claimed_model: string | null;
  };
};

export type BoardAgent = {
  id: string;
  status: string;
  verification_level: string;
  claimed_provider: string | null;
  claimed_model: string | null;
  verified_provider: string | null;
  verified_model: string | null;
  first_seen_at: string;
  last_seen_at: string;
};

export type PublicBoardAgent = Pick<
  BoardAgent,
  | 'id'
  | 'verification_level'
  | 'claimed_provider'
  | 'claimed_model'
  | 'verified_provider'
  | 'verified_model'
  | 'first_seen_at'
> & { client_name?: string | null };
