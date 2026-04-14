'use client';

import { useActionState, useEffect, useState } from 'react';
import { useFormStatus } from 'react-dom';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { registerAction, type RegisterFormState } from '@/features/auth/actions';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';
import { RegistrationReminderModal } from '@/features/auth/registration-reminder-modal';

const initialRegisterFormState: RegisterFormState = {
  error: null,
  email: '',
  registered: false,
};

export function RegisterForm({
  mode = 'page',
  onSwitchToSignIn,
}: {
  mode?: 'page' | 'modal';
  onSwitchToSignIn?: () => void;
}) {
  const router = useRouter();
  const [state, action] = useActionState(registerAction, initialRegisterFormState);
  const [isReminderOpen, setIsReminderOpen] = useState(false);

  useEffect(() => {
    if (state.registered) {
      setIsReminderOpen(true);
    }
  }, [state.registered]);

  const handleReturnToSignIn = () => {
    setIsReminderOpen(false);
    if (onSwitchToSignIn) {
      onSwitchToSignIn();
      return;
    }
    router.push('/login');
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
          <Input type="email" name="email" required defaultValue={state.email} />
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
        <RegisterSubmitButton />
      </form>
      {mode === 'modal' ? (
        <button type="button" className="auth-link-button" onClick={onSwitchToSignIn}>
          Already part of Lavoval?
        </button>
      ) : (
        <Link href="/login" className="muted">
          Already part of Lavoval?
        </Link>
      )}
      <RegistrationReminderModal
        email={state.email}
        isOpen={isReminderOpen}
        onClose={() => setIsReminderOpen(false)}
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
