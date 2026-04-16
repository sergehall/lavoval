import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Input } from '@/shared/ui/input';
import { fetchAdminEnrollments, withValidSession } from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';
import { assignSkillAction, updateEnrollmentAction } from '@/features/admin/enrollments/actions';

const STATUS_TONE = {
  assigned: 'neutral',
  in_progress: 'warning',
  completed: 'success',
} as const;

export default async function AdminEnrollmentsPage() {
  const { data: enrollments } = await withValidSession((session) =>
    fetchAdminEnrollments(session.accessToken),
  );

  return (
    <div className="stack stack--lg">
      {/* ── Assign form ──────────────────────────────────── */}
      <Card>
        <h2 className="card__title">Assign skill to user</h2>
        <form action={assignSkillAction} className="stack stack--md form-grid">
          <label>
            <span>User ID</span>
            <Input
              name="userId"
              required
              placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
              pattern="^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
              title="Must be a valid UUID"
            />
          </label>
          <label>
            <span>Skill ID</span>
            <Input
              name="skillId"
              required
              placeholder="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
              pattern="^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
              title="Must be a valid UUID"
            />
          </label>
          <div className="form-grid__full">
            <Button type="submit">Assign</Button>
          </div>
        </form>
      </Card>

      {/* ── Enrollments list ─────────────────────────────── */}
      <Card>
        <h2 className="card__title">
          All enrollments{' '}
          <span style={{ fontWeight: 400, textTransform: 'none', letterSpacing: 0 }}>
            ({enrollments.length})
          </span>
        </h2>

        {enrollments.length === 0 ? (
          <p className="muted">No enrollments yet.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>User</th>
                <th>Skill</th>
                <th>Status</th>
                <th>Progress</th>
                <th>Assigned</th>
                <th>Update</th>
              </tr>
            </thead>
            <tbody>
              {enrollments.map((e) => {
                const boundUpdate = updateEnrollmentAction.bind(null, e.id);
                return (
                  <tr key={e.id}>
                    <td style={{ fontSize: 13 }}>{e.userEmail}</td>
                    <td>
                      <span style={{ fontWeight: 500 }}>{e.skillTitle}</span>
                      <br />
                      <span style={{ fontSize: 12, color: 'var(--muted)' }}>{e.skillSlug}</span>
                    </td>
                    <td>
                      <Badge tone={STATUS_TONE[e.status as keyof typeof STATUS_TONE] ?? 'neutral'}>
                        {e.status.replace('_', ' ')}
                      </Badge>
                    </td>
                    <td>{e.progressPercent}%</td>
                    <td style={{ fontSize: 12 }}>{formatDate(e.assignedAt)}</td>
                    <td>
                      <form action={boundUpdate} className="enrollment-update-form">
                        <select name="status" defaultValue={e.status} className="enrollment-select">
                          <option value="assigned">assigned</option>
                          <option value="in_progress">in progress</option>
                          <option value="completed">completed</option>
                        </select>
                        <input
                          type="number"
                          name="progressPercent"
                          defaultValue={e.progressPercent}
                          min={0}
                          max={100}
                          className="enrollment-progress"
                          aria-label="Progress percent"
                        />
                        <button type="submit" className="enrollment-save-btn">
                          Save
                        </button>
                      </form>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </Card>
    </div>
  );
}
