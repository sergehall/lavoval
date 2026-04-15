'use server';

import { revalidatePath } from 'next/cache';
import { profileUpdateSchema } from '@lavoval/contracts';
import { requireSession, updateProfile } from '@/shared/api/server-client';

/** Parse a comma-separated form field into a trimmed, non-empty string array.
 *  Returns null (→ clears the field) when the input is blank. */
function parseArray(raw: FormDataEntryValue | null): string[] | null {
  if (!raw || typeof raw !== 'string' || raw.trim() === '') return null;
  const items = raw
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
  return items.length > 0 ? items : null;
}

/** Coerce a nullable string form value to null when blank. */
function nullableString(raw: FormDataEntryValue | null): string | null {
  if (!raw || typeof raw !== 'string' || raw.trim() === '') return null;
  return raw.trim();
}

export async function updateProfileAction(formData: FormData) {
  const session = await requireSession();

  // profileUpdateSchema (Zod) is the application-layer gate.
  // It rejects invalid URLs, oversized arrays, bad enum values, etc.
  // The Go backend re-validates independently, and the DB enforces
  // CHECK constraints as a final defence-in-depth layer.
  const payload = profileUpdateSchema.parse({
    firstName: formData.get('firstName'),
    lastName: formData.get('lastName'),
    bio: nullableString(formData.get('bio')),
    timezone: formData.get('timezone'),

    username: nullableString(formData.get('username')),
    avatarUrl: nullableString(formData.get('avatarUrl')),
    location: nullableString(formData.get('location')),

    skills: parseArray(formData.get('skills')),
    languages: parseArray(formData.get('languages')),

    websiteUrl: nullableString(formData.get('websiteUrl')),
    linkedinUrl: nullableString(formData.get('linkedinUrl')),
    githubUrl: nullableString(formData.get('githubUrl')),
    twitterUrl: nullableString(formData.get('twitterUrl')),

    availabilityStatus: formData.get('availabilityStatus') ?? 'open',
    // Checkbox: present in FormData only when checked.
    isPublicProfile: formData.get('isPublicProfile') === 'true',
  });

  await updateProfile(session.accessToken, payload);
  revalidatePath('/account/profile');
}
