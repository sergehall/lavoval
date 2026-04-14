import type { PropsWithChildren } from 'react';
import { requireAdminSession } from '@/shared/api/server-client';

export default async function AdminLayout({ children }: PropsWithChildren) {
  await requireAdminSession();
  return <>{children}</>;
}
