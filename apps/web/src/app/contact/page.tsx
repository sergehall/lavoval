import type { Metadata } from 'next';
import { Card } from '@/shared/ui/card';
import { Button } from '@/shared/ui/button';

const directChannels = [
  {
    title: 'Email',
    value: 'serge.hall.dev@gmail.com',
    href: 'mailto:serge.hall.dev@gmail.com',
    note: 'Best for product, partnership, architecture, and delivery conversations.',
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
  'Shape the product and platform architecture before complexity starts compounding.',
  'Turn Next.js and Go delivery into a cleaner, faster, more reliable production system.',
  'Sharpen marketplace strategy, governance, and operational flow before scale creates drag.',
  'Design and package human skills for the AI era so expertise becomes clearer, more usable, and easier to exchange.',
  'Push through hardening, cleanup, and the kind of improvements that make the product feel ready.',
  'Translate ambitious ideas into a roadmap the team can actually ship with confidence.',
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
    <div className="stack stack--lg contact-page">
      <section className="hero">
        <div className="hero__copy">
          <span className="badge badge--warning">Get in touch</span>
          <h1 style={{ fontSize: 'clamp(1.9rem, 4vw, 3.1rem)', marginBottom: 12 }}>
            Let&apos;s talk about Lavoval, delivery, and what comes next.
          </h1>
          <p>
            If you want to collaborate on the platform, discuss product direction, or move through
            backend and frontend improvements together, this is the fastest path to reach me.
          </p>
          <div className="toolbar">
            <a href="mailto:serge.hall.dev@gmail.com">
              <Button>Email now</Button>
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

      <section className="contact-page__grid">
        <Card>
          <div className="stack stack--sm">
            <div className="card__title">Email and profiles</div>
            <div className="data-list contact-page__list">
              {directChannels.map((channel) => (
                <div
                  key={channel.title}
                  className="data-list__item contact-page__item contact-page__item--interactive"
                >
                  <strong>{channel.title}</strong>
                  <a href={channel.href} className="muted">
                    {channel.value}
                  </a>
                  <p className="muted">{channel.note}</p>
                </div>
              ))}
              {socialProfiles.map((profile) => (
                <div
                  key={profile.label}
                  className="data-list__item contact-page__item contact-page__item--interactive"
                >
                  <div className="stack stack--sm contact-page__item-stack">
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
          <div className="stack stack--sm">
            <div className="card__title">What to reach out about</div>
            <div className="data-list contact-page__list">
              {collaborationTopics.map((topic) => (
                <div key={topic} className="data-list__item contact-page__item">
                  <strong>{topic}</strong>
                </div>
              ))}
            </div>
          </div>
        </Card>
      </section>
    </div>
  );
}
