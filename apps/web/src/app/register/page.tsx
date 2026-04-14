import type { Metadata } from 'next';
import Link from 'next/link';
import { registerAction } from '@/features/auth/actions';
import { AuthCard } from '@/features/auth/auth-card';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

export const metadata: Metadata = {
  title: 'Create Account',
  description: 'Create a Lavoval account to publish expertise, discover specialists, and join the skill exchange marketplace.',
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
      <form action={registerAction} className="stack stack--md">
        <div className="form-grid">
          <label>
            <span>First name</span>
            <Input name="firstName" required minLength={2} />
          </label>
          <label>
            <span>Last name</span>
            <Input name="lastName" required minLength={2} />
          </label>
        </div>
        <label>
          <span>Email</span>
          <Input type="email" name="email" required />
        </label>
        <label>
          <span>Password</span>
          <Input type="password" name="password" required minLength={12} />
        </label>
        <Button type="submit" fullWidth>
          Create my profile
        </Button>
      </form>
      <Link href="/login" className="muted">
        Already part of Lavoval?
      </Link>
    </AuthCard>
  );
}
