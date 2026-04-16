import type { Metadata } from 'next';
import Link from 'next/link';
import { Card } from '@/shared/ui/card';
import { Button } from '@/shared/ui/button';

const directChannels = [
  {
    title: 'Email',
    value: 'sergioartg@gmail.com',
    href: 'mailto:sergioartg@gmail.com',
    note: 'Best for product, partnership, architecture, and delivery conversations.',
  },
  {
    title: 'Phone',
    value: '+1 347 210 2000',
    href: 'tel:+13472102000',
    note: 'Useful for urgent coordination or quick production decisions.',
  },
  {
    title: 'Location',
    value: 'Los Angeles, CA',
    href: 'https://www.google.com/maps/place/Los+Angeles,+CA',
    note: 'Remote-friendly, with a Pacific Time workflow.',
  },
] as const;

const socialProfiles = [
  {
    label: 'Instagram',
    handle: '@sergioartg',
    href: 'https://www.instagram.com/sergioartg/',
    description: 'Visual work, current projects, and creative direction.',
  },
  {
    label: 'GitHub',
    handle: 'SergeHall',
    href: 'https://github.com/SergeHall',
    description: 'Code, backend systems, platform work, and shipping history.',
  },
] as const;

const collaborationTopics = [
  'Product and platform architecture',
  'Next.js and Go delivery systems',
  'Skill marketplace strategy and operations',
  'Production hardening, cleanup, and ongoing improvements',
] as const;

export const metadata: Metadata = {
  title: 'Contact Lavoval',
  description:
    'Get in touch with Lavoval for product collaboration, platform architecture, delivery support, and direct contact.',
  alternates: {
    canonical: '/contact',
  },
  openGraph: {
    title: 'Contact Lavoval',
    description: 'Reach out for Lavoval product, engineering, and collaboration conversations.',
    url: '/contact',
  },
  twitter: {
    title: 'Contact Lavoval',
    description: 'Reach out for Lavoval product, engineering, and collaboration conversations.',
  },
};

export default function ContactPage() {
  return (
    <div className="stack stack--lg">
      <section className="hero">
        <div className="hero__copy">
          <span className="badge badge--warning">Get in touch</span>
          <h1>Let&apos;s talk about Lavoval, delivery, and what comes next.</h1>
          <p>
            If you want to collaborate on the platform, discuss product direction, or move through
            backend and frontend improvements together, this is the fastest path to reach me.
          </p>
          <div className="toolbar">
            <a href="mailto:sergioartg@gmail.com">
              <Button>Email now</Button>
            </a>
            <a href="tel:+13472102000">
              <Button variant="secondary">Call</Button>
            </a>
          </div>
        </div>

        <div className="hero__panel stack stack--md">
          <h2>Best fit for focused collaboration</h2>
          <p>
            I usually reply within 24 hours on weekdays. The smoothest conversations are clear,
            direct, and tied to a real deliverable, audit, rollout, or architecture decision.
          </p>
          <div className="metrics">
            <div>
              <div className="metric-value">24h</div>
              <div>Typical weekday response window</div>
            </div>
            <div>
              <div className="metric-value">LA</div>
              <div>Pacific Time coordination base</div>
            </div>
            <div>
              <div className="metric-value">Go + Next</div>
              <div>Primary stack for production work</div>
            </div>
          </div>
        </div>
      </section>

      <section className="grid">
        {directChannels.map((channel) => (
          <Card key={channel.title}>
            <div className="stack stack--sm">
              <div className="card__title">{channel.title}</div>
              <a href={channel.href} className="site-nav__link site-nav__link--active">
                {channel.value}
              </a>
              <p className="muted">{channel.note}</p>
            </div>
          </Card>
        ))}
      </section>

      <section className="grid">
        <Card>
          <div className="stack stack--md">
            <div className="card__title">What to reach out about</div>
            <div className="data-list">
              {collaborationTopics.map((topic) => (
                <div key={topic} className="data-list__item">
                  <strong>{topic}</strong>
                </div>
              ))}
            </div>
          </div>
        </Card>

        <Card>
          <div className="stack stack--md">
            <div className="card__title">Profiles</div>
            <div className="data-list">
              {socialProfiles.map((profile) => (
                <div key={profile.label} className="data-list__item">
                  <div className="stack stack--sm">
                    <strong>{profile.label}</strong>
                    <a href={profile.href} target="_blank" rel="noreferrer" className="muted">
                      {profile.handle}
                    </a>
                    <p className="muted">{profile.description}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </Card>

        <Card>
          <div className="stack stack--md">
            <div className="card__title">Quick path</div>
            <p className="muted">
              If the conversation is about Lavoval itself, send the goal, the current state, and the
              concrete next step you want to land. That makes it much easier to move fast.
            </p>
            <div className="toolbar">
              <a href="mailto:sergioartg@gmail.com?subject=Lavoval%20Collaboration">
                <Button>Start by email</Button>
              </a>
              <Link href="/skills">
                <Button variant="ghost">Explore skills</Button>
              </Link>
            </div>
          </div>
        </Card>
      </section>
    </div>
  );
}
