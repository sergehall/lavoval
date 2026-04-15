'use client';

import { useActionState } from 'react';
import { useFormStatus } from 'react-dom';
import { forgotPasswordAction, type ForgotPasswordState } from '@/features/auth/actions';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

const initialForgotPasswordState: ForgotPasswordState = {
  error: null,
  success: null,
  email: '',
};

export function ForgotPasswordForm({ initialEmail = '' }: { initialEmail?: string }) {
  const [state, action] = useActionState(forgotPasswordAction, {
    ...initialForgotPasswordState,
    email: initialEmail,
  });

  return (
    <form action={action} className="stack stack--md">
      <label>
        <span>Email</span>
        <Input
          type="email"
          name="email"
          placeholder="you@company.com"
          required
          defaultValue={state.email || initialEmail}
        />
      </label>
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
      <ForgotPasswordSubmitButton />
    </form>
  );
}

function ForgotPasswordSubmitButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" fullWidth disabled={pending} aria-disabled={pending}>
      {pending ? 'Sending reset link...' : 'Send reset link'}
    </Button>
  );
}
