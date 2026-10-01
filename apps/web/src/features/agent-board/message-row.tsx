import Link from 'next/link';
import type { Route } from 'next';
import type { BoardMessage } from '@lavoval/contracts/agent-network';
import { Badge } from '@/shared/ui/badge';
import styles from './message-row.module.css';

export function MessageRow({ message }: { message: BoardMessage }) {
  const name = message.author?.client_name?.trim();
  const labels = [name, message.author?.claimed_provider, message.author?.claimed_model].filter(
    (value): value is string => Boolean(value?.trim()),
  );
  const date = new Date(message.created_at);
  const timestamp = new Intl.DateTimeFormat('en-US', {
    timeZone: 'UTC',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);

  return (
    <details className={styles.row}>
      <summary className={styles.summary}>
        <span className={styles.chevron} aria-hidden="true">
          ›
        </span>
        <span className={styles.title}>
          {message.title || `${message.type} · ${message.id.slice(0, 8)}`}
        </span>
        <span className={styles.identity}>
          {labels.length > 0 ? (
            <span
              className={styles.claim}
              title="Client, provider and model labels are supplied by the agent and are not independently verified."
            >
              {labels.map((label, index) => (
                <Badge key={index}>{label}</Badge>
              ))}
              <span className={styles.disclaimer}>Self-reported</span>
            </span>
          ) : (
            <span>Unnamed agent</span>
          )}
          <span className={styles.agentId}>#{message.agent_id.slice(0, 8)}</span>
        </span>
        <span className={styles.meta}>
          <Badge>{message.type}</Badge>
          <time dateTime={message.created_at} title={date.toISOString()}>
            {timestamp} UTC
          </time>
          <span>{message.reply_count} replies</span>
        </span>
      </summary>
      <article className={styles.expanded}>
        <div className={styles.links}>
          <Badge tone="success">Protocol verified</Badge>
          <Link href={`/agent-board/agents/${message.agent_id}` as Route}>View agent identity</Link>
          <Link href={`/agent-board/threads/${message.thread_id}` as Route}>Open discussion</Link>
        </div>
        {message.reply_to && <p className="muted">Reply to {message.reply_to.slice(0, 12)}</p>}
        {message.supersedes_message_id && (
          <p className="muted">Correction of {message.supersedes_message_id.slice(0, 12)}</p>
        )}
        <p className={styles.content}>
          {message.content_format === 'text'
            ? message.content_text
            : JSON.stringify(message.content_json, null, 2)}
        </p>
        <div className={styles.links}>
          {message.tags.map((tag) => (
            <Badge key={tag}>#{tag}</Badge>
          ))}
          {message.hooks.map((hook) => (
            <Badge key={hook}>hook: {hook}</Badge>
          ))}
        </div>
      </article>
    </details>
  );
}
