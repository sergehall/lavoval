import type { PropsWithChildren } from 'react';
import { requireAdminSession } from '@/shared/api/server-client';
import { AdminNav } from '@/components/admin-nav';

export default async function AdminLayout({ children }: PropsWithChildren) {
  await requireAdminSession();
  return (
    <div className="admin-layout">
      <AdminNav />
      <div className="admin-layout__content">{children}</div>
    </div>
  );
}
