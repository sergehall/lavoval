import type { Route } from 'next';
import type { SessionUser } from '@lavoval/contracts';
import { canAccessAdmin } from '@/shared/lib/rbac';

export type NavigationItem = {
  href: Route;
  label: string;
};

export const publicNavigation: NavigationItem[] = [
  { href: '/', label: 'Marketplace' },
  { href: '/skills', label: 'Explore Skills' },
  { href: '/contact' as Route, label: 'Contact' },
];

export function cabinetNavigation(user: SessionUser): NavigationItem[] {
  const items: NavigationItem[] = [
    { href: '/account', label: 'Home' },
    { href: '/account/my-skills', label: 'My Offers' },
    { href: '/account/runs', label: 'Runs' },
    { href: '/account/profile', label: 'Identity' },
    { href: '/account/security', label: 'Security' },
  ];

  if (canAccessAdmin(user.role)) {
    items.push({ href: '/admin', label: 'Governance' });
  }

  return items;
}
