'use server';

import { revalidatePath } from 'next/cache';
import { ZodError } from 'zod';
import { profileUpdateSchema } from '@lavoval/contracts';
import {
  ApiError as ServerApiError,
  requireSession,
  updateProfile,
} from '@/shared/api/server-client';

export type ProfileFormState = {
  error: string | null;
  success: string | null;
  savedAt: string | null;
};

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

export async function updateProfileAction(
  _previousState: ProfileFormState,
  formData: FormData,
): Promise<ProfileFormState> {
  try {
    const session = await requireSession();

    const parsed = profileUpdateSchema.safeParse({
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
      isPublicProfile: formData.get('isPublicProfile') === 'true',
      showAvatar: formData.get('showAvatar') === 'true',
      showBio: formData.get('showBio') === 'true',
      showLocation: formData.get('showLocation') === 'true',
      showSkills: formData.get('showSkills') === 'true',
      showLanguages: formData.get('showLanguages') === 'true',
      showAvailabilityStatus: formData.get('showAvailabilityStatus') === 'true',
      showWebsiteUrl: formData.get('showWebsiteUrl') === 'true',
      showLinkedinUrl: formData.get('showLinkedinUrl') === 'true',
      showGithubUrl: formData.get('showGithubUrl') === 'true',
      showTwitterUrl: formData.get('showTwitterUrl') === 'true',
    });

    if (!parsed.success) {
      return {
        error: parsed.error.issues[0]?.message ?? 'Check the form fields and try again.',
        success: null,
        savedAt: null,
      };
    }

    await updateProfile(session.accessToken, parsed.data);
    revalidatePath('/account/profile');
    revalidatePath('/account');

    return {
      error: null,
      success: 'Identity saved successfully.',
      savedAt: new Date().toISOString(),
    };
  } catch (error) {
    if (error instanceof ServerApiError) {
      return {
        error: error.message || 'Could not save your identity right now.',
        success: null,
        savedAt: null,
      };
    }

    if (error instanceof ZodError) {
      return {
        error: error.message || 'Check the form fields and try again.',
        success: null,
        savedAt: null,
      };
    }

    return {
      error: 'Could not save your identity right now.',
      success: null,
      savedAt: null,
    };
  }
}
