import type { Metadata } from 'next';
import { AuthCard } from '@/features/auth/auth-card';
import Link from 'next/link';
import { ForgotPasswordForm } from '@/features/auth/forgot-password-form';
import { signInHref } from '@/shared/lib/auth-navigation';

export const metadata: Metadata = {
  title: 'Password Recovery',
  description:
    'Recover access to your Lavoval account and return to your skill exchange workspace.',
  robots: {
    index: false,
    follow: false,
  },
};

export default async function ForgotPasswordPage({
  searchParams,
}: {
  searchParams?: Promise<Record<string, string | string[] | undefined>>;
}) {
  const params = searchParams ? await searchParams : {};
  const initialEmail = Array.isArray(params.email) ? (params.email[0] ?? '') : (params.email ?? '');

  return (
    <AuthCard
      title="Reset your password"
      description="Enter the email address tied to your Lavoval account and we'll send a secure reset link if it exists in our system."
    >
      <ForgotPasswordForm initialEmail={initialEmail} />
      <div className="inline-actions">
        <Link href={signInHref} className="muted">
          Return to sign in
        </Link>
      </div>
    </AuthCard>
  );
}
