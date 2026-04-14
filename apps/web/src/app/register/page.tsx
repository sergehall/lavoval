import type { Metadata } from 'next';
import { AuthCard } from '@/features/auth/auth-card';
import { RegisterForm } from '@/features/auth/register-form';

export const metadata: Metadata = {
  title: 'Create Account',
  description:
    'Create a Lavoval account to publish expertise, discover specialists, and join the skill exchange marketplace.',
  robots: {
    index: false,
    follow: false,
  },
};

export default function RegisterPage() {
  return (
    <AuthCard
      title="Create account"
      description="Create your presence in the marketplace and start exchanging real-world skills in the age of AI."
    >
      <RegisterForm />
    </AuthCard>
  );
}
