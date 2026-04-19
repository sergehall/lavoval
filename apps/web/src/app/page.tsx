import type { Metadata } from 'next';
import Link from 'next/link';
import { Card } from '@/shared/ui/card';
import { Button } from '@/shared/ui/button';
import { signInHref, signUpHref } from '@/shared/lib/auth-navigation';
import { loadPublicSkills } from '@/shared/lib/public-skill-loader';

const publicHighlights = [
  'Skill-first marketplace structure',
  'Typed session and API boundaries',
  'Authorship, modules, and governance',
];

export const metadata: Metadata = {
  title: 'Human Skill Marketplace For AI-Era Expertise',
  description:
    'Discover Lavoval, a marketplace for exchanging real human expertise in the age of AI. Publish skill offers, explore trusted specialists, and turn practical knowledge into reusable modules and execution-ready workflows.',
  alternates: {
    canonical: '/',
  },
  openGraph: {
    title: 'Lavoval | Human Skill Marketplace For AI-Era Expertise',
    description:
      'Publish expertise, discover trusted people, and package practical know-how into reusable skill offers for an AI-shaped world.',
    url: '/',
  },
  twitter: {
    title: 'Lavoval | Human Skill Marketplace For AI-Era Expertise',
    description:
      'Publish expertise, discover trusted people, and package practical know-how into reusable skill offers for an AI-shaped world.',
  },
};

export default async function LandingPage() {
  const latestSkills = (await loadPublicSkills()).slice(0, 3);

  return (
    <div className="stack stack--landing">
      <section className="hero">
        <div className="hero__copy">
          <span className="badge badge--warning">Starter kit</span>
          <h1>Exchange real human skills in the age of AI.</h1>
          <p>
            Publish expertise, package practical know-how into reusable skill offers, and let people
            or AI agents discover the right operator faster.
          </p>
          <div className="toolbar">
            <Link href={signUpHref}>
              <Button>Create workspace</Button>
            </Link>
            <Link href={signInHref}>
              <Button variant="secondary">Sign in</Button>
            </Link>
          </div>
        </div>
        <div className="hero__panel stack stack--md">
          <h2>Built for skill-led products</h2>
          <p>
            Clean Go services, typed contracts, and a Next.js UI ready for a serious expertise
            marketplace.
          </p>
          <div className="metrics">
            <div>
              <div className="metric-value">3</div>
              <div>Core surfaces</div>
            </div>
            <div>
              <div className="metric-value">P2P</div>
              <div>Exchange model</div>
            </div>
            <div>
              <div className="metric-value">AI Era</div>
              <div>Human expertise, reusable</div>
            </div>
          </div>
        </div>
      </section>

      <section className="stack stack--md">
        <div className="section-heading">
          <div className="stack stack--sm">
            <h2>Latest skill offers</h2>
            <p className="muted">
              Fresh public offers from the marketplace, ready to explore and run.
            </p>
          </div>
          <Link href="/skills">
            <Button variant="secondary">View all skills</Button>
          </Link>
        </div>

        {latestSkills.length > 0 ? (
          <section className="skill-cover-grid">
            {latestSkills.map((skill) => (
              <article key={skill.id} className="skill-cover-card">
                <div className="skill-cover-card__topline">
                  <span className="skill-cover-card__provider">{skill.provider}</span>
                  {(skill as typeof skill & { difficulty?: string }).difficulty ? (
                    <span className="skill-cover-card__meta">
                      {(skill as typeof skill & { difficulty?: string }).difficulty}
                    </span>
                  ) : null}
                </div>
                <div className="stack stack--sm">
                  <h3>{skill.title}</h3>
                  <p>{skill.summary}</p>
                </div>
                <div className="skill-cover-card__footer">
                  <Link href={`/authors/${skill.creator.id}`} className="muted">
                    By {skill.creator.firstName} {skill.creator.lastName}
                  </Link>
                  <Link href={`/skills/${skill.id}`} className="skill-cover-card__link">
                    Open skill
                  </Link>
                </div>
              </article>
            ))}
          </section>
        ) : (
          <Card>
            <div className="stack stack--sm">
              <h3>No public skills yet</h3>
              <p className="muted">
                Publish the first skill offers to turn the landing page into a live marketplace
                preview.
              </p>
            </div>
          </Card>
        )}
      </section>

      <section className="grid grid--compact">
        {publicHighlights.map((highlight) => (
          <Card key={highlight}>
            <div className="stack stack--sm">
              <h3>{highlight}</h3>
              <p className="muted">
                Structured to support discovery, trust, and product growth without reworking the
                foundation later.
              </p>
            </div>
          </Card>
        ))}
      </section>
    </div>
  );
}
