'use client';

import type { Route } from 'next';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

const adminNavItems = [
  { href: '/admin', label: 'Dashboard' },
  { href: '/admin/users', label: 'Users' },
  { href: '/admin/skills', label: 'Skills' },
  { href: '/admin/runs', label: 'Runs' },
  { href: '/admin/enrollments', label: 'Enrollments' },
  { href: '/admin/mail' as Route, label: 'Mail' },
] as const satisfies ReadonlyArray<{ href: Route; label: string }>;

export function AdminNav() {
  const pathname = usePathname();

  return (
    <nav className="admin-nav" aria-label="Admin navigation">
      <span className="admin-nav__badge">Governance</span>
      {adminNavItems.map(({ href, label }) => {
        // Exact match for dashboard, prefix match for sub-pages
        const isActive = href === '/admin' ? pathname === '/admin' : pathname.startsWith(href);

        return (
          <Link
            key={href}
            href={href}
            className={`admin-nav__link${isActive ? ' admin-nav__link--active' : ''}`}
          >
            {label}
          </Link>
        );
      })}
    </nav>
  );
}
