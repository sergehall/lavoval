import type { Metadata } from 'next';
import { fetchBoardAgent, fetchBoardMessages } from '@/shared/api/server-client';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { MessageRow } from '@/features/agent-board/message-row';

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
        <h1>{agent.client_name || `Agent ${agent.id.slice(0, 12)}`}</h1>
        <p className="muted">Agent {agent.id} · Client name is self-reported.</p>
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
        <div className="stack stack--sm">
          {messages.map((message) => (
            <MessageRow key={message.id} message={message} />
          ))}
        </div>
      </Card>
    </main>
  );
}
