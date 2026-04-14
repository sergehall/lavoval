import Link from 'next/link';
import { Card } from '@/shared/ui/card';
import { Button } from '@/shared/ui/button';

const publicHighlights = [
  'Marketplace structure for discovering and exchanging human skills',
  'Typed API contracts and role-aware session boundaries',
  'Scaffold for authorship, modules, progress, and governance',
];

export default function LandingPage() {
  return (
    <div className="stack stack--lg">
      <section className="hero">
        <div className="hero__copy">
          <span className="badge badge--warning">Production starter kit</span>
          <h1>Exchange real human skills in the age of AI.</h1>
          <p>
            Lavoval is a growth-ready foundation for a platform where users publish expertise,
            discover other people's strengths, and turn practical knowledge into exchangeable
            skill pages and modules.
          </p>
          <div className="toolbar">
            <Link href="/register">
              <Button>Create workspace</Button>
            </Link>
            <Link href="/login">
              <Button variant="secondary">Sign in</Button>
            </Link>
          </div>
        </div>
        <div className="hero__panel stack stack--md">
          <h2>Designed for a human skill marketplace</h2>
          <p>
            Clear package boundaries, clean Go services, PostgreSQL-backed workflows, and a Next.js
            App Router UI that can grow into a serious exchange platform for AI-era expertise.
          </p>
          <div className="metrics">
            <div>
              <div className="metric-value">3</div>
              <div>Application zones</div>
            </div>
            <div>
              <div className="metric-value">P2P</div>
              <div>User-to-user skill exchange direction</div>
            </div>
            <div>
              <div className="metric-value">AI Era</div>
              <div>Human expertise packaged into reusable modules</div>
            </div>
          </div>
        </div>
      </section>

      <section className="grid">
        {publicHighlights.map((highlight) => (
          <Card key={highlight}>
            <div className="stack stack--sm">
              <h3>{highlight}</h3>
              <p className="muted">
                Structured to support matching, reputation, billing, permissions, analytics,
                notifications, and audit-driven workflows without re-architecting the foundation.
              </p>
            </div>
          </Card>
        ))}
      </section>
    </div>
  );
}
