'use client';

import Link from 'next/link';
import { useActionState, useEffect, useState } from 'react';
import { useFormStatus } from 'react-dom';
import { loginAction, type AuthFormState } from '@/features/auth/actions';
import { ResendVerificationForm } from '@/features/auth/resend-verification-form';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

const initialAuthFormState: AuthFormState = {
  error: null,
  email: '',
  needsVerification: false,
  mfaRequired: false,
  challengeId: null,
  recoveryMode: false,
};

export function LoginForm() {
  const [state, action] = useActionState(loginAction, initialAuthFormState);
  const [recoveryMode, setRecoveryMode] = useState(false);

  useEffect(() => {
    setRecoveryMode(Boolean(state.recoveryMode));
  }, [state.recoveryMode]);

  return (
    <form action={action} className="stack stack--md">
      <input type="hidden" name="challengeId" value={state.challengeId ?? ''} />
      <input type="hidden" name="recoveryMode" value={recoveryMode ? 'true' : 'false'} />
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
          disabled={Boolean(state.mfaRequired)}
        />
      </label>
      {state.mfaRequired ? (
        <>
          {!recoveryMode ? (
            <label>
              <span>Authenticator code</span>
              <Input
                name="code"
                inputMode="numeric"
                pattern="[0-9]{6}"
                minLength={6}
                maxLength={6}
                placeholder="123456"
                required
              />
            </label>
          ) : (
            <label>
              <span>Recovery code</span>
              <Input name="recoveryCode" placeholder="ABCD-EFGH-IJKL" required />
            </label>
          )}
          <button
            type="button"
            className="button button--ghost"
            onClick={() => setRecoveryMode((value) => !value)}
          >
            {recoveryMode ? 'Use authenticator code instead' : 'Use a recovery code instead'}
          </button>
        </>
      ) : (
        <div className="inline-actions">
          <span />
          <Link href="/forgot-password" className="muted">
            Forgot password?
          </Link>
        </div>
      )}
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
