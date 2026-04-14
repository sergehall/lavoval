import { Card } from '@/shared/ui/card';
import { ProfileForm } from '@/features/profile/profile-form';
import { fetchProfile, withValidSession } from '@/shared/api/server-client';

export default async function ProfilePage() {
  const { data: profile } = await withValidSession((session) => fetchProfile(session.accessToken));

  return (
    <div className="stack stack--md">
      <Card>
        <ProfileForm profile={profile} />
      </Card>
    </div>
  );
}
