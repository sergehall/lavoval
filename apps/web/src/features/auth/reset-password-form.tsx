'use client';

import Link from 'next/link';
import { useActionState } from 'react';
import { useFormStatus } from 'react-dom';
import { resetPasswordAction, type ResetPasswordState } from '@/features/auth/actions';
import { signInHref } from '@/shared/lib/auth-navigation';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

const initialResetPasswordState: ResetPasswordState = {
  error: null,
  success: null,
  email: '',
};

export function ResetPasswordForm({
  token,
  email = '',
}: {
  token: string;
  email?: string;
}) {
  const [state, action] = useActionState(resetPasswordAction, {
    ...initialResetPasswordState,
    email,
  });

  return (
    <form action={action} className="stack stack--md">
      <input type="hidden" name="token" value={token} />
      <input type="hidden" name="email" value={email} />
      <label>
        <span>New password</span>
        <Input
          type="password"
          name="newPassword"
          placeholder="Choose a strong new password"
          required
          minLength={12}
        />
      </label>
      {state.error ? (
        <div className="stack stack--sm">
          <p className="form-message form-message--error" role="alert">
            {state.error}
          </p>
          <Link
            href={email ? `/forgot-password?email=${encodeURIComponent(email)}` : '/forgot-password'}
            className="muted"
          >
            Request a fresh reset link
          </Link>
        </div>
      ) : null}
      {state.success ? (
        <div className="stack stack--sm">
          <p className="form-message form-message--success" role="status">
            {state.success}
          </p>
          <Link href={signInHref} className="button button--secondary button--full">
            Return to sign in
          </Link>
        </div>
      ) : (
        <ResetPasswordSubmitButton />
      )}
    </form>
  );
}

function ResetPasswordSubmitButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" fullWidth disabled={pending} aria-disabled={pending}>
      {pending ? 'Updating password...' : 'Update password'}
    </Button>
  );
}
