import type { Metadata } from 'next';
import './globals.css';
import { AppShell } from '@/components/app-shell';
import { getSession } from '@/shared/api/server-client';

export const metadata: Metadata = {
  title: 'Lavoval',
  description:
    'A skill-exchange marketplace for the AI era where people publish expertise, discover each other, and turn human know-how into reusable modules.',
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
