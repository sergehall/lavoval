'use client';

import { useState } from 'react';
import { AuthModal } from '@/features/auth/auth-modal';

export function GuestEntryBar() {
  const [authMode, setAuthMode] = useState<'sign-in' | 'sign-up'>('sign-in');
  const [isAuthOpen, setIsAuthOpen] = useState(false);

  return (
    <>
      <div className="site-auth-actions">
        <button
          type="button"
          className="guest-sign-in"
          onClick={() => {
            setAuthMode('sign-in');
            setIsAuthOpen(true);
          }}
        >
          Sign in
        </button>
      </div>
      <AuthModal
        isOpen={isAuthOpen}
        mode={authMode}
        onClose={() => setIsAuthOpen(false)}
        onChangeMode={setAuthMode}
      />
    </>
  );
}
