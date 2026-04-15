import Link from 'next/link';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { fetchAdminUser, withValidSession } from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';
import { updateUserAction } from '@/features/admin/users/actions';

export default async function AdminUserPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const { data: detail } = await withValidSession((session) =>
    fetchAdminUser(session.accessToken, id),
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
            <Badge tone={user.status === 'active' ? 'success' : 'warning'}>{user.status}</Badge>
          </dd>
          <dt>Email verified</dt>
          <dd>{(user as any).emailVerifiedAt ? formatDate((user as any).emailVerifiedAt) : '—'}</dd>
          <dt>Joined</dt>
          <dd>{formatDate(user.createdAt)}</dd>
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
            </select>
          </label>
          <div>
            <Button type="submit">Save changes</Button>
          </div>
        </form>
      </Card>
    </div>
  );
}
