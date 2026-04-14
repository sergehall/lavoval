import type { Route } from 'next';
import type { SessionUser } from '@lavoval/contracts';

export type NavigationItem = {
  href: Route;
  label: string;
};

export const publicNavigation: NavigationItem[] = [
  { href: '/', label: 'Marketplace' },
  { href: '/skills', label: 'Explore Skills' },
];

export function accountNavigation(user: SessionUser): NavigationItem[] {
  const items: NavigationItem[] = [
    { href: '/account', label: 'Home' },
    { href: '/skills', label: 'Explore Skills' },
    { href: '/account/my-skills', label: 'My Offers' },
    { href: '/account/profile', label: 'Identity' },
  ];

  if (user.role === 'admin') {
    items.push({ href: '/admin', label: 'Governance' });
  }

  return items;
}
