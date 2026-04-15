'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import Link from 'next/link';
import type { SessionUser } from '@lavoval/contracts';
import { logoutAction } from '@/features/auth/actions';

function getDisplayName(user: SessionUser) {
  const fullName = [user.firstName, user.lastName].filter(Boolean).join(' ').trim();
  if (fullName) {
    return fullName;
  }

  return user.email.split('@')[0] ?? 'Lavoval member';
}

function getInitial(user: SessionUser) {
  return getDisplayName(user).charAt(0).toUpperCase() || 'L';
}

export function AuthenticatedEntryBar({ user }: { user: SessionUser }) {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  const displayName = useMemo(() => getDisplayName(user), [user]);
  const initial = useMemo(() => getInitial(user), [user]);
  const roleLabel = user.role === 'admin' ? 'ADMIN' : 'USER';

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handlePointerDown = (event: MouseEvent) => {
      if (!menuRef.current?.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handlePointerDown);
    window.addEventListener('keydown', handleEscape);

    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      window.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen]);

  return (
    <div className="auth-entry" ref={menuRef}>
      <button
        type="button"
        className={`account-trigger${isOpen ? ' account-trigger--open' : ''}`}
        onClick={() => setIsOpen((open) => !open)}
        aria-haspopup="menu"
        aria-expanded={isOpen}
      >
        <span className="account-trigger__avatar">{initial}</span>
        <span className="account-trigger__identity">
          <span className="account-trigger__headline">
            <span className="account-trigger__name">{displayName}</span>
            <span className="account-trigger__role">{roleLabel}</span>
          </span>
          <span className="account-trigger__email">{user.email}</span>
        </span>
        <span className="account-trigger__chevron" aria-hidden="true">
          <svg viewBox="0 0 20 20">
            <path d="m5 7 5 6 5-6" />
          </svg>
        </span>
      </button>

      {isOpen ? (
        <div className="account-menu" role="menu">
          <div className="account-menu__label">Account</div>
          <Link
            href="/account"
            className="account-menu__item"
            role="menuitem"
            onClick={() => setIsOpen(false)}
          >
            Cabinet
          </Link>
          {user.role === 'admin' ? (
            <Link
              href="/admin"
              className="account-menu__item"
              role="menuitem"
              onClick={() => setIsOpen(false)}
            >
              Governance
            </Link>
          ) : null}
          <form action={logoutAction}>
            <button
              type="submit"
              className="account-menu__item account-menu__item--danger"
              role="menuitem"
            >
              <span className="account-menu__logout-icon" aria-hidden="true">
                <svg viewBox="0 0 20 20">
                  <path d="M8 4H4.8A1.8 1.8 0 0 0 3 5.8v8.4A1.8 1.8 0 0 0 4.8 16H8" />
                  <path d="M10 10h7" />
                  <path d="m14 6 4 4-4 4" />
                </svg>
              </span>
              Log out
            </button>
          </form>
        </div>
      ) : null}
    </div>
  );
}
