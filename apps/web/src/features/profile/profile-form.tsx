'use client';

import { useActionState, useEffect, useRef, useState } from 'react';
import { useFormStatus } from 'react-dom';
import type { Profile } from '@lavoval/contracts';
import { updateProfileAction, type ProfileFormState } from './actions';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { Textarea } from '@/shared/ui/textarea';

const initialProfileFormState: ProfileFormState = {
  error: null,
  success: null,
  savedAt: null,
};

function arrayToInput(arr: string[] | null | undefined): string {
  return arr?.join(', ') ?? '';
}

function serializeForm(form: HTMLFormElement): string {
  const entries = Array.from(new FormData(form).entries()).map(([key, value]) => [
    key,
    String(value),
  ]);
  entries.sort(([leftKey, leftValue], [rightKey, rightValue]) => {
    if (leftKey === rightKey) return leftValue.localeCompare(rightValue);
    return leftKey.localeCompare(rightKey);
  });
  return JSON.stringify(entries);
}

function formatSavedAt(savedAt: string): string {
  const date = new Date(savedAt);
  return new Intl.DateTimeFormat('en-US', {
    hour: 'numeric',
    minute: '2-digit',
  }).format(date);
}

export function ProfileForm({ profile }: { profile: Profile }) {
  const [state, action] = useActionState(updateProfileAction, initialProfileFormState);
  const formRef = useRef<HTMLFormElement>(null);
  const initialSnapshotRef = useRef('');
  const successHideTimeoutRef = useRef<number | null>(null);
  const [isDirty, setIsDirty] = useState(false);
  const [showSuccessMessage, setShowSuccessMessage] = useState(false);

  useEffect(() => {
    if (!formRef.current) return;
    initialSnapshotRef.current = serializeForm(formRef.current);
    return () => {
      if (successHideTimeoutRef.current) {
        window.clearTimeout(successHideTimeoutRef.current);
      }
    };
  }, []);

  useEffect(() => {
    if (!state.success || !state.savedAt || !formRef.current) return;

    initialSnapshotRef.current = serializeForm(formRef.current);
    const showSuccessFrame = window.requestAnimationFrame(() => {
      setIsDirty(false);
      setShowSuccessMessage(true);
    });

    if (successHideTimeoutRef.current) {
      window.clearTimeout(successHideTimeoutRef.current);
    }

    successHideTimeoutRef.current = window.setTimeout(() => {
      setShowSuccessMessage(false);
    }, 4000);

    return () => {
      window.cancelAnimationFrame(showSuccessFrame);
    };
  }, [state.success, state.savedAt]);

  useEffect(() => {
    if (!formRef.current) return;
    initialSnapshotRef.current = serializeForm(formRef.current);
  }, [profile]);

  useEffect(() => {
    const warning = 'You have unsaved changes. Leave this page without saving?';

    const handleBeforeUnload = (event: BeforeUnloadEvent) => {
      if (!isDirty) return;
      event.preventDefault();
      event.returnValue = warning;
    };

    const handleDocumentClick = (event: MouseEvent) => {
      if (!isDirty) return;
      if (event.defaultPrevented || event.button !== 0) return;
      if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;

      const target = event.target;
      if (!(target instanceof Element)) return;

      const anchor = target.closest('a[href]');
      if (!(anchor instanceof HTMLAnchorElement)) return;

      const href = anchor.getAttribute('href');
      if (!href || href.startsWith('#') || href.startsWith('mailto:') || href.startsWith('tel:')) {
        return;
      }

      const destination = new URL(anchor.href, window.location.href);
      const current = new URL(window.location.href);

      if (destination.href === current.href) return;

      const confirmed = window.confirm(warning);
      if (!confirmed) {
        event.preventDefault();
        event.stopPropagation();
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);
    document.addEventListener('click', handleDocumentClick, true);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
      document.removeEventListener('click', handleDocumentClick, true);
    };
  }, [isDirty]);

  const syncDirtyState = () => {
    if (!formRef.current) return;
    setIsDirty(serializeForm(formRef.current) !== initialSnapshotRef.current);
    if (showSuccessMessage) setShowSuccessMessage(false);
  };

  return (
    <form
      ref={formRef}
      action={action}
      className="stack stack--md form-grid"
      onInput={syncDirtyState}
      onChange={syncDirtyState}
    >
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
          placeholder="https://avatars.githubusercontent.com/u/123456"
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

      <div className="form-grid__full stack stack--sm">
        <div>
          <strong>Public profile visibility</strong>
          <p className="muted" style={{ margin: '4px 0 0' }}>
            Choose which fields other people can see on your public author page.
          </p>
        </div>
        <div className="grid grid--compact">
          <label>
            <span>
              <input
                type="checkbox"
                name="showAvatar"
                value="true"
                defaultChecked={profile.showAvatar}
              />{' '}
              Show avatar
            </span>
          </label>
          <label>
            <span>
              <input type="checkbox" name="showBio" value="true" defaultChecked={profile.showBio} />{' '}
              Show bio
            </span>
          </label>
          <label>
            <span>
              <input
                type="checkbox"
                name="showLocation"
                value="true"
                defaultChecked={profile.showLocation}
              />{' '}
              Show location
            </span>
          </label>
          <label>
            <span>
              <input
                type="checkbox"
                name="showSkills"
                value="true"
                defaultChecked={profile.showSkills}
              />{' '}
              Show skills
            </span>
          </label>
          <label>
            <span>
              <input
                type="checkbox"
                name="showLanguages"
                value="true"
                defaultChecked={profile.showLanguages}
              />{' '}
              Show languages
            </span>
          </label>
          <label>
            <span>
              <input
                type="checkbox"
                name="showAvailabilityStatus"
                value="true"
                defaultChecked={profile.showAvailabilityStatus}
              />{' '}
              Show availability
            </span>
          </label>
          <label>
            <span>
              <input
                type="checkbox"
                name="showWebsiteUrl"
                value="true"
                defaultChecked={profile.showWebsiteUrl}
              />{' '}
              Show website
            </span>
          </label>
          <label>
            <span>
              <input
                type="checkbox"
                name="showLinkedinUrl"
                value="true"
                defaultChecked={profile.showLinkedinUrl}
              />{' '}
              Show LinkedIn
            </span>
          </label>
          <label>
            <span>
              <input
                type="checkbox"
                name="showGithubUrl"
                value="true"
                defaultChecked={profile.showGithubUrl}
              />{' '}
              Show GitHub
            </span>
          </label>
          <label>
            <span>
              <input
                type="checkbox"
                name="showTwitterUrl"
                value="true"
                defaultChecked={profile.showTwitterUrl}
              />{' '}
              Show Twitter / X
            </span>
          </label>
        </div>
      </div>

      <div className="form-grid__full">
        <div className="stack stack--sm">
          {state.error ? (
            <p className="form-message form-message--error" role="alert">
              {state.error}
            </p>
          ) : null}
          {showSuccessMessage && state.success ? (
            <p className="form-message form-message--success" role="status">
              {state.success}
            </p>
          ) : null}
          <div className="inline-actions">
            <ProfileSubmitButton isDirty={isDirty} />
            {isDirty ? <span className="muted">You have unsaved changes.</span> : null}
            {state.savedAt && !isDirty ? (
              <span className="muted" role="status" aria-live="polite">
                Saved just now at {formatSavedAt(state.savedAt)}
              </span>
            ) : null}
          </div>
        </div>
      </div>
    </form>
  );
}

function ProfileSubmitButton({ isDirty }: { isDirty: boolean }) {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" disabled={pending || !isDirty} aria-disabled={pending || !isDirty}>
      {pending ? 'Saving identity...' : 'Save identity'}
    </Button>
  );
}
