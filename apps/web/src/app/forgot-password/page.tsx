import { AuthCard } from '@/features/auth/auth-card';

export default function ForgotPasswordPage() {
  return (
    <AuthCard
      title="Recovery flow placeholder"
      description="Account recovery will help people safely regain access to their marketplace identity, authored skills, and trusted exchange history."
    >
      <p className="muted">
        Add email delivery, rate limiting, reset tokens, and activity logging here without
        reshaping the rest of the product foundation.
      </p>
    </AuthCard>
  );
}
