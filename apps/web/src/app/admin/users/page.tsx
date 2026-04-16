import Link from 'next/link';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { fetchAdminUsers, withValidSession } from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';

export default async function AdminUsersPage() {
  const { data: users } = await withValidSession((session) => fetchAdminUsers(session.accessToken));

  return (
    <div className="stack stack--lg">
      <Card>
        <table className="table">
          <thead>
            <tr>
              <th>Email</th>
              <th>Role</th>
              <th>Status</th>
              <th>Created</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => (
              <tr key={user.id}>
                <td>{user.email}</td>
                <td>{user.role}</td>
                <td>
                  <Badge tone={user.status === 'active' ? 'success' : 'warning'}>
                    {user.status}
                  </Badge>
                </td>
                <td>{formatDate(user.createdAt)}</td>
                <td>
                  <Link href={`/admin/users/${user.id}`} className="table-action-link">
                    Open
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </div>
  );
}
