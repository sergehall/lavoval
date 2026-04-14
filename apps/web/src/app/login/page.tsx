import type { Metadata } from 'next';
import Link from 'next/link';
import { AuthCard } from '@/features/auth/auth-card';
import { LoginForm } from '@/features/auth/login-form';

export const metadata: Metadata = {
  title: 'Sign In',
  description: 'Sign in to your Lavoval account to manage skills, runs, and marketplace activity.',
  robots: {
    index: false,
    follow: false,
  },
};

export default function LoginPage() {
  return (
    <AuthCard
      title="Sign in"
      description="Return to your skill exchange space to publish expertise, discover people, and manage your marketplace identity."
    >
      <LoginForm />
      <div className="inline-actions">
        <Link href="/register" className="muted">
          Join the marketplace
        </Link>
        <Link href="/forgot-password" className="muted">
          Forgot password?
        </Link>
      </div>
    </AuthCard>
  );
}
