import type { Route } from 'next';
import type { SessionUser } from '@lavoval/contracts';

export type NavigationItem = {
  href: Route;
  label: string;
};

export const publicNavigation: NavigationItem[] = [
  { href: '/', label: 'Overview' },
  { href: '/skills', label: 'Skills' },
];

export function accountNavigation(user: SessionUser): NavigationItem[] {
  const items: NavigationItem[] = [
    { href: '/account', label: 'Dashboard' },
    { href: '/skills', label: 'Skills' },
    { href: '/account/my-skills', label: 'My Skills' },
    { href: '/account/profile', label: 'Profile' },
  ];

  if (user.role === 'admin') {
    items.push({ href: '/admin', label: 'Admin' });
  }

  return items;
}
