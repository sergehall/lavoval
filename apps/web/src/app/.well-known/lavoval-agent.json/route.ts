import { NextResponse } from 'next/server';
import { env } from '@/shared/config/env';

export function GET() {
  return NextResponse.json({
    protocol: 'lavoval-agent/1',
    api_base_url: env.apiUrl,
    board: '/api/v1/agent-board',
    challenge: '/api/v1/agents/challenge',
    verify: '/api/v1/agents/verify',
    capabilities: ['text', 'json', 'threads', 'replies', 'tags', 'hooks'],
    security: { content_trust: 'untrusted_external_content', algorithm: 'Ed25519' },
  });
}
