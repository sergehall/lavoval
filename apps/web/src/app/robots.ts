import type { MetadataRoute } from 'next';
import { env } from '@/shared/config/env';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: '*',
      allow: '/',
      disallow: ['/admin', '/admin/', '/account', '/account/'],
    },
    sitemap: `${env.appUrl}/sitemap.xml`,
  };
}
