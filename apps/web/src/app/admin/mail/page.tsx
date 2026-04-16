import type {
  ApiEnvelope,
  MailCleanupRun,
  MailEvent,
  MailJob,
  MailOperationalSnapshot,
  MailRetentionSnapshot,
  MailSuppression,
} from '@/shared/api/types';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Input } from '@/shared/ui/input';
import {
  fetchAdminDeadLetters,
  fetchAdminMailEvents,
  fetchAdminMailCleanupRuns,
  fetchAdminMailOperations,
  fetchAdminMailRetention,
  fetchAdminMailSuppressions,
  withValidSession,
} from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';
import {
  cleanupMailRetentionAction,
  createMailSuppressionAction,
  deleteMailSuppressionAction,
  replayMailJobAction,
  requeueDeadLetterAction,
} from '@/features/admin/mail/actions';

const STATUS_TONE = {
  queued: 'neutral',
  retrying: 'warning',
  processing: 'warning',
  sent: 'success',
  dead_letter: 'warning',
} as const;

type AdminMailPageProps = {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
};

function firstQueryValue(value: string | string[] | undefined) {
  if (Array.isArray(value)) {
    return value[0] ?? '';
  }
  return value ?? '';
}

function unwrapData<T>(payload: T | ApiEnvelope<T>): T {
  if (typeof payload === 'object' && payload !== null && 'data' in payload) {
    return payload.data;
  }

  return payload;
}

export default async function AdminMailPage({ searchParams }: AdminMailPageProps) {
  const params = (await searchParams) ?? {};
  const deadLetterQuery = firstQueryValue(params.deadLetterQuery);
  const deadLetterMessageType = firstQueryValue(params.deadLetterMessageType);
  const deadLetterProvider = firstQueryValue(params.deadLetterProvider);
  const deadLetterErrorCode = firstQueryValue(params.deadLetterErrorCode);
  const eventQuery = firstQueryValue(params.eventQuery);
  const eventType = firstQueryValue(params.eventType);
  const eventMessageType = firstQueryValue(params.eventMessageType);
  const eventProvider = firstQueryValue(params.eventProvider);
  const eventErrorCode = firstQueryValue(params.eventErrorCode);

  const { mailOps, retention, cleanupRuns, deadLetters, events, suppressions } =
    await withValidSession(async (session) => {
      const [
        mailOpsPayload,
        retentionPayload,
        cleanupRunsPayload,
        deadLettersPayload,
        eventsPayload,
        suppressionsPayload,
      ] = await Promise.all([
        fetchAdminMailOperations(session.accessToken),
        fetchAdminMailRetention(session.accessToken),
        fetchAdminMailCleanupRuns(session.accessToken, 10),
        fetchAdminDeadLetters(session.accessToken, {
          query: deadLetterQuery,
          messageType: deadLetterMessageType,
          provider: deadLetterProvider,
          errorCode: deadLetterErrorCode,
          limit: 50,
        }),
        fetchAdminMailEvents(session.accessToken, {
          query: eventQuery,
          eventType,
          messageType: eventMessageType,
          provider: eventProvider,
          errorCode: eventErrorCode,
          limit: 50,
        }),
        fetchAdminMailSuppressions(session.accessToken),
      ]);

      return {
        mailOps: unwrapData<MailOperationalSnapshot>(mailOpsPayload),
        retention: unwrapData<MailRetentionSnapshot>(retentionPayload),
        cleanupRuns: unwrapData<MailCleanupRun[]>(cleanupRunsPayload),
        deadLetters: unwrapData<MailJob[]>(deadLettersPayload),
        events: unwrapData<MailEvent[]>(eventsPayload),
        suppressions: unwrapData<MailSuppression[]>(suppressionsPayload),
      };
    });

  return (
    <div className="stack stack--lg">
      <section className="grid">
        <Card>
          <h2>{mailOps.countsByStatus.queued}</h2>
          <p>Queued jobs waiting for an available worker and provider send slot.</p>
        </Card>
        <Card>
          <h2>{mailOps.countsByStatus.retrying}</h2>
          <p>Jobs scheduled for another attempt after a transient delivery failure.</p>
        </Card>
        <Card>
          <h2>{mailOps.countsByStatus.dead_letter}</h2>
          <p>Jobs that exhausted retries and now need operator review or manual replay.</p>
        </Card>
        <Card>
          <h2>{Math.round(mailOps.oldestReadyAgeSeconds)}s</h2>
          <p>
            Age of the oldest ready job in the queue, useful as a quick dispatch latency signal.
          </p>
        </Card>
      </section>

      <Card>
        <h2 className="card__title">Dead-letter reasons</h2>
        {Object.keys(mailOps.deadLettersByErrorCode).length === 0 ? (
          <p className="muted">No dead-letter error buckets right now.</p>
        ) : (
          <div className="inline-actions" style={{ rowGap: 10, flexWrap: 'wrap' }}>
            {Object.entries(mailOps.deadLettersByErrorCode).map(([errorCode, count]) => (
              <Badge key={errorCode} tone="warning">
                {errorCode}: {count}
              </Badge>
            ))}
          </div>
        )}
      </Card>

      <Card>
        <h2 className="card__title">Retention cleanup</h2>
        <p className="muted" style={{ marginBottom: 16 }}>
          Clean terminal mail jobs and historical mail events in bounded batches so the queue tables
          do not grow without limit.
        </p>
        <div
          style={{
            display: 'grid',
            gap: 12,
            gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            marginBottom: 16,
          }}
        >
          <div>
            <div style={{ fontSize: 12, color: 'var(--muted)', marginBottom: 4 }}>
              Jobs retention
            </div>
            <strong>{retention.jobsRetentionActive ? retention.jobsRetention : 'disabled'}</strong>
            <div className="muted" style={{ fontSize: 12 }}>
              Eligible: {retention.eligibleJobs}
            </div>
          </div>
          <div>
            <div style={{ fontSize: 12, color: 'var(--muted)', marginBottom: 4 }}>
              Events retention
            </div>
            <strong>
              {retention.eventsRetentionActive ? retention.eventsRetention : 'disabled'}
            </strong>
            <div className="muted" style={{ fontSize: 12 }}>
              Eligible: {retention.eligibleEvents}
            </div>
          </div>
          <div>
            <div style={{ fontSize: 12, color: 'var(--muted)', marginBottom: 4 }}>Batch size</div>
            <strong>{retention.cleanupBatchSize}</strong>
            <div className="muted" style={{ fontSize: 12 }}>
              One cleanup run removes up to this many jobs and events each.
            </div>
          </div>
          <div>
            <div style={{ fontSize: 12, color: 'var(--muted)', marginBottom: 4 }}>Auto cleanup</div>
            <strong>
              {retention.autoCleanupEnabled ? retention.cleanupInterval : 'disabled'}
              {retention.cleanupDryRun ? ' (dry run)' : ''}
            </strong>
            <div className="muted" style={{ fontSize: 12 }}>
              Periodic retention worker inside the Go API process.
            </div>
          </div>
          <div>
            <div style={{ fontSize: 12, color: 'var(--muted)', marginBottom: 4 }}>
              Webhook alerts
            </div>
            <strong>{retention.webhookAlertingEnabled ? 'enabled' : 'disabled'}</strong>
            <div className="muted" style={{ fontSize: 12 }}>
              Sends notifications on cleanup failures and backlog threshold breaches.
            </div>
          </div>
        </div>
        <div className="stack stack--sm" style={{ marginBottom: 16 }}>
          {retention.alerts.length > 0 ? (
            <div className="inline-actions" style={{ rowGap: 10, flexWrap: 'wrap' }}>
              {retention.alerts.map((alert) => (
                <Badge
                  key={alert.code}
                  tone={alert.severity === 'critical' ? 'warning' : 'neutral'}
                >
                  {alert.message}
                </Badge>
              ))}
            </div>
          ) : (
            <p className="muted" style={{ fontSize: 12 }}>
              No active retention alerts.
            </p>
          )}
          <p className="muted" style={{ fontSize: 12 }}>
            Jobs cutoff: {retention.jobsCutoff ? formatDate(retention.jobsCutoff) : 'disabled'}
          </p>
          <p className="muted" style={{ fontSize: 12 }}>
            Events cutoff:{' '}
            {retention.eventsCutoff ? formatDate(retention.eventsCutoff) : 'disabled'}
          </p>
        </div>
        <form action={cleanupMailRetentionAction}>
          <Button type="submit">Run cleanup batch</Button>
        </form>
      </Card>

      <Card>
        <h2 className="card__title">
          Cleanup history
          <span style={{ fontWeight: 400, textTransform: 'none', letterSpacing: 0 }}>
            {' '}
            ({cleanupRuns.length})
          </span>
        </h2>
        {cleanupRuns.length === 0 ? (
          <p className="muted">No cleanup history recorded yet.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>When</th>
                <th>Mode</th>
                <th>Status</th>
                <th>Candidates</th>
                <th>Deleted</th>
                <th>Duration</th>
                <th>Error</th>
              </tr>
            </thead>
            <tbody>
              {cleanupRuns.map((run) => (
                <tr key={run.id}>
                  <td>{formatDate(run.createdAt)}</td>
                  <td>
                    <Badge tone="neutral">
                      {run.mode}
                      {run.dryRun ? ' dry-run' : ''}
                    </Badge>
                  </td>
                  <td>
                    <Badge tone={run.status === 'failed' ? 'warning' : 'success'}>
                      {run.status}
                    </Badge>
                  </td>
                  <td>
                    jobs {run.candidateJobs}
                    <br />
                    events {run.candidateEvents}
                  </td>
                  <td>
                    jobs {run.deletedJobs}
                    <br />
                    events {run.deletedEvents}
                  </td>
                  <td>{run.durationMs} ms</td>
                  <td>{run.errorMessage ?? '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>

      <Card>
        <h2 className="card__title">Suppression list</h2>
        <form
          action={createMailSuppressionAction}
          className="stack stack--md"
          style={{ marginBottom: 16 }}
        >
          <div
            style={{
              display: 'grid',
              gap: 12,
              gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            }}
          >
            <label>
              <span>Kind</span>
              <select name="kind" defaultValue="email">
                <option value="email">email</option>
                <option value="domain">domain</option>
              </select>
            </label>
            <label>
              <span>Value</span>
              <Input name="value" required placeholder="user@example.com or example.com" />
            </label>
            <label>
              <span>Reason</span>
              <Input name="reason" required placeholder="hard bounce, legal block, test inbox" />
            </label>
          </div>
          <div className="inline-actions">
            <Button type="submit">Add suppression</Button>
          </div>
        </form>

        {suppressions.length === 0 ? (
          <p className="muted">No suppressions configured.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Kind</th>
                <th>Value</th>
                <th>Reason</th>
                <th>Created</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {suppressions.map((item) => (
                <tr key={item.id}>
                  <td>
                    <Badge tone="warning">{item.kind}</Badge>
                  </td>
                  <td>{item.value}</td>
                  <td>{item.reason}</td>
                  <td>{formatDate(item.createdAt)}</td>
                  <td>
                    <form action={deleteMailSuppressionAction.bind(null, item.id)}>
                      <Button type="submit">Remove</Button>
                    </form>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>

      <Card>
        <h2 className="card__title">
          Dead-letter jobs
          <span style={{ fontWeight: 400, textTransform: 'none', letterSpacing: 0 }}>
            {' '}
            ({deadLetters.length})
          </span>
        </h2>
        <form method="GET" className="stack stack--md" style={{ marginBottom: 16 }}>
          <div
            style={{
              display: 'grid',
              gap: 12,
              gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            }}
          >
            <label>
              <span>Search</span>
              <Input
                name="deadLetterQuery"
                defaultValue={deadLetterQuery}
                placeholder="recipient, job id, error"
              />
            </label>
            <label>
              <span>Message type</span>
              <Input
                name="deadLetterMessageType"
                defaultValue={deadLetterMessageType}
                placeholder="verification"
              />
            </label>
            <label>
              <span>Provider</span>
              <Input
                name="deadLetterProvider"
                defaultValue={deadLetterProvider}
                placeholder="smtp"
              />
            </label>
            <label>
              <span>Error code</span>
              <Input
                name="deadLetterErrorCode"
                defaultValue={deadLetterErrorCode}
                placeholder="smtp_auth_failed"
              />
            </label>
          </div>
          <div className="inline-actions">
            <Button type="submit">Apply dead-letter filters</Button>
          </div>
        </form>

        {deadLetters.length === 0 ? (
          <p className="muted">No dead-letter jobs. Delivery is currently flowing cleanly.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Recipient</th>
                <th>Type</th>
                <th>Status</th>
                <th>Error</th>
                <th>Attempts</th>
                <th>Updated</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {deadLetters.map((job) => {
                const requeueAction = requeueDeadLetterAction.bind(null, job.id);
                const replayAction = replayMailJobAction.bind(null, job.id);
                return (
                  <tr key={job.id}>
                    <td>
                      <span style={{ fontWeight: 500 }}>{job.recipientEmail}</span>
                      <br />
                      <span style={{ fontSize: 12, color: 'var(--muted)' }}>{job.id}</span>
                    </td>
                    <td>{job.messageType}</td>
                    <td>
                      <Badge tone={STATUS_TONE[job.status] ?? 'neutral'}>{job.status}</Badge>
                    </td>
                    <td>
                      <span style={{ fontSize: 13 }}>{job.lastErrorCode ?? 'unknown'}</span>
                      {job.lastError ? (
                        <>
                          <br />
                          <span style={{ fontSize: 12, color: 'var(--muted)' }}>
                            {job.lastError}
                          </span>
                        </>
                      ) : null}
                    </td>
                    <td>
                      {job.attempts}/{job.maxAttempts}
                    </td>
                    <td>{formatDate(job.updatedAt)}</td>
                    <td>
                      <div className="inline-actions">
                        <form action={requeueAction}>
                          <Button type="submit">Requeue</Button>
                        </form>
                        <form action={replayAction}>
                          <Button type="submit">Replay</Button>
                        </form>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </Card>

      <Card>
        <h2 className="card__title">
          Delivery events
          <span style={{ fontWeight: 400, textTransform: 'none', letterSpacing: 0 }}>
            {' '}
            ({events.length})
          </span>
        </h2>
        <form method="GET" className="stack stack--md" style={{ marginBottom: 16 }}>
          <div
            style={{
              display: 'grid',
              gap: 12,
              gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
            }}
          >
            <label>
              <span>Search</span>
              <Input
                name="eventQuery"
                defaultValue={eventQuery}
                placeholder="recipient, job id, provider message id"
              />
            </label>
            <label>
              <span>Event type</span>
              <Input name="eventType" defaultValue={eventType} placeholder="retry_scheduled" />
            </label>
            <label>
              <span>Message type</span>
              <Input
                name="eventMessageType"
                defaultValue={eventMessageType}
                placeholder="password_reset"
              />
            </label>
            <label>
              <span>Provider</span>
              <Input name="eventProvider" defaultValue={eventProvider} placeholder="gmail_api" />
            </label>
            <label>
              <span>Error code</span>
              <Input
                name="eventErrorCode"
                defaultValue={eventErrorCode}
                placeholder="gmail_api_429"
              />
            </label>
          </div>
          <div className="inline-actions">
            <Button type="submit">Apply event filters</Button>
          </div>
        </form>

        {events.length === 0 ? (
          <p className="muted">No mail events yet.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>When</th>
                <th>Event</th>
                <th>Type</th>
                <th>Recipient</th>
                <th>Provider</th>
                <th>Error</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {events.map((event) => (
                <tr key={event.id}>
                  <td>{formatDate(event.createdAt)}</td>
                  <td>
                    <span style={{ fontWeight: 500 }}>{event.eventType}</span>
                    {event.attempt != null ? (
                      <>
                        <br />
                        <span style={{ fontSize: 12, color: 'var(--muted)' }}>
                          attempt {event.attempt}
                        </span>
                      </>
                    ) : null}
                  </td>
                  <td>{event.messageType}</td>
                  <td>{event.recipientEmail}</td>
                  <td>{event.provider ?? 'queue'}</td>
                  <td>{event.errorCode ?? 'n/a'}</td>
                  <td>
                    <form action={replayMailJobAction.bind(null, event.jobId)}>
                      <Button type="submit">Replay</Button>
                    </form>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Card>
    </div>
  );
}
