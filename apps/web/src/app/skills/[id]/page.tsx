import type { Metadata } from 'next';
import Link from 'next/link';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import {
  ApiError,
  fetchSkillById,
  fetchSkillReviews,
  fetchRecommendedAgents,
  getValidatedSession,
} from '@/shared/api/server-client';
import { signInHref, signUpHref } from '@/shared/lib/auth-navigation';
import { formatDate } from '@/shared/lib/utils';
import { SkillMarkdown } from '@/features/skills/skill-markdown';
import { RunSkillForm } from '@/features/runtime/run-skill-form';
import { runSkillAction } from '@/features/runtime/actions';
import type { SkillDetail } from '@lavoval/registry';
import type { SkillReview, SkillAgentCompatibility } from '@lavoval/contracts';

export async function generateMetadata({
  params,
}: {
  params: Promise<{ id: string }>;
}): Promise<Metadata> {
  const { id } = await params;
  try {
    const { data: skill } = await fetchSkillById(id);
    if (skill.status !== 'published' || skill.visibility !== 'public') {
      return { title: 'Skill Unavailable', robots: { index: false, follow: false } };
    }
    const description = skill.summary || `Explore ${skill.title} on Lavoval.`;
    return {
      title: `${skill.title} by ${skill.creator.firstName} ${skill.creator.lastName}`,
      description,
      alternates: { canonical: `/skills/${skill.id}` },
      openGraph: { title: `${skill.title} | Lavoval`, description, url: `/skills/${skill.id}` },
      twitter: { title: `${skill.title} | Lavoval`, description },
    };
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return { title: 'Skill Not Found', robots: { index: false, follow: false } };
    }
    throw error;
  }
}

function StarRating({ rating }: { rating: number }) {
  const filled = Math.round(rating);
  return (
    <span aria-label={`${rating.toFixed(1)} out of 5`}>
      {'★'.repeat(filled)}
      {'☆'.repeat(5 - filled)} {rating.toFixed(1)}
    </span>
  );
}

export default async function PublicSkillDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams?: Promise<{ runError?: string; runMessage?: string }>;
}) {
  const { id } = await params;
  const resolvedSearchParams = searchParams ? await searchParams : undefined;
  const session = await getValidatedSession();
  const runError = resolvedSearchParams?.runError;
  const runMessage = resolvedSearchParams?.runMessage;

  let skill: SkillDetail | null = null;
  let reviews: SkillReview[] = [];
  let recommended: SkillAgentCompatibility[] = [];
  let notFound = false;

  try {
    const [skillResult, reviewsResult, recommendedResult] = await Promise.allSettled([
      fetchSkillById(id),
      fetchSkillReviews(id),
      fetchRecommendedAgents(id),
    ]);

    if (skillResult.status === 'fulfilled') {
      skill = skillResult.value?.data ?? null;
    } else if ((skillResult.reason as ApiError)?.status === 404) {
      notFound = true;
    } else {
      throw skillResult.reason;
    }

    if (reviewsResult.status === 'fulfilled') {
      reviews = reviewsResult.value ?? [];
    }
    if (recommendedResult.status === 'fulfilled') {
      recommended = recommendedResult.value ?? [];
    }
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
          <h1>This skill is private or not published yet.</h1>
          <p className="muted">
            Public pages only show skills that are both <strong>published</strong> and{' '}
            <strong>public</strong>. If you are the author, open this item from My Skills.
          </p>
          <div className="toolbar">
            <Link href="/skills" className="site-nav__link site-nav__link--subtle">
              Back to Skills
            </Link>
            <Link href="/account/my-skills" className="site-nav__link site-nav__link--cta">
              My Skills
            </Link>
          </div>
        </div>
      </Card>
    );
  }

  const extendedSkill = skill as SkillDetail & {
    difficulty?: string;
    skillType?: string;
    isAgentReady?: boolean;
    avgRating?: number;
    runsCount?: number;
    savesCount?: number;
    estimatedTimeMinutes?: number;
    tags?: Array<{ id: string; name: string }>;
  };

  return (
    <div className="stack stack--lg">
      {/* Header */}
      <section className="section-heading">
        <div className="stack stack--sm">
          <div className="inline-actions">
            <h1 style={{ margin: 0 }}>{skill.title}</h1>
            {extendedSkill.isAgentReady && <Badge tone="neutral">Agent-ready</Badge>}
          </div>
          <p>{skill.summary}</p>
          <div className="inline-actions muted" style={{ flexWrap: 'wrap', gap: '8px' }}>
            <Link href={`/authors/${skill.creator.id}`}>
              By {skill.creator.firstName} {skill.creator.lastName}
            </Link>
            {extendedSkill.difficulty && <Badge>{extendedSkill.difficulty}</Badge>}
            {extendedSkill.skillType && <Badge tone="neutral">{extendedSkill.skillType}</Badge>}
            {extendedSkill.avgRating != null && extendedSkill.avgRating > 0 && (
              <StarRating rating={extendedSkill.avgRating} />
            )}
            {extendedSkill.runsCount != null && extendedSkill.runsCount > 0 && (
              <span>{extendedSkill.runsCount} runs</span>
            )}
            {extendedSkill.estimatedTimeMinutes && (
              <span>~{extendedSkill.estimatedTimeMinutes} min</span>
            )}
            <span>Updated {formatDate(skill.updatedAt)}</span>
          </div>
          {extendedSkill.tags && extendedSkill.tags.length > 0 && (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '4px' }}>
              {extendedSkill.tags.map((tag) => (
                <Link
                  key={tag.id}
                  href={`/skills?tags=${tag.name}`}
                  style={{
                    padding: '1px 8px',
                    borderRadius: '12px',
                    border: '1px solid var(--border)',
                    fontSize: '0.8rem',
                    textDecoration: 'none',
                    color: 'inherit',
                  }}
                >
                  {tag.name}
                </Link>
              ))}
            </div>
          )}
        </div>
        <Badge tone="success">{skill.status}</Badge>
      </section>

      {/* Description */}
      <Card>
        <details className="skill-contract">
          <summary className="skill-contract__summary">
            <div className="section-heading">
              <h2>What this skill covers</h2>
              <Badge tone="neutral">Overview</Badge>
            </div>
          </summary>
          <div className="stack stack--md skill-contract__body">
            <SkillMarkdown content={skill.description} />
          </div>
        </details>
      </Card>

      {skill.currentVersion ? (
        <Card>
          <details className="skill-contract" open={false}>
            <summary className="skill-contract__summary">
              <div className="section-heading">
                <h2>AI contract</h2>
                <Badge tone="neutral">v{skill.currentVersion.versionNo}</Badge>
              </div>
            </summary>
            <div className="stack stack--md skill-contract__body">
              <p className="muted">
                Machine-readable prompt and schema snapshot for agent-oriented usage.
              </p>
              {skill.currentVersion.systemInstructions ? (
                <div className="stack stack--sm">
                  <strong>System instructions</strong>
                  <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                    {skill.currentVersion.systemInstructions}
                  </pre>
                </div>
              ) : null}
              {skill.currentVersion.promptTemplate ? (
                <div className="stack stack--sm">
                  <strong>Prompt template</strong>
                  <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                    {skill.currentVersion.promptTemplate}
                  </pre>
                </div>
              ) : null}
              {skill.currentVersion.inputSchema ? (
                <div className="stack stack--sm">
                  <strong>Input schema</strong>
                  <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                    {JSON.stringify(skill.currentVersion.inputSchema, null, 2)}
                  </pre>
                </div>
              ) : null}
              {skill.currentVersion.outputSchema ? (
                <div className="stack stack--sm">
                  <strong>Output schema</strong>
                  <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                    {JSON.stringify(skill.currentVersion.outputSchema, null, 2)}
                  </pre>
                </div>
              ) : null}
              {skill.currentVersion.errorSchema ? (
                <div className="stack stack--sm">
                  <strong>Error schema</strong>
                  <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                    {JSON.stringify(skill.currentVersion.errorSchema, null, 2)}
                  </pre>
                </div>
              ) : null}
            </div>
          </details>
        </Card>
      ) : null}

      {/* Modules */}
      {skill.modules.length > 0 && (
        <Card>
          <details className="skill-contract">
            <summary className="skill-contract__summary">
              <div className="section-heading">
                <h2>Steps & content</h2>
                <Badge tone="neutral">{skill.modules.length} items</Badge>
              </div>
            </summary>
            <div className="stack stack--md skill-contract__body">
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
          </details>
        </Card>
      )}

      {/* Recommended agents */}
      {recommended.length > 0 && (
        <Card>
          <div className="stack stack--md">
            <h2>Recommended agents</h2>
            <div className="data-list">
              {recommended.map((compat) => (
                <article key={compat.agentId} className="data-list__item">
                  <div className="section-heading">
                    <div className="inline-actions">
                      <strong>{compat.agent.name}</strong>
                      <Badge tone="neutral">{compat.agent.provider}</Badge>
                    </div>
                    <span className="muted" style={{ fontSize: '0.85rem' }}>
                      Score: {compat.compatibilityScore.toFixed(0)}%
                    </span>
                  </div>
                  <span className="muted" style={{ fontSize: '0.8rem' }}>
                    {compat.agent.modelName}
                  </span>
                </article>
              ))}
            </div>
          </div>
        </Card>
      )}

      {/* Run */}
      <Card>
        {session ? (
          <div className="stack stack--md">
            {runError ? (
              <div className="empty-state stack stack--sm" style={{ padding: '1rem' }}>
                <Badge tone="warning">Run failed</Badge>
                <p className="muted" style={{ margin: 0 }}>
                  {runMessage || 'We could not execute this skill right now. Please try again.'}
                </p>
              </div>
            ) : null}
            <RunSkillForm
              action={runSkillAction.bind(null, skill.id)}
              entrypoint={skill.entrypoint}
            />
          </div>
        ) : (
          <div className="stack stack--md">
            <h2>Sign in to run this skill</h2>
            <p className="muted">
              Skill execution is available for signed-in users. Run this skill and keep a history of
              your results.
            </p>
            <div className="toolbar">
              <Link href={signInHref} className="site-nav__link site-nav__link--cta">
                Sign in
              </Link>
              <Link href={signUpHref} className="site-nav__link site-nav__link--subtle">
                Create account
              </Link>
            </div>
          </div>
        )}
      </Card>

      {/* Reviews */}
      <Card>
        <div className="stack stack--md">
          <div className="section-heading">
            <h2>Reviews</h2>
            <span className="muted">
              {reviews.length} review{reviews.length !== 1 ? 's' : ''}
            </span>
          </div>
          {reviews.length > 0 ? (
            <div className="data-list">
              {reviews.map((rv) => (
                <article key={rv.id} className="data-list__item">
                  <div className="section-heading">
                    <div className="inline-actions">
                      <strong>
                        {rv.reviewer.firstName} {rv.reviewer.lastName}
                      </strong>
                      <StarRating rating={rv.rating} />
                    </div>
                    <span className="muted" style={{ fontSize: '0.8rem' }}>
                      {formatDate(rv.createdAt)}
                    </span>
                  </div>
                  {rv.reviewText && <p>{rv.reviewText}</p>}
                </article>
              ))}
            </div>
          ) : (
            <p className="muted">No reviews yet. Run this skill and share your experience.</p>
          )}
        </div>
      </Card>
    </div>
  );
}
