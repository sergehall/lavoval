import type { Profile } from '@lavoval/contracts';
import { updateProfileAction } from './actions';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Textarea } from '@/shared/ui/textarea';

export function ProfileForm({ profile }: { profile: Profile }) {
  return (
    <form action={updateProfileAction} className="stack stack--md form-grid">
      <label>
        <span>First name</span>
        <Input name="firstName" defaultValue={profile.firstName} required />
      </label>
      <label>
        <span>Last name</span>
        <Input name="lastName" defaultValue={profile.lastName} required />
      </label>
      <label className="form-grid__full">
        <span>Timezone</span>
        <Input name="timezone" defaultValue={profile.timezone} required />
      </label>
      <label className="form-grid__full">
        <span>Bio</span>
        <Textarea name="bio" defaultValue={profile.bio ?? ''} rows={6} />
      </label>
      <div className="form-grid__full">
        <Button type="submit">Update profile</Button>
      </div>
    </form>
  );
}
