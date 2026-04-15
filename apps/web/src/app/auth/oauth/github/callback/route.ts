import { NextResponse } from 'next/server';
import { completeGitHubOAuth, persistSession } from '@/shared/api/server-client';
import { env } from '@/shared/config/env';

function redirectToSignIn(message: string) {
  const target = new URL('/?auth=sign-in', env.appUrl);
  target.searchParams.set('auth', 'sign-in');
  target.searchParams.set('oauthError', message);
  return NextResponse.redirect(target);
}

export async function GET(request: Request) {
  const requestURL = new URL(request.url);
  const error = requestURL.searchParams.get('error');
  const code = requestURL.searchParams.get('code');
  const state = requestURL.searchParams.get('state');

  if (error) {
    return redirectToSignIn('GitHub sign-in was cancelled or denied.');
  }

  if (!code || !state) {
    return redirectToSignIn('GitHub sign-in returned an incomplete callback.');
  }

  try {
    const response = await completeGitHubOAuth({ code, state });
    await persistSession(response.data);

    return NextResponse.redirect(
      new URL(response.data.user.role === 'admin' ? '/admin' : '/account', env.appUrl),
    );
  } catch (oauthError) {
    if (oauthError instanceof Error) {
      return redirectToSignIn(oauthError.message);
    }

    return redirectToSignIn('Could not finish GitHub sign-in right now.');
  }
}
