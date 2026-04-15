'use client';

import Link from 'next/link';
import { useState } from 'react';
import { AuthModal } from '@/features/auth/auth-modal';

export function GuestEntryBar() {
  const [authMode, setAuthMode] = useState<'sign-in' | 'sign-up'>('sign-in');
  const [isAuthOpen, setIsAuthOpen] = useState(false);

  return (
    <>
      <div className="site-auth-actions">
        <Link href="/skills" aria-label="Search skills" className="header-icon-button">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="11" cy="11" r="6.8" />
            <path d="M16.2 16.2 21 21" />
          </svg>
        </Link>
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
