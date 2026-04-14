import Link from 'next/link';
import { loginAction } from '@/features/auth/actions';
import { AuthCard } from '@/features/auth/auth-card';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

export default function LoginPage() {
  return (
    <AuthCard
      title="Sign in"
      description="Return to your skill exchange space to publish expertise, discover people, and manage your marketplace identity."
    >
      <form action={loginAction} className="stack stack--md">
        <label>
          <span>Email</span>
          <Input type="email" name="email" placeholder="admin@lavoval.local" required />
        </label>
        <label>
          <span>Password</span>
          <Input
            type="password"
            name="password"
            placeholder="Your secure password"
            required
            minLength={8}
          />
        </label>
        <Button type="submit" fullWidth>
          Enter Lavoval
        </Button>
      </form>
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
