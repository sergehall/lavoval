'use client';

import { useActionState } from 'react';
import { useFormStatus } from 'react-dom';
import { resendVerificationAction, type VerificationRequestState } from '@/features/auth/actions';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

const initialVerificationState: VerificationRequestState = {
  error: null,
  success: null,
  email: '',
};

export function ResendVerificationForm({
  defaultEmail = '',
  compact = false,
}: {
  defaultEmail?: string;
  compact?: boolean;
}) {
  const [state, action] = useActionState(resendVerificationAction, {
    ...initialVerificationState,
    email: defaultEmail,
  });

  return (
    <form action={action} className={compact ? 'stack stack--sm' : 'stack stack--md resend-verification'}>
      <div className="stack stack--xs">
        <strong>Need a fresh confirmation email?</strong>
        <span className="muted">
          Enter your address and we&apos;ll send a new Google-backed verification link.
        </span>
      </div>
      <div className={compact ? 'form-grid form-grid--single' : 'form-inline'}>
        <Input
          type="email"
          name="email"
          placeholder="you@company.com"
          required
          defaultValue={state.email || defaultEmail}
        />
        <ResendVerificationButton />
      </div>
      {state.error ? (
        <p className="form-message form-message--error" role="alert">
          {state.error}
        </p>
      ) : null}
      {state.success ? (
        <p className="form-message form-message--success" role="status">
          {state.success}
        </p>
      ) : null}
    </form>
  );
}

function ResendVerificationButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" variant="secondary" disabled={pending} aria-disabled={pending}>
      {pending ? 'Sending...' : 'Resend link'}
    </Button>
  );
}
