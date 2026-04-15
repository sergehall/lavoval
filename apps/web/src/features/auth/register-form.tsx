'use client';

import { useActionState, useEffect, useState } from 'react';
import { useFormStatus } from 'react-dom';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { registerAction, type RegisterFormState } from '@/features/auth/actions';
import { env } from '@/shared/config/env';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { RegistrationReminderModal } from '@/features/auth/registration-reminder-modal';
import { signInHref } from '@/shared/lib/auth-navigation';

const initialRegisterFormState: RegisterFormState = {
  error: null,
  email: '',
  registered: false,
};

export function RegisterForm({
  mode = 'page',
  onSwitchToSignIn,
  onRegistered,
  initialEmail = '',
}: {
  mode?: 'page' | 'modal';
  onSwitchToSignIn?: () => void;
  onRegistered?: () => void;
  initialEmail?: string;
}) {
  const router = useRouter();
  const [state, action] = useActionState(registerAction, initialRegisterFormState);
  const [dismissedReminderForEmail, setDismissedReminderForEmail] = useState<string | null>(null);

  useEffect(() => {
    if (state.registered) {
      onRegistered?.();
    }
  }, [state.registered, onRegistered]);
  const isReminderOpen = state.registered && state.email !== dismissedReminderForEmail;
  const googleOAuthStartURL = `${env.apiUrl}/api/v1/auth/oauth/google/start`;
  const githubOAuthStartURL = `${env.apiUrl}/api/v1/auth/oauth/github/start`;

  const handleReturnToSignIn = () => {
    setDismissedReminderForEmail(state.email);
    if (onSwitchToSignIn) {
      onSwitchToSignIn();
      return;
    }
    router.push(signInHref);
  };

  return (
    <>
      <form action={action} className="stack stack--md">
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
          <Input type="email" name="email" required defaultValue={state.email || initialEmail} />
        </label>
        <label>
          <span>Password</span>
          <Input type="password" name="password" required minLength={12} />
        </label>
        {state.error ? (
          <p className="form-message form-message--error" role="alert">
            {state.error}
          </p>
        ) : null}
        <div className="stack stack--sm">
          <a href={googleOAuthStartURL} className="button button--secondary button--full">
            Continue with Google
          </a>
          <a href={githubOAuthStartURL} className="button button--secondary button--full">
            Continue with GitHub
          </a>
        </div>
        <RegisterSubmitButton />
      </form>
      {mode === 'modal' ? (
        <button type="button" className="auth-link-button" onClick={onSwitchToSignIn}>
          Already part of Lavoval?
        </button>
      ) : (
        <Link href={signInHref} className="muted">
          Already part of Lavoval?
        </Link>
      )}
      <RegistrationReminderModal
        email={state.email}
        isOpen={isReminderOpen}
        onClose={() => setDismissedReminderForEmail(state.email)}
        onShowSignIn={handleReturnToSignIn}
      />
    </>
  );
}

function RegisterSubmitButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" fullWidth disabled={pending} aria-disabled={pending}>
      {pending ? 'Creating account...' : 'Create my profile'}
    </Button>
  );
}
