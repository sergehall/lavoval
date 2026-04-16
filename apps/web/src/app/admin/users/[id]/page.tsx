import Link from 'next/link';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { fetchAdminUser, fetchUserAuditLog, withValidSession } from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';
import { updateUserAction } from '@/features/admin/users/actions';

function getStatusTone(status: string) {
  switch (status) {
    case 'active':
      return 'success';
    case 'blocked':
      return 'warning';
    case 'suspended':
      return 'warning';
    default:
      return 'neutral';
  }
}

export default async function AdminUserPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [{ data: detail }, auditLog] = await withValidSession((session) =>
    Promise.all([
      fetchAdminUser(session.accessToken, id),
      fetchUserAuditLog(session.accessToken, id).catch(() => []),
    ]),
  );
  const { user, profile } = detail;

  const boundUpdateUser = updateUserAction.bind(null, id);

  return (
    <div className="stack stack--lg">
      {/* ── Back link ───────────────────────────────── */}
      <Link href="/admin/users" className="muted" style={{ fontSize: 13 }}>
        ← Back to users
      </Link>

      <h1>
        {profile.firstName} {profile.lastName}
      </h1>

      {/* ── Account ─────────────────────────────────── */}
      <Card>
        <h2 className="card__title">Account</h2>
        <dl className="detail-grid">
          <dt>Email</dt>
          <dd>{user.email}</dd>
          <dt>Role</dt>
          <dd>
            <Badge tone={user.role === 'admin' ? 'warning' : 'neutral'}>{user.role}</Badge>
          </dd>
          <dt>Status</dt>
          <dd>
            <Badge tone={getStatusTone(user.status)}>{user.status}</Badge>
          </dd>
          <dt>Email verified</dt>
          <dd>{user.emailVerifiedAt ? formatDate(user.emailVerifiedAt) : '—'}</dd>
          <dt>Joined</dt>
          <dd>{formatDate(user.createdAt)}</dd>
          <dt>Suspended</dt>
          <dd>{user.suspendedAt ? formatDate(user.suspendedAt) : '—'}</dd>
          <dt>Blocked</dt>
          <dd>{user.blockedAt ? formatDate(user.blockedAt) : '—'}</dd>
        </dl>
      </Card>

      {/* ── Profile ─────────────────────────────────── */}
      <Card>
        <h2 className="card__title">Identity</h2>
        <dl className="detail-grid">
          <dt>Username</dt>
          <dd>{profile.username ?? '—'}</dd>
          <dt>Location</dt>
          <dd>{profile.location ?? '—'}</dd>
          <dt>Availability</dt>
          <dd>{profile.availabilityStatus}</dd>
          <dt>Public profile</dt>
          <dd>{profile.isPublicProfile ? 'Yes' : 'No'}</dd>
          {profile.bio && (
            <>
              <dt>Bio</dt>
              <dd>{profile.bio}</dd>
            </>
          )}
          {profile.skills && profile.skills.length > 0 && (
            <>
              <dt>Skills</dt>
              <dd>{profile.skills.join(', ')}</dd>
            </>
          )}
          {profile.languages && profile.languages.length > 0 && (
            <>
              <dt>Languages</dt>
              <dd>{profile.languages.join(', ')}</dd>
            </>
          )}
          {profile.websiteUrl && (
            <>
              <dt>Website</dt>
              <dd>
                <a href={profile.websiteUrl} target="_blank" rel="noreferrer">
                  {profile.websiteUrl}
                </a>
              </dd>
            </>
          )}
          {profile.linkedinUrl && (
            <>
              <dt>LinkedIn</dt>
              <dd>
                <a href={profile.linkedinUrl} target="_blank" rel="noreferrer">
                  {profile.linkedinUrl}
                </a>
              </dd>
            </>
          )}
          {profile.githubUrl && (
            <>
              <dt>GitHub</dt>
              <dd>
                <a href={profile.githubUrl} target="_blank" rel="noreferrer">
                  {profile.githubUrl}
                </a>
              </dd>
            </>
          )}
          {profile.twitterUrl && (
            <>
              <dt>Twitter / X</dt>
              <dd>
                <a href={profile.twitterUrl} target="_blank" rel="noreferrer">
                  {profile.twitterUrl}
                </a>
              </dd>
            </>
          )}
        </dl>
      </Card>

      {/* ── Governance ──────────────────────────────── */}
      <Card>
        <h2 className="card__title">Governance</h2>
        <form action={boundUpdateUser} className="stack stack--md">
          <label>
            <span>Role</span>
            <select name="role" defaultValue={user.role}>
              <option value="user">user</option>
              <option value="admin">admin</option>
            </select>
          </label>
          <label>
            <span>Status</span>
            <select name="status" defaultValue={user.status}>
              <option value="active">active</option>
              <option value="invited">invited</option>
              <option value="suspended">suspended</option>
              <option value="blocked">blocked</option>
            </select>
          </label>
          <label>
            <span>Reason (required for suspend / block)</span>
            <textarea
              name="reason"
              rows={3}
              maxLength={500}
              placeholder="Explain the reason for this status change…"
              defaultValue={user.suspensionReason ?? user.blockReason ?? ''}
            />
          </label>
          <div>
            <Button type="submit">Save changes</Button>
          </div>
        </form>
      </Card>

      {/* ── Audit log ───────────────────────────────── */}
      {auditLog && auditLog.length > 0 && (
        <Card>
          <h2 className="card__title">Audit log</h2>
          <table style={{ width: '100%', fontSize: 13, borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ textAlign: 'left', borderBottom: '1px solid var(--border)' }}>
                <th style={{ padding: '4px 8px' }}>When</th>
                <th style={{ padding: '4px 8px' }}>Action</th>
                <th style={{ padding: '4px 8px' }}>Reason</th>
              </tr>
            </thead>
            <tbody>
              {auditLog.map((entry) => (
                <tr key={entry.id} style={{ borderBottom: '1px solid var(--border-subtle)' }}>
                  <td style={{ padding: '4px 8px', whiteSpace: 'nowrap' }}>
                    {formatDate(entry.createdAt)}
                  </td>
                  <td style={{ padding: '4px 8px' }}>{entry.action}</td>
                  <td style={{ padding: '4px 8px', color: 'var(--muted)' }}>
                    {entry.reason ?? '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
}
