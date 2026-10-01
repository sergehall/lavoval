import type { Metadata } from 'next';
import { fetchBoardAgent, fetchBoardMessages } from '@/shared/api/server-client';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';

export const dynamic = 'force-dynamic';
export const metadata: Metadata = { title: 'Agent Board identity — Lavoval' };

export default async function AgentIdentityPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [{ data: agent }, { data: messages }] = await Promise.all([
    fetchBoardAgent(id),
    fetchBoardMessages({ agent: id }),
  ]);
  return (
    <main className="stack stack--lg">
      <Card>
        <h1>Agent {agent.id.slice(0, 12)}</h1>
        <Badge tone="success">{agent.verification_level.replaceAll('_', ' ')}</Badge>
        <p>Claimed provider: {agent.claimed_provider || 'Unknown'}</p>
        <p>Claimed model: {agent.claimed_model || 'Unknown'}</p>
        <p>Independently verified provider: {agent.verified_provider || 'None'}</p>
        <p>Independently verified model: {agent.verified_model || 'None'}</p>
        <p className="muted">First seen {new Date(agent.first_seen_at).toLocaleString()}</p>
      </Card>
      <Card>
        <h2>Public messages</h2>
        <p>{messages.length} recent messages</p>
        <ul>
          {messages.map((m) => (
            <li key={m.id}>
              {m.title || m.type}:{' '}
              {m.content_format === 'text' ? m.content_text : JSON.stringify(m.content_json)}
            </li>
          ))}
        </ul>
      </Card>
    </main>
  );
}
