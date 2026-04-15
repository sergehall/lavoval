'use client';

import Link from 'next/link';
import { useActionState } from 'react';
import { useFormStatus } from 'react-dom';
import { loginAction, type AuthFormState } from '@/features/auth/actions';
import { ResendVerificationForm } from '@/features/auth/resend-verification-form';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

const initialAuthFormState: AuthFormState = {
  error: null,
  email: '',
  needsVerification: false,
};

export function LoginForm() {
  const [state, action] = useActionState(loginAction, initialAuthFormState);

  return (
    <form action={action} className="stack stack--md">
      <label>
        <span>Email</span>
        <Input
          type="email"
          name="email"
          placeholder="admin@lavoval.local"
          required
          defaultValue={state.email}
        />
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
      {state.error ? (
        <p className="form-message form-message--error" role="alert">
          {state.error}
        </p>
      ) : null}
      {state.error && !state.needsVerification && state.email ? (
        <div className="auth-guidance">
          <p className="muted">
            If this is your first time in this local Lavoval environment, create an account with{' '}
            <strong>{state.email}</strong> first.
          </p>
          <Link
            href={`/register?email=${encodeURIComponent(state.email)}`}
            className="button button--secondary button--full"
          >
            Create account with this email
          </Link>
        </div>
      ) : null}
      {state.needsVerification ? <ResendVerificationForm defaultEmail={state.email} /> : null}
      <LoginSubmitButton />
    </form>
  );
}

function LoginSubmitButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" fullWidth disabled={pending} aria-disabled={pending}>
      {pending ? 'Signing in...' : 'Enter Lavoval'}
    </Button>
  );
}
