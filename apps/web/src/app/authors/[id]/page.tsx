import type { Metadata } from 'next';
import Image from 'next/image';
import Link from 'next/link';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { ApiError, fetchPublicCreatorProfile } from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>;
}): Promise<Metadata> {
  const { id } = await params;

  try {
    const { data: profile } = await fetchPublicCreatorProfile(id);
    return {
      title: `${profile.fullName} | Lavoval`,
      description:
        profile.bio ?? `${profile.fullName} on Lavoval. Explore public skill offers and expertise.`,
      alternates: { canonical: `/authors/${id}` },
    };
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return { title: 'Author Not Found', robots: { index: false, follow: false } };
    }
    throw error;
  }
}

function initials(firstName: string, lastName: string) {
  return `${firstName.charAt(0)}${lastName.charAt(0)}`.toUpperCase();
}

export default async function PublicAuthorPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  let profile: Awaited<ReturnType<typeof fetchPublicCreatorProfile>>['data'] | null = null;
  let notFound = false;

  try {
    const response = await fetchPublicCreatorProfile(id);
    profile = response.data;
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound = true;
    } else {
      throw error;
    }
  }

  if (notFound || !profile) {
    return (
      <Card>
        <div className="empty-state stack stack--md">
          <Badge tone="warning">Not available</Badge>
          <h1>This public author profile is not available.</h1>
          <p className="muted">
            The creator may have disabled their public profile or this profile does not exist.
          </p>
          <Link href="/skills" className="site-nav__link site-nav__link--subtle">
            Back to Skills
          </Link>
        </div>
      </Card>
    );
  }

  const linkItems = [
    profile.websiteUrl ? { href: profile.websiteUrl, label: 'Website' } : null,
    profile.linkedinUrl ? { href: profile.linkedinUrl, label: 'LinkedIn' } : null,
    profile.githubUrl ? { href: profile.githubUrl, label: 'GitHub' } : null,
    profile.twitterUrl ? { href: profile.twitterUrl, label: 'Twitter / X' } : null,
  ].filter(Boolean) as Array<{ href: string; label: string }>;

  return (
    <div className="stack stack--lg">
      <Card>
        <div className="stack stack--md">
          <div
            className="inline-actions"
            style={{ justifyContent: 'space-between', alignItems: 'flex-start', gap: '20px' }}
          >
            <div className="inline-actions" style={{ alignItems: 'center', gap: '16px' }}>
              {profile.avatarUrl ? (
                <Image
                  src={profile.avatarUrl}
                  alt={profile.fullName}
                  width={96}
                  height={96}
                  unoptimized
                  style={{
                    width: '96px',
                    height: '96px',
                    borderRadius: '24px',
                    objectFit: 'cover',
                    border: '1px solid rgba(255,255,255,0.7)',
                  }}
                />
              ) : (
                <div
                  style={{
                    width: '96px',
                    height: '96px',
                    borderRadius: '24px',
                    display: 'grid',
                    placeItems: 'center',
                    fontSize: '1.5rem',
                    fontWeight: 700,
                    border: '1px solid rgba(255,255,255,0.7)',
                    background:
                      'linear-gradient(180deg, rgba(255,255,255,0.92), rgba(245,238,230,0.92))',
                  }}
                >
                  {initials(profile.firstName, profile.lastName)}
                </div>
              )}

              <div className="stack stack--sm" style={{ minWidth: 0, paddingTop: '10px' }}>
                <div className="inline-actions" style={{ gap: '10px', alignItems: 'center' }}>
                  <h1 style={{ margin: 0 }}>{profile.fullName}</h1>
                  {profile.availabilityStatus ? (
                    <Badge tone="success">{profile.availabilityStatus}</Badge>
                  ) : null}
                </div>
                {profile.username ? (
                  <p className="muted" style={{ marginTop: '2px' }}>
                    @{profile.username}
                  </p>
                ) : null}
              </div>
            </div>

            <Link href="/skills" className="site-nav__link site-nav__link--subtle">
              Back to Skills
            </Link>
          </div>

          {profile.bio ? <p style={{ margin: 0, maxWidth: '760px' }}>{profile.bio}</p> : null}

          <div className="inline-actions muted" style={{ flexWrap: 'wrap', gap: '10px 14px' }}>
            {profile.location ? <span>{profile.location}</span> : null}
            {profile.languages && profile.languages.length > 0 ? (
              <span>Languages: {profile.languages.join(', ')}</span>
            ) : null}
            <span>{profile.publicSkills.length} public skill offers</span>
          </div>

          {linkItems.length > 0 ? (
            <div className="inline-actions" style={{ flexWrap: 'wrap', gap: '10px' }}>
              {linkItems.map((item) => (
                <a
                  key={item.label}
                  href={item.href}
                  target="_blank"
                  rel="noreferrer noopener"
                  className="site-nav__link site-nav__link--subtle"
                >
                  {item.label}
                </a>
              ))}
            </div>
          ) : null}
        </div>
      </Card>

      <Card>
        <div className="stack stack--md">
          <div className="section-heading">
            <h2>Public skill offers</h2>
            <span className="muted">{profile.publicSkills.length} visible</span>
          </div>
          {profile.publicSkills.length > 0 ? (
            <div className="data-list">
              {profile.publicSkills.map((skill) => (
                <article key={skill.id} className="skill-catalog-card">
                  <div className="stack stack--sm">
                    <div className="inline-actions" style={{ justifyContent: 'space-between' }}>
                      <h3 style={{ margin: 0 }}>{skill.title}</h3>
                      <Badge tone="neutral">{skill.skillType}</Badge>
                    </div>
                    <p>{skill.summary}</p>
                  </div>
                  <div className="skill-catalog-card__stats muted">
                    <span>{skill.difficulty}</span>
                    {skill.avgRating > 0 ? <span>{skill.avgRating.toFixed(1)} rating</span> : null}
                    {skill.runsCount > 0 ? <span>{skill.runsCount} runs</span> : null}
                    <span>Updated {formatDate(skill.updatedAt)}</span>
                  </div>
                  <div className="skill-catalog-card__footer">
                    <span />
                    <Link href={`/skills/${skill.id}`} className="skill-catalog-card__link">
                      View skill
                    </Link>
                  </div>
                </article>
              ))}
            </div>
          ) : (
            <p className="muted">This author has not published any public skill offers yet.</p>
          )}
        </div>
      </Card>
    </div>
  );
}
