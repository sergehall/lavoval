import Link from 'next/link';
import { registerAction } from '@/features/auth/actions';
import { AuthCard } from '@/features/auth/auth-card';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

export default function RegisterPage() {
  return (
    <AuthCard
      title="Create account"
      description="Register a new operator account for skill discovery and delivery."
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
          Create account
        </Button>
      </form>
      <Link href="/login" className="muted">
        Already have an account?
      </Link>
    </AuthCard>
  );
}
