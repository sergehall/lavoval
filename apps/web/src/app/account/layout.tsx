import type { PropsWithChildren } from 'react';
import { requireSession } from '@/shared/api/server-client';

export default async function AccountLayout({ children }: PropsWithChildren) {
  await requireSession();
  return <>{children}</>;
}
