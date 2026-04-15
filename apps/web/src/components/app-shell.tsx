'use client';

import Link from 'next/link';
import { type PropsWithChildren, useEffect, useRef, useState } from 'react';
import { usePathname } from 'next/navigation';
import type { SessionUser } from '@lavoval/contracts';
import { AuthModal } from '@/features/auth/auth-modal';
import { logoutAction } from '@/features/auth/actions';
import { cabinetNavigation, publicNavigation } from '@/shared/lib/navigation';
import { AuthenticatedEntryBar } from '@/components/authenticated-entry-bar';
import { GuestEntryBar } from '@/components/guest-entry-bar';

export function AppShell({ children, user }: PropsWithChildren<{ user?: SessionUser }>) {
  const pathname = usePathname();
  const workspaceNavigation = user ? cabinetNavigation(user) : [];
  const showWorkspaceNav = Boolean(user);
  const [isMobileNavOpen, setIsMobileNavOpen] = useState(false);
  const [authMode, setAuthMode] = useState<'sign-in' | 'sign-up'>('sign-in');
  const [isAuthOpen, setIsAuthOpen] = useState(false);
  const mobileMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    setIsMobileNavOpen(false);
  }, [pathname]);

  useEffect(() => {
    if (!isMobileNavOpen) {
      return;
    }

    const handlePointerDown = (event: MouseEvent) => {
      if (!mobileMenuRef.current?.contains(event.target as Node)) {
        setIsMobileNavOpen(false);
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        setIsMobileNavOpen(false);
      }
    };

    document.addEventListener('mousedown', handlePointerDown);
    window.addEventListener('keydown', handleEscape);

    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      window.removeEventListener('keydown', handleEscape);
    };
  }, [isMobileNavOpen]);

  return (
    <div className="app-frame">
      <header className="site-header">
        <Link href="/" className="brand-mark">
          <span className="brand-mark__label">Lavoval</span>
          <span className="brand-mark__caption">Human skill exchange for the AI era</span>
        </Link>
        <nav className="site-nav site-nav--primary">
          {publicNavigation.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={`site-nav__link${pathname === item.href ? ' site-nav__link--active' : ''}`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="site-header__actions site-header__actions--desktop">
          {user ? (
            <AuthenticatedEntryBar user={user} />
          ) : (
            <GuestEntryBar />
          )}
        </div>
        <div className="site-header__actions site-header__actions--mobile" ref={mobileMenuRef}>
          <Link
            href="/skills"
            aria-label="Search skills"
            className={`header-icon-button${pathname === '/skills' ? ' header-icon-button--active' : ''}`}
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <circle cx="11" cy="11" r="6.8" />
              <path d="M16.2 16.2 21 21" />
            </svg>
          </Link>
          <button
            type="button"
            className="mobile-nav-trigger"
            aria-label="Open navigation"
            aria-expanded={isMobileNavOpen}
            onClick={() => setIsMobileNavOpen((open) => !open)}
          >
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path d="M4 7h16" />
              <path d="M4 12h16" />
              <path d="M4 17h16" />
            </svg>
          </button>
          {isMobileNavOpen ? (
            <div className="mobile-nav-panel" role="menu" aria-label="Site navigation">
              <div className="mobile-nav-panel__label">Navigation</div>
              <div className="mobile-nav-panel__list">
                {publicNavigation.map((item) => (
                  <Link
                    key={item.href}
                    href={item.href}
                    role="menuitem"
                    onClick={() => setIsMobileNavOpen(false)}
                    className={`mobile-nav-panel__item${pathname === item.href ? ' mobile-nav-panel__item--active' : ''}`}
                  >
                    {item.label}
                  </Link>
                ))}
              </div>

              {user ? (
                <>
                  <div className="mobile-nav-panel__divider" />
                  <div className="mobile-nav-panel__label">Workspace</div>
                  <div className="mobile-nav-panel__list">
                    {workspaceNavigation.map((item) => (
                      <Link
                        key={item.href}
                        href={item.href}
                        role="menuitem"
                        onClick={() => setIsMobileNavOpen(false)}
                        className={`mobile-nav-panel__item${pathname === item.href ? ' mobile-nav-panel__item--active' : ''}`}
                      >
                        {item.label}
                      </Link>
                    ))}
                  </div>
                  <div className="mobile-nav-panel__divider" />
                  <form action={logoutAction}>
                    <button type="submit" className="mobile-nav-panel__item mobile-nav-panel__item--danger">
                      Log out
                    </button>
                  </form>
                </>
              ) : (
                <>
                  <div className="mobile-nav-panel__divider" />
                  <button
                    type="button"
                    className="mobile-nav-panel__item"
                    onClick={() => {
                      setIsMobileNavOpen(false);
                      setAuthMode('sign-in');
                      setIsAuthOpen(true);
                    }}
                  >
                    Sign in
                  </button>
                </>
              )}
            </div>
          ) : null}
        </div>
      </header>
      {showWorkspaceNav ? (
        <nav className="workspace-nav" aria-label="Account navigation">
          {workspaceNavigation.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={`workspace-nav__link${pathname === item.href ? ' workspace-nav__link--active' : ''}`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      ) : null}
      <main className="site-main">{children}</main>
      {!user ? (
        <AuthModal
          isOpen={isAuthOpen}
          mode={authMode}
          onClose={() => setIsAuthOpen(false)}
          onChangeMode={setAuthMode}
        />
      ) : null}
    </div>
  );
}
