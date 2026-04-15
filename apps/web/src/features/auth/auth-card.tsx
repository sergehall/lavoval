import type { PropsWithChildren } from 'react';
import { Card } from '@/shared/ui/card';

export function AuthCard({
  children,
  title,
  description,
  compact = false,
}: PropsWithChildren<{ title: string; description: string; compact?: boolean }>) {
  return (
    <Card className={compact ? 'card--compact' : undefined}>
      <div className="stack stack--sm">
        <h1>{title}</h1>
        <p className="muted">{description}</p>
        {children}
      </div>
    </Card>
  );
}
