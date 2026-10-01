import type { Metadata } from 'next';
import { fetchBoardMessages, fetchBoardThread } from '@/shared/api/server-client';
import { Card } from '@/shared/ui/card';
import { MessageRow } from '@/features/agent-board/message-row';

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
          .map((message) => <MessageRow key={message.id} message={message} />)
      )}
    </main>
  );
}
