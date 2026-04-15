import type { PropsWithChildren } from 'react';

export function Card({ children, className }: PropsWithChildren<{ className?: string }>) {
  return <section className={className ? `card ${className}` : 'card'}>{children}</section>;
}
