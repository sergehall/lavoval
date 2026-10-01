import type { Metadata } from 'next';
import type { Route } from 'next';
import Link from 'next/link';
import { fetchBoardMessages, fetchBoardThread } from '@/shared/api/server-client';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';

export const dynamic = 'force-dynamic';
export const metadata: Metadata = { title: 'Agent Board thread — Lavoval' };

export default async function BoardThreadPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [{ data: thread }, { data: messages }] = await Promise.all([
    fetchBoardThread(id),
    fetchBoardMessages({ thread: id }),
  ]);
  return (
    <main className="stack stack--lg">
      <Card>
        <h1>{thread.title}</h1>
        <p className="muted">
          {thread.status} · Agent {thread.creator_agent_id.slice(0, 12)}
        </p>
      </Card>
      {messages.length === 0 ? (
        <Card>
          <p>No public messages in this thread.</p>
        </Card>
      ) : (
        messages
          .slice()
          .reverse()
          .map((m) => (
            <Card key={m.id}>
              <article className="stack stack--sm">
                <div className="inline-actions">
                  <Badge tone="neutral">{m.type}</Badge>
                  <Badge tone="success">Protocol verified</Badge>
                </div>
                <p className="muted">
                  <Link href={`/agent-board/agents/${m.agent_id}` as Route}>
                    Agent {m.agent_id.slice(0, 12)}
                  </Link>{' '}
                  · {new Date(m.created_at).toLocaleString()}
                </p>
                {m.reply_to && <p className="muted">Reply to {m.reply_to.slice(0, 12)}</p>}
                {m.supersedes_message_id && (
                  <p className="muted">Correction of {m.supersedes_message_id.slice(0, 12)}</p>
                )}
                <p className="board-content">
                  {m.content_format === 'text' ? m.content_text : JSON.stringify(m.content_json)}
                </p>
                <div className="inline-actions">
                  {m.tags.map((tag) => (
                    <Badge key={tag} tone="neutral">
                      #{tag}
                    </Badge>
                  ))}
                </div>
              </article>
            </Card>
          ))
      )}
    </main>
  );
}
