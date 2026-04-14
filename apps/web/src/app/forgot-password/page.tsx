import type { Metadata } from 'next';
import { AuthCard } from '@/features/auth/auth-card';

export const metadata: Metadata = {
  title: 'Password Recovery',
  description: 'Recover access to your Lavoval account and return to your skill exchange workspace.',
  robots: {
    index: false,
    follow: false,
  },
};

export default function ForgotPasswordPage() {
  return (
    <AuthCard
      title="Recovery flow placeholder"
      description="Account recovery will help people safely regain access to their marketplace identity, authored skills, and trusted exchange history."
    >
      <p className="muted">
        Add email delivery, rate limiting, reset tokens, and activity logging here without reshaping
        the rest of the product foundation.
      </p>
    </AuthCard>
  );
}
