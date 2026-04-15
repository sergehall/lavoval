import type { Metadata } from 'next';
import Link from 'next/link';
import { AuthCard } from '@/features/auth/auth-card';
import { ForgotPasswordForm } from '@/features/auth/forgot-password-form';
import { ResetPasswordForm } from '@/features/auth/reset-password-form';
import { signInHref } from '@/shared/lib/auth-navigation';

export const metadata: Metadata = {
  title: 'Set A New Password',
  description: 'Use your secure Lavoval recovery link to set a new password and return to your workspace.',
  robots: {
    index: false,
    follow: false,
  },
};

export default async function ResetPasswordPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = searchParams ? await searchParams : {};
  const token = Array.isArray(params.token) ? (params.token[0] ?? '') : (params.token ?? '');
  const email = Array.isArray(params.email) ? (params.email[0] ?? '') : (params.email ?? '');

  if (!token) {
    return (
      <AuthCard
        title="Missing reset link"
        description="This password reset page needs a valid recovery token. Request a fresh email below."
      >
        <ForgotPasswordForm initialEmail={email} />
        <div className="inline-actions">
          <Link href={signInHref} className="muted">
            Return to sign in
          </Link>
        </div>
      </AuthCard>
    );
  }

  return (
    <AuthCard
      title="Choose a new password"
      description="Set a strong new password for your Lavoval account. This recovery link can only be used once."
    >
      <ResetPasswordForm token={token} email={email} />
    </AuthCard>
  );
}
