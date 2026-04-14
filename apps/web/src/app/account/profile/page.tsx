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
      </div>
      <Card>
        <ProfileForm profile={profile} />
      </Card>
    </div>
  );
}
