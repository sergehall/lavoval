import Link from 'next/link';
import { Card } from '@/shared/ui/card';
import { ProfileForm } from '@/features/profile/profile-form';
import { fetchProfile, withValidSession } from '@/shared/api/server-client';

export default async function ProfilePage() {
  const { data: profile } = await withValidSession((session) => fetchProfile(session.accessToken));

  return (
    <div className="stack stack--md">
      <div className="stack stack--sm">
        <h1>Your marketplace identity</h1>
        <p className="muted">
          This profile gives context to your skill offers and helps other people understand the
          experience behind what you publish.
        </p>
        {profile.isPublicProfile ? (
          <div className="inline-actions">
            <Link
              href={`/authors/${profile.userId}`}
              className="site-nav__link site-nav__link--subtle"
            >
              Preview public profile
            </Link>
          </div>
        ) : null}
      </div>
      <Card>
        <ProfileForm profile={profile} />
      </Card>
    </div>
  );
}
