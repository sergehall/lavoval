import type { Metadata } from 'next';
import './globals.css';
import { AppShell } from '@/components/app-shell';
import { getSession } from '@/shared/api/server-client';

export const metadata: Metadata = {
  title: 'Lavoval',
  description:
    'Production-ready foundation for skills, modules, governance, and learning operations.',
};

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const session = await getSession();

  return (
    <html lang="en">
      <body>
        <AppShell user={session?.user}>{children}</AppShell>
      </body>
    </html>
  );
}
