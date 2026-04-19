'use client';

import Link from 'next/link';
import { useRouter, usePathname } from 'next/navigation';
import { useCallback } from 'react';
import type { SkillSummary } from '@lavoval/registry';
import type { Category, Tag, SkillFilterParams } from '@lavoval/contracts';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { formatDate } from '@/shared/lib/utils';

const SORT_OPTIONS = [
  { value: '', label: 'Newest first' },
  { value: 'popular', label: 'Most popular' },
  { value: 'rating', label: 'Top rated' },
  { value: 'runs', label: 'Most runs' },
  { value: 'new', label: 'Recently published' },
];

const DIFFICULTY_OPTIONS = [
  { value: '', label: 'All levels' },
  { value: 'junior', label: 'Junior' },
  { value: 'middle', label: 'Middle' },
  { value: 'senior', label: 'Senior' },
];

const SKILL_TYPE_OPTIONS = [
  { value: '', label: 'All types' },
  { value: 'guide', label: 'Guide' },
  { value: 'workflow', label: 'Workflow' },
  { value: 'prompt-pack', label: 'Prompt pack' },
  { value: 'agent-ready', label: 'Agent-ready' },
  { value: 'service', label: 'Service' },
];

function StarRating({ rating }: { rating: number }) {
  const filled = Math.round(rating);
  return (
    <span aria-label={`${rating.toFixed(1)} out of 5`} title={`${rating.toFixed(1)} stars`}>
      {'★'.repeat(filled)}
      {'☆'.repeat(5 - filled)}
    </span>
  );
}

export function SkillsCatalog({
  skills,
  categories,
  tags,
  activeFilter,
}: {
  skills: SkillSummary[];
  categories: Category[];
  tags: Tag[];
  activeFilter: SkillFilterParams & { q?: string };
}) {
  const router = useRouter();
  const pathname = usePathname();

  const updateFilter = useCallback(
    (updates: Record<string, string>) => {
      const params = new URLSearchParams();
      const merged = { ...activeFilter, ...updates };
      for (const [k, v] of Object.entries(merged)) {
        if (v && v !== '') params.set(k, String(v));
      }

      (router.push as any)(`${pathname}?${params.toString()}`);
    },
    [router, pathname, activeFilter],
  );

  const clearFilter = (key: string) => {
    const params = new URLSearchParams();
    for (const [k, v] of Object.entries(activeFilter)) {
      if (k !== key && v && v !== '') params.set(k, String(v));
    }

    (router.push as any)(`${pathname}?${params.toString()}`);
  };

  const hasFilters = Object.values(activeFilter).some((v) => v && v !== '');

  return (
    <div className="skills-layout">
      <div className="skills-sidebar stack stack--md">
        <Card className="skills-filters-card">
          <div className="skills-filters">
            <div className="section-heading skills-filters__header">
              <div>
                <h3 style={{ margin: 0 }}>Filters</h3>
                <p className="muted" style={{ fontSize: '0.78rem', marginTop: '2px' }}>
                  {skills.length} skill{skills.length !== 1 ? 's' : ''}
                  {activeFilter.q && <> matching &quot;{activeFilter.q}&quot;</>}
                </p>
              </div>
              {hasFilters && (
                <button
                  className="muted"
                  style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    fontSize: '0.82rem',
                    flexShrink: 0,
                  }}
                  onClick={() => (router.push as any)(pathname)}
                >
                  Clear all
                </button>
              )}
            </div>

            <div className="filter-group">
              <span className="filter-label">Sort by</span>
              <select
                className="input"
                value={activeFilter.sort ?? ''}
                onChange={(e) => updateFilter({ sort: e.target.value })}
              >
                {SORT_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>

            {categories.length > 0 && (
              <div className="filter-group">
                <span className="filter-label">Category</span>
                <select
                  className="input"
                  value={activeFilter.category ?? ''}
                  onChange={(e) => updateFilter({ category: e.target.value, subcategory: '' })}
                >
                  <option value="">All categories</option>
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <div className="filter-group">
              <span className="filter-label">Difficulty</span>
              <select
                className="input"
                value={activeFilter.difficulty ?? ''}
                onChange={(e) => updateFilter({ difficulty: e.target.value })}
              >
                {DIFFICULTY_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>

            <div className="filter-group">
              <span className="filter-label">Type</span>
              <select
                className="input"
                value={activeFilter.skillType ?? ''}
                onChange={(e) => updateFilter({ skillType: e.target.value })}
              >
                {SKILL_TYPE_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>

            <label className="filter-checkbox-row skills-filters__toggle">
              <input
                type="checkbox"
                checked={
                  activeFilter.agentReady === true ||
                  activeFilter.agentReady === ('true' as unknown as boolean)
                }
                onChange={(e) => updateFilter({ agentReady: e.target.checked ? 'true' : '' })}
              />
              Agent-ready only
            </label>

            {tags.length > 0 && (
              <div className="filter-group skills-filters__tags">
                <span className="filter-label">Tags</span>
                <div className="filter-tag-list">
                  {tags.slice(0, 20).map((tag) => {
                    const activeTags = (activeFilter.tags ?? '').split(',').filter(Boolean);
                    const isActive = activeTags.includes(tag.slug);
                    return (
                      <button
                        key={tag.id}
                        className={`filter-tag${isActive ? ' filter-tag--active' : ''}`}
                        onClick={() => {
                          const next = isActive
                            ? activeTags.filter((t) => t !== tag.slug)
                            : [...activeTags, tag.slug];
                          updateFilter({ tags: next.join(',') });
                        }}
                      >
                        {tag.name}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        </Card>
      </div>

      <div className="skills-main">
        <Card className="skills-results-card">
          <div className="data-list">
            {skills.length > 0 ? (
              skills.map((skill) => (
                <article key={skill.id} className="skill-catalog-card">
                  <div className="skill-catalog-card__eyebrow">
                    <span className="skill-catalog-card__provider">{skill.provider}</span>
                    {(skill as SkillSummary & { skillType?: string }).skillType ? (
                      <span className="skill-catalog-card__type">
                        {(skill as SkillSummary & { skillType?: string }).skillType}
                      </span>
                    ) : null}
                  </div>

                  <div className="stack stack--sm">
                    <h2 style={{ margin: 0 }}>{skill.title}</h2>
                    <p>{skill.summary}</p>
                  </div>

                  {((
                    skill as SkillSummary & {
                      tags?: Array<{ id: string; name: string; slug: string }>;
                    }
                  ).tags?.length ?? 0) > 0 && (
                    <div className="skill-catalog-card__tags">
                      {(
                        (
                          skill as SkillSummary & {
                            tags?: Array<{ id: string; name: string; slug: string }>;
                          }
                        ).tags ?? []
                      )
                        .slice(0, 3)
                        .map((tag) => (
                          <button
                            key={tag.id}
                            className="skill-catalog-card__tag"
                            onClick={() => {
                              const activeTags = (activeFilter.tags ?? '')
                                .split(',')
                                .filter(Boolean);
                              if (!activeTags.includes(tag.slug)) {
                                updateFilter({ tags: [...activeTags, tag.slug].join(',') });
                              }
                            }}
                          >
                            {tag.name}
                          </button>
                        ))}
                    </div>
                  )}

                  <div className="skill-catalog-card__stats muted">
                    <Link href={`/authors/${skill.creator.id}`}>
                      {skill.creator.firstName} {skill.creator.lastName}
                    </Link>
                    {(skill as SkillSummary & { difficulty?: string }).difficulty && (
                      <span>{(skill as SkillSummary & { difficulty?: string }).difficulty}</span>
                    )}
                    {((skill as SkillSummary & { avgRating?: number }).avgRating ?? 0) > 0 && (
                      <span title="Average rating">
                        <StarRating
                          rating={(skill as SkillSummary & { avgRating?: number }).avgRating ?? 0}
                        />{' '}
                        {((skill as SkillSummary & { avgRating?: number }).avgRating ?? 0).toFixed(
                          1,
                        )}
                      </span>
                    )}
                    {((skill as SkillSummary & { runsCount?: number }).runsCount ?? 0) > 0 && (
                      <span>{(skill as SkillSummary & { runsCount?: number }).runsCount} runs</span>
                    )}
                    <span>Updated {formatDate(skill.updatedAt)}</span>
                  </div>

                  <div className="skill-catalog-card__footer">
                    <Badge tone={skill.status === 'published' ? 'success' : 'warning'}>
                      {skill.status}
                    </Badge>
                    <Link href={`/skills/${skill.id}`} className="skill-catalog-card__link">
                      View skill
                    </Link>
                  </div>
                </article>
              ))
            ) : (
              <div className="empty-state stack stack--md">
                <p className="muted">
                  {hasFilters
                    ? 'No skills matched your filters. Adjust or clear them in the sidebar.'
                    : 'No published skills yet. Check back soon.'}
                </p>
              </div>
            )}
          </div>
        </Card>
      </div>
    </div>
  );
}
