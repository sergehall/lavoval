'use client';

import { useActionState } from 'react';
import { useFormStatus } from 'react-dom';
import { loginAction, type AuthFormState } from '@/features/auth/actions';
import { Button } from '@/shared/ui/button';
import { Input } from '@/shared/ui/input';

const initialAuthFormState: AuthFormState = {
  error: null,
  email: '',
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
