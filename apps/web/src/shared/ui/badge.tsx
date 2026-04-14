import clsx from 'clsx';
import type { PropsWithChildren } from 'react';

export function Badge({
  children,
  tone = 'neutral',
}: PropsWithChildren<{ tone?: 'neutral' | 'success' | 'warning' }>) {
  return <span className={clsx('badge', `badge--${tone}`)}>{children}</span>;
}
