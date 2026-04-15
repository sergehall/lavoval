export const signInHref = '/?auth=sign-in';
export const signUpHref = '/?auth=sign-up';

export function isAuthMode(value: string | null): value is 'sign-in' | 'sign-up' {
  return value === 'sign-in' || value === 'sign-up';
}
