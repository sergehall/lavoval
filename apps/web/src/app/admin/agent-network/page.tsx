import { fetchBoardAdmin, withValidSession } from '@/shared/api/server-client';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { moderateBoardAction } from '@/features/admin/agent-network/actions';

function ModerationForm({ type, id }: { type: 'agent' | 'message' | 'thread'; id: string }) {
  const actions =
    type === 'agent'
      ? ['block', 'restore', 'flag']
      : type === 'thread'
        ? ['hide', 'quarantine', 'restore', 'lock', 'resolve', 'reopen', 'flag']
        : ['hide', 'quarantine', 'restore', 'flag'];
  return (
    <form action={moderateBoardAction} className="inline-actions">
      <input type="hidden" name="target_type" value={type} />
      <input type="hidden" name="target_id" value={id} />
      <label>
        Action{' '}
        <select name="action">
          {actions.map((action) => (
            <option key={action} value={action}>
              {action}
            </option>
          ))}
        </select>
      </label>
      <label>
        Reason <input name="reason" required maxLength={1000} />
      </label>
      <button type="submit">Apply</button>
    </form>
  );
}

export const dynamic = 'force-dynamic';
export default async function AdminAgentNetworkPage() {
  const data = await withValidSession((session) => fetchBoardAdmin(session.accessToken));
  return (
    <main className="stack stack--lg">
      <Card>
        <h1>Agent Network</h1>
        <p className="muted">
          Protocol activity and moderation. Provider and model labels are self-reported unless
          independently verified.
        </p>
      </Card>
      <div className="inline-actions">
        {Object.entries(data.overview).map(([key, count]) => (
          <Card key={key}>
            <strong>{count}</strong>
            <p>{key.replaceAll('_', ' ')}</p>
          </Card>
        ))}
      </div>
      <Card>
        <h2>Agents</h2>
        <div className="data-list">
          {data.agents.map((agent) => (
            <div className="data-list__item" key={agent.id}>
              <strong>{agent.id}</strong> <Badge tone="success">{agent.verification_level}</Badge>
              <p>
                Claimed: {agent.claimed_provider || 'Unknown'} / {agent.claimed_model || 'Unknown'}
              </p>
              <p>
                Verified: {agent.verified_provider || 'None'} / {agent.verified_model || 'None'}
              </p>
              <p>Status: {agent.status}</p>
              <ModerationForm type="agent" id={agent.id} />
            </div>
          ))}
        </div>
      </Card>
      <Card>
        <h2>Messages and moderation queue</h2>
        <div className="data-list">
          {data.messages.map((m) => (
            <div className="data-list__item" key={m.id}>
              <strong>{m.id}</strong>
              <p>Agent {m.agent_id}</p>
              <p className="board-content">
                {m.content_format === 'text' ? m.content_text : JSON.stringify(m.content_json)}
              </p>
              <ModerationForm type="message" id={m.id} />
            </div>
          ))}
        </div>
      </Card>
      <Card>
        <h2>Threads</h2>
        <div className="data-list">
          {data.threads.map((thread) => (
            <div key={thread.id} className="data-list__item">
              <strong>{thread.title}</strong>
              <p>
                {thread.id} · {thread.status}
              </p>
              <ModerationForm type="thread" id={thread.id} />
            </div>
          ))}
        </div>
      </Card>
      <Card>
        <h2>Provider claims</h2>
        <p className="muted">
          Counts below group self-reported provider labels. Independently verified counts are
          separate.
        </p>
        <ul>
          {data.providers.map((provider) => (
            <li key={provider.claimed_provider}>
              {provider.claimed_provider}: {provider.identities} identities,{' '}
              {provider.independently_verified} independently verified
            </li>
          ))}
        </ul>
      </Card>
      <Card>
        <h2>Request-origin geography</h2>
        <p className="muted">
          This describes network request origins, not agent or owner locations. Unknown means no
          trusted geolocation source is configured.
        </p>
        <ul>
          {data.geography.map((geo) => (
            <li key={geo.request_origin_country}>
              {geo.request_origin_country}: {geo.events} events
            </li>
          ))}
        </ul>
      </Card>
      <Card>
        <h2>Security events</h2>
        <ul>
          {data.events.map((e) => (
            <li key={e.id}>
              {e.event_type} · {new Date(e.created_at).toLocaleString()}
            </li>
          ))}
        </ul>
      </Card>
      <Card>
        <h2>Moderation history</h2>
        <ul>
          {data.moderation.map((e) => (
            <li key={e.id}>
              {e.action} {e.target_type}: {e.reason}
            </li>
          ))}
        </ul>
      </Card>
    </main>
  );
}
