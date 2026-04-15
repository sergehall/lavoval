import type { Profile } from '@lavoval/contracts';
import { updateProfileAction } from './actions';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Textarea } from '@/shared/ui/textarea';

function arrayToInput(arr: string[] | null | undefined): string {
  return arr?.join(', ') ?? '';
}

export function ProfileForm({ profile }: { profile: Profile }) {
  return (
    <form action={updateProfileAction} className="stack stack--md form-grid">
      {/* ── Core identity ─────────────────────────────────────── */}
      <label>
        <span>First name</span>
        <Input
          name="firstName"
          defaultValue={profile.firstName}
          required
          minLength={2}
          maxLength={100}
        />
      </label>
      <label>
        <span>Last name</span>
        <Input
          name="lastName"
          defaultValue={profile.lastName}
          required
          minLength={2}
          maxLength={100}
        />
      </label>
      <label className="form-grid__full">
        <span>Bio</span>
        <Textarea
          name="bio"
          defaultValue={profile.bio ?? ''}
          rows={6}
          maxLength={500}
          placeholder="Share what you know, how you learned it, and why someone would want to exchange knowledge with you."
        />
      </label>
      <input type="hidden" name="timezone" value={profile.timezone} />

      {/* ── Public identity ───────────────────────────────────── */}
      <label>
        <span>Username</span>
        <Input
          name="username"
          defaultValue={profile.username ?? ''}
          minLength={3}
          maxLength={30}
          pattern="^[a-zA-Z0-9_-]+$"
          title="Letters, digits, dash and underscore only"
          placeholder="your-handle"
        />
      </label>
      <label>
        <span>Location</span>
        <Input
          name="location"
          defaultValue={profile.location ?? ''}
          minLength={2}
          maxLength={100}
          placeholder="City, Country"
        />
      </label>
      <label className="form-grid__full">
        <span>Avatar URL</span>
        <Input
          name="avatarUrl"
          type="url"
          defaultValue={profile.avatarUrl ?? ''}
          maxLength={2048}
          placeholder="https://example.com/avatar.png"
        />
      </label>

      {/* ── Marketplace tags ──────────────────────────────────── */}
      <label className="form-grid__full">
        <span>
          Skills <small>(comma-separated, max 20)</small>
        </span>
        <Input
          name="skills"
          defaultValue={arrayToInput(profile.skills)}
          maxLength={1200}
          placeholder="TypeScript, Go, System design"
        />
      </label>
      <label className="form-grid__full">
        <span>
          Languages <small>(comma-separated codes, e.g. en, ru)</small>
        </span>
        <Input
          name="languages"
          defaultValue={arrayToInput(profile.languages)}
          maxLength={120}
          placeholder="en, ru"
        />
      </label>

      {/* ── Social links ──────────────────────────────────────── */}
      <label>
        <span>Website</span>
        <Input
          name="websiteUrl"
          type="url"
          defaultValue={profile.websiteUrl ?? ''}
          maxLength={2048}
          placeholder="https://yoursite.com"
        />
      </label>
      <label>
        <span>LinkedIn</span>
        <Input
          name="linkedinUrl"
          type="url"
          defaultValue={profile.linkedinUrl ?? ''}
          maxLength={2048}
          placeholder="https://linkedin.com/in/yourname"
        />
      </label>
      <label>
        <span>GitHub</span>
        <Input
          name="githubUrl"
          type="url"
          defaultValue={profile.githubUrl ?? ''}
          maxLength={2048}
          placeholder="https://github.com/yourname"
        />
      </label>
      <label>
        <span>Twitter / X</span>
        <Input
          name="twitterUrl"
          type="url"
          defaultValue={profile.twitterUrl ?? ''}
          maxLength={2048}
          placeholder="https://x.com/yourname"
        />
      </label>

      {/* ── Marketplace mechanics ─────────────────────────────── */}
      <label>
        <span>Availability</span>
        <select name="availabilityStatus" defaultValue={profile.availabilityStatus}>
          <option value="open">Open to exchange</option>
          <option value="limited">Limited availability</option>
          <option value="closed">Not available</option>
        </select>
      </label>
      <label>
        <span>
          <input
            type="checkbox"
            name="isPublicProfile"
            value="true"
            defaultChecked={profile.isPublicProfile}
          />{' '}
          Show profile in public catalogue
        </span>
      </label>

      <div className="form-grid__full">
        <Button type="submit">Save identity</Button>
      </div>
    </form>
  );
}
