'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { AuthModal } from '@/features/auth/auth-modal';

export function GuestEntryBar() {
  const router = useRouter();
  const [query, setQuery] = useState('');
  const [authMode, setAuthMode] = useState<'sign-in' | 'sign-up'>('sign-in');
  const [isAuthOpen, setIsAuthOpen] = useState(false);

  return (
    <>
      <div className="site-auth-actions">
        <form
          className="header-search"
          onSubmit={(event) => {
            event.preventDefault();
            const trimmed = query.trim();
            router.push(trimmed ? `/skills?q=${encodeURIComponent(trimmed)}` : '/skills');
          }}
        >
          <input
            aria-label="Search skills and creators"
            className="input header-search__input"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search skills, creators, or topics"
          />
        </form>
        <button
          type="button"
          className="site-nav__link site-nav__link--subtle"
          onClick={() => {
            setAuthMode('sign-in');
            setIsAuthOpen(true);
          }}
        >
          Sign in
        </button>
        <button
          type="button"
          className="site-nav__link site-nav__link--cta"
          onClick={() => {
            setAuthMode('sign-up');
            setIsAuthOpen(true);
          }}
        >
          Create account
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
