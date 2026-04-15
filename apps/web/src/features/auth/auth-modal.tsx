'use client';

import { useEffect, useState } from 'react';
import { LoginForm } from '@/features/auth/login-form';
import { RegisterForm } from '@/features/auth/register-form';
import { RegistrationReminderModal } from '@/features/auth/registration-reminder-modal';

export function AuthModal({
  isOpen,
  mode,
  onClose,
  onChangeMode,
}: {
  isOpen: boolean;
  mode: 'sign-in' | 'sign-up';
  onClose: () => void;
  onChangeMode: (mode: 'sign-in' | 'sign-up') => void;
}) {
  const [registeredEmail, setRegisteredEmail] = useState<string | null>(null);

  const handleClose = () => {
    setRegisteredEmail(null);
    onClose();
  };

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        handleClose();
      }
    };

    document.body.style.overflow = 'hidden';
    window.addEventListener('keydown', handleEscape);
    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen]); // eslint-disable-line react-hooks/exhaustive-deps

  if (!isOpen) {
    return null;
  }

  if (registeredEmail) {
    return (
      <RegistrationReminderModal
        isOpen
        email={registeredEmail}
        onClose={handleClose}
        onShowSignIn={() => {
          setRegisteredEmail(null);
          onChangeMode('sign-in');
        }}
      />
    );
  }

  return (
    <div className="auth-modal auth-modal--visible" role="dialog" aria-modal="true">
      <button
        type="button"
        className="auth-modal__overlay"
        aria-label="Close authentication dialog"
        onClick={handleClose}
      />
      <div className="auth-modal__panel" key={mode}>
        <button type="button" className="auth-modal__close" onClick={handleClose} aria-label="Close">
          ×
        </button>
        <div className="auth-modal__glow" />
        <div className="stack stack--lg auth-modal__content">
          {mode === 'sign-in' ? (
            <>
              <div className="stack stack--sm">
                <span className="badge badge--warning">Member access</span>
                <h2>Sign in to your skill exchange account</h2>
                <p className="muted">
                  Use your email and password to return to authored skill offers, runtime history,
                  and marketplace identity.
                </p>
              </div>
              <LoginForm key="sign-in-form" />
              <p className="muted">
                Need an account?{' '}
                <button
                  type="button"
                  className="auth-link-button"
                  onClick={() => onChangeMode('sign-up')}
                >
                  Create one
                </button>
              </p>
            </>
          ) : (
            <>
              <div className="stack stack--sm">
                <span className="badge badge--success">New creator</span>
                <h2>Create your Lavoval account</h2>
                <p className="muted">
                  Join the marketplace, publish expertise, and start packaging practical know-how
                  into skill offers for the AI era.
                </p>
              </div>
              <RegisterForm
                key="sign-up-form"
                mode="modal"
                onSwitchToSignIn={() => onChangeMode('sign-in')}
                onRegistered={setRegisteredEmail}
              />
            </>
          )}
        </div>
      </div>
    </div>
  );
}
