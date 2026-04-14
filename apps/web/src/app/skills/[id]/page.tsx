import Link from 'next/link';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { ApiError, fetchSkillById, getSession } from '@/shared/api/server-client';
import { formatDate } from '@/shared/lib/utils';
import { SkillMarkdown } from '@/features/skills/skill-markdown';
import { RunSkillForm } from '@/features/runtime/run-skill-form';
import { runSkillAction } from '@/features/runtime/actions';
import type { SkillDetail } from '@lavoval/registry';

export default async function PublicSkillDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  const session = await getSession();

  // Data fetching is isolated inside try/catch — no JSX here.
  // The rule react-hooks/error-boundaries disallows returning JSX from
  // inside a try/catch because React cannot catch rendering errors that way.
  let skill: SkillDetail | null = null;
  let notFound = false;

  try {
    const { data } = await fetchSkillById(id);
    skill = data;
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      notFound = true;
    } else {
      throw error;
    }
  }

  if (notFound || !skill) {
    return (
      <Card>
        <div className="empty-state stack stack--md">
          <Badge tone="warning">Not publicly available</Badge>
          <h1>This skill offer is private or not published yet.</h1>
          <p className="muted">
            Public marketplace pages only open records that are both <strong>published</strong>{' '}
            and <strong>public</strong>. If you are the author, open this item from{' '}
            <strong>My Offers</strong> and change its visibility or status.
          </p>
          <div className="toolbar">
            <Link href="/skills" className="site-nav__link site-nav__link--subtle">
              Back to marketplace
            </Link>
            <Link href="/account/my-skills" className="site-nav__link site-nav__link--cta">
              Open My Offers
            </Link>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <div className="stack stack--lg">
      <section className="section-heading">
        <div className="stack stack--sm">
          <h1>{skill.title}</h1>
          <p>{skill.summary}</p>
          <div className="inline-actions muted">
            <span>
              Offered by {skill.creator.firstName} {skill.creator.lastName} ({skill.creator.email})
            </span>
            <span>Updated {formatDate(skill.updatedAt)}</span>
          </div>
        </div>
        <Badge tone={skill.status === 'published' ? 'success' : 'warning'}>{skill.status}</Badge>
      </section>
      <Card>
        <div className="stack stack--md">
          <h2>What this person is offering</h2>
          <SkillMarkdown content={skill.description} />
        </div>
      </Card>
      <Card>
        <div className="stack stack--md">
          <div className="section-heading">
            <h2>Exchange structure</h2>
            <span className="muted">{skill.modules.length} items</span>
          </div>
          <div className="data-list">
            {skill.modules.map((module) => (
              <article key={module.id} className="data-list__item">
                <div className="inline-actions">
                  <Badge>{String(module.position + 1).padStart(2, '0')}</Badge>
                  <h3>{module.title}</h3>
                </div>
                <p>{module.summary}</p>
                <SkillMarkdown content={module.content} />
              </article>
            ))}
          </div>
        </div>
      </Card>
      <Card>
        {session ? (
          <RunSkillForm action={runSkillAction.bind(null, skill.id)} entrypoint={skill.entrypoint} />
        ) : (
          <div className="stack stack--md">
            <h2>Sign in to run this skill</h2>
            <p className="muted">
              Runtime execution is available for signed-in users. Join the marketplace to run this
              skill and keep a history of your results.
            </p>
            <div className="toolbar">
              <Link href="/login" className="site-nav__link site-nav__link--cta">
                Sign in
              </Link>
              <Link href="/register" className="site-nav__link site-nav__link--subtle">
                Create account
              </Link>
            </div>
          </div>
        )}
      </Card>
    </div>
  );
}
