'use client';

import Link from 'next/link';
import type { PropsWithChildren } from 'react';
import { usePathname } from 'next/navigation';
import type { SessionUser } from '@lavoval/contracts';
import { accountNavigation, publicNavigation } from '@/shared/lib/navigation';
import { Button } from '@/shared/ui/button';
import { logoutAction } from '@/features/auth/actions';

export function AppShell({ children, user }: PropsWithChildren<{ user?: SessionUser }>) {
  const pathname = usePathname();
  const navigation = user ? accountNavigation(user) : publicNavigation;

  return (
    <div className="app-frame">
      <header className="site-header">
        <Link href="/" className="brand-mark">
          <span className="brand-mark__label">Lavoval</span>
          <span className="brand-mark__caption">Skill operations platform</span>
        </Link>
        <nav className="site-nav site-nav--primary">
          {navigation.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className={`site-nav__link${pathname === item.href ? ' site-nav__link--active' : ''}`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="site-header__actions">
          {user ? (
            <form action={logoutAction}>
              <Button variant="ghost" type="submit">
                Sign out
              </Button>
            </form>
          ) : (
            <div className="site-auth-actions">
              <Link href="/login" className="site-nav__link site-nav__link--subtle">
                Sign in
              </Link>
              <Link href="/register" className="site-nav__link site-nav__link--cta">
                Create account
              </Link>
            </div>
          )}
        </div>
      </header>
      <main className="site-main">{children}</main>
    </div>
  );
}
