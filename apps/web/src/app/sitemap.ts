import type { MetadataRoute } from 'next';
import { fetchSkills } from '@/shared/api/server-client';
import { env } from '@/shared/config/env';

export const dynamic = 'force-dynamic';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const base = env.appUrl;
  const now = new Date();
  const staticRoutes: MetadataRoute.Sitemap = [
    { url: base, lastModified: now, changeFrequency: 'weekly', priority: 1 },
    { url: `${base}/skills`, lastModified: now, changeFrequency: 'daily', priority: 0.9 },
  ];

  try {
    const result = await fetchSkills();
    const skills = result?.data ?? [];
    const publicSkillRoutes: MetadataRoute.Sitemap = skills
      .filter((skill) => skill.status === 'published' && skill.visibility === 'public')
      .map((skill) => ({
        url: `${base}/skills/${skill.id}`,
        lastModified: new Date(skill.updatedAt),
        changeFrequency: 'weekly',
        priority: 0.8,
      }));

    return [...staticRoutes, ...publicSkillRoutes];
  } catch {
    return staticRoutes;
  }
}
