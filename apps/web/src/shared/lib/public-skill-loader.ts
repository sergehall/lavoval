import { fetchSkills } from '@/shared/api/server-client';

const PUBLIC_SKILLS_TIMEOUT_MS = 5000;

export async function loadPublicSkills(timeoutMs = PUBLIC_SKILLS_TIMEOUT_MS) {
  const timeoutPromise = new Promise<never>((_, reject) => {
    const timer = setTimeout(() => {
      reject(new Error('public skills request timed out'));
    }, timeoutMs);

    // avoid keeping the process alive for the timer in server runtimes that support it
    if (
      typeof timer === 'object' &&
      timer &&
      'unref' in timer &&
      typeof timer.unref === 'function'
    ) {
      timer.unref();
    }
  });

  try {
    const response = await Promise.race([fetchSkills(), timeoutPromise]);
    return response.data;
  } catch {
    return [];
  }
}
