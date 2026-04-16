import type { Metadata } from 'next';
import './globals.css';
import { AppShell } from '@/components/app-shell';
import { getValidatedSession } from '@/shared/api/server-client';
import { env } from '@/shared/config/env';
import { loadPublicSkills } from '@/shared/lib/public-skill-loader';

const siteDescription =
  'Lavoval is a skill-exchange marketplace for the AI era where people publish expertise, discover trusted specialists, and turn human know-how into reusable skill offers, practical modules, and run-ready workflows.';

export const metadata: Metadata = {
  metadataBase: new URL(env.appUrl),
  applicationName: 'Lavoval',
  manifest: '/manifest.webmanifest',

  title: {
    default: 'Lavoval | Skill Exchange Marketplace For The AI Era',
    template: '%s | Lavoval',
  },

  description: siteDescription,
  keywords: [
    'skill exchange marketplace',
    'AI era skills',
    'human expertise platform',
    'knowledge marketplace',
    'peer to peer learning',
    'skill offers',
    'expertise discovery',
    'lavoval',
  ],
  category: 'education',
  creator: 'Lavoval',
  publisher: 'Lavoval',
  alternates: {
    canonical: '/',
  },
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },

  openGraph: {
    title: 'Lavoval',
    description: siteDescription,
    url: env.appUrl,
    siteName: 'Lavoval',
    images: [{ url: '/og-image.png', width: 1200, height: 630 }],
    locale: 'en_US',
    type: 'website',
  },

  twitter: {
    card: 'summary_large_image',
    title: 'Lavoval',
    description: siteDescription,
    images: ['/og-image.png'],
  },

  icons: {
    icon: [
      { url: '/favicon.ico' },
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
    ],
    shortcut: ['/favicon.ico'],
    apple: '/apple-touch-icon.png',
  },

  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-image-preview': 'large',
      'max-snippet': -1,
      'max-video-preview': -1,
    },
  },
};

export default async function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  const session = await getValidatedSession();
  const publicSkills = await loadPublicSkills();

  return (
    <html lang="en">
      <body>
        <AppShell user={session?.user} searchSkills={publicSkills}>
          {children}
        </AppShell>
      </body>
    </html>
  );
}
