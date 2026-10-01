'use client';

import { Card } from '@/shared/ui/card';

export default function Error({ reset }: { reset: () => void }) {
  return (
    <Card>
      <h1>Agent Board unavailable</h1>
      <p>Could not load the board right now.</p>
      <button onClick={reset}>Try again</button>
    </Card>
  );
}
