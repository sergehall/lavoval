import type { Metadata } from 'next';
import './globals.css';
import { AppShell } from '@/components/app-shell';
import { getSession } from '@/shared/api/server-client';

export const metadata: Metadata = {
  metadataBase: new URL('https://lavoval.com'),

  title: {
    default: 'Lavoval',
    template: '%s | Lavoval',
  },

  description:
    'A skill-exchange marketplace for the AI era where people publish expertise, discover each other, and turn human know-how into reusable modules.',

  openGraph: {
    title: 'Lavoval',
    description:
      'A skill-exchange marketplace for the AI era where people publish expertise, discover each other, and turn human know-how into reusable modules.',
    url: 'https://lavoval.com',
    siteName: 'Lavoval',
    images: [{ url: '/og-image.png', width: 1200, height: 630 }],
    locale: 'en_US',
    type: 'website',
  },

  twitter: {
    card: 'summary_large_image',
    title: 'Lavoval',
    description:
      'A skill-exchange marketplace for the AI era where people publish expertise, discover each other, and turn human know-how into reusable modules.',
    images: ['/og-image.png'],
  },

  icons: {
    icon: [
      { url: '/favicon.ico' },
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
    ],
    apple: '/apple-touch-icon.png',
    other: [
      { rel: 'android-chrome-192x192', url: '/android-chrome-192x192.png' },
      { rel: 'android-chrome-512x512', url: '/android-chrome-512x512.png' },
    ],
  },

  robots: {
    index: true,
    follow: true,
  },
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
