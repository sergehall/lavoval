'use client';

import Link from 'next/link';
import { useActionState, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useFormStatus } from 'react-dom';
import { loginAction, type AuthFormState } from '@/features/auth/actions';
import { ResendVerificationForm } from '@/features/auth/resend-verification-form';
import { env } from '@/shared/config/env';
import { signUpHref } from '@/shared/lib/auth-navigation';
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
  const searchParams = useSearchParams();
  const oauthError = searchParams.get('oauthError');
  const googleOAuthStartURL = `${env.apiUrl}/api/v1/auth/oauth/google/start`;
  const githubOAuthStartURL = `${env.apiUrl}/api/v1/auth/oauth/github/start`;

  return (
    <form action={action} className="stack stack--md">
      <label>
        <span>Email</span>
        <Input
          type="email"
          name="email"
          placeholder="yoursemail@lavoval.com"
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
        <MFAStep
          key={state.challengeId ?? 'mfa'}
          challengeId={state.challengeId ?? null}
          initialRecoveryMode={Boolean(state.recoveryMode)}
        />
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
      ) : oauthError ? (
        <p className="form-message form-message--error" role="alert">
          {oauthError}
        </p>
      ) : null}
      {!state.mfaRequired ? (
        <div className="stack stack--sm">
          <a href={googleOAuthStartURL} className="button button--secondary button--full">
            Continue with Google
          </a>
          <a href={githubOAuthStartURL} className="button button--secondary button--full">
            Continue with GitHub
          </a>
        </div>
      ) : null}
      {state.error && !state.needsVerification && state.email ? (
        <div className="auth-guidance">
          <p className="muted">
            If this is your first time in this local Lavoval environment, create an account with{' '}
            <strong>{state.email}</strong> first.
          </p>
          <Link
            href={`${signUpHref}&email=${encodeURIComponent(state.email)}`}
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

function MFAStep({
  challengeId,
  initialRecoveryMode,
}: {
  challengeId: string | null;
  initialRecoveryMode: boolean;
}) {
  const [recoveryMode, setRecoveryMode] = useState(initialRecoveryMode);

  return (
    <div className="stack stack--md">
      <input type="hidden" name="challengeId" value={challengeId ?? ''} />
      <input type="hidden" name="recoveryMode" value={recoveryMode ? 'true' : 'false'} />
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
    </div>
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
