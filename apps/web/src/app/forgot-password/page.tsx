import { AuthCard } from '@/features/auth/auth-card';

export default function ForgotPasswordPage() {
  return (
    <AuthCard
      title="Recovery flow placeholder"
      description="The foundation is ready for email verification, magic links, or password reset orchestration."
    >
      <p className="muted">
        Add email delivery, rate limiting, reset tokens, and activity logging here without reshaping
        the rest of the application.
      </p>
    </AuthCard>
  );
}
