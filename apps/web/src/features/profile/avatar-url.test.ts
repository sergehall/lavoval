import { describe, expect, it } from 'vitest';
import {
  avatarFallbackColorSchema,
  isSafeAvatarUrl,
  profileUpdateSchema,
} from '@lavoval/contracts';

describe('avatar URL safety', () => {
  it('allows only trusted HTTPS avatar hosts without credentials or custom ports', () => {
    expect(isSafeAvatarUrl('https://avatars.githubusercontent.com/u/60080971?v=4')).toBe(true);
    expect(isSafeAvatarUrl('https://secure.gravatar.com/avatar/hash?s=96')).toBe(true);
    expect(isSafeAvatarUrl('http://avatars.githubusercontent.com/u/60080971')).toBe(false);
    expect(isSafeAvatarUrl('https://user:pass@avatars.githubusercontent.com/u/60080971')).toBe(
      false,
    );
    expect(isSafeAvatarUrl('https://avatars.githubusercontent.com:8443/u/60080971')).toBe(false);
    expect(isSafeAvatarUrl('data:image/png;base64,abcd')).toBe(false);
    expect(isSafeAvatarUrl('https://avatars.githubusercontent.com.evil.test/avatar.png')).toBe(
      false,
    );
  });

  it('enforces avatar URL policy in profile updates', () => {
    const validPayload = {
      firstName: 'Serge',
      lastName: 'Hall',
      bio: null,
      timezone: 'America/Los_Angeles',
      avatarUrl: 'https://lh3.googleusercontent.com/a/example',
    };

    expect(profileUpdateSchema.safeParse(validPayload).success).toBe(true);
    expect(
      profileUpdateSchema.safeParse({
        ...validPayload,
        avatarUrl: 'https://example.com/avatar.png',
      }).success,
    ).toBe(false);
  });

  it('keeps fallback avatar color values on a closed allowlist', () => {
    expect(avatarFallbackColorSchema.safeParse('cyan').success).toBe(true);
    expect(avatarFallbackColorSchema.safeParse('url(javascript:alert(1))').success).toBe(false);
  });
});
