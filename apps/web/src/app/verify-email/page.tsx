import type { Metadata } from 'next';
import Link from 'next/link';
import { verifyEmailAction } from '@/features/auth/actions';
import { ResendVerificationForm } from '@/features/auth/resend-verification-form';
import { AuthCard } from '@/features/auth/auth-card';

export const metadata: Metadata = {
  title: 'Confirm Your Email',
  description: 'Confirm your Lavoval email address to unlock sign in and activate your account.',
  robots: {
    index: false,
    follow: false,
  },
};

export default async function VerifyEmailPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = searchParams ? await searchParams : {};
  const token = Array.isArray(params.token) ? params.token[0] ?? '' : params.token ?? '';
  const email = Array.isArray(params.email) ? params.email[0] ?? '' : params.email ?? '';
  const result = await verifyEmailAction(token);

  return (
    <AuthCard title={result.title} description={result.message}>
      <div className="stack stack--md">
        {result.ok ? (
          <>
            <div className="form-message form-message--success" role="status">
              {result.email}
            </div>
            <Link href="/login" className="site-nav__link site-nav__link--cta">
              Continue to sign in
            </Link>
          </>
        ) : (
          <>
            <ResendVerificationForm defaultEmail={email} compact />
            <Link href="/login" className="muted">
              Back to sign in
            </Link>
          </>
        )}
      </div>
    </AuthCard>
  );
}
