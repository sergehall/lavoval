import type { Role } from '@lavoval/contracts';

export function canAccessAdmin(role: Role | null | undefined) {
  return role === 'admin' || role === 'root_owner';
}

export function isRootOwner(role: Role | null | undefined) {
  return role === 'root_owner';
}

export function roleBadgeLabel(role: Role) {
  switch (role) {
    case 'root_owner':
      return 'ROOT OWNER';
    case 'admin':
      return 'ADMIN';
    default:
      return 'USER';
  }
}
