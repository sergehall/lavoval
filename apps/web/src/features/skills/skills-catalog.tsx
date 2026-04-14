'use client';

import Link from 'next/link';
import { useDeferredValue, useState } from 'react';
import type { SkillSummary } from '@lavoval/contracts';
import { Badge } from '@/shared/ui/badge';
import { Card } from '@/shared/ui/card';
import { Input } from '@/shared/ui/input';
import { formatDate } from '@/shared/lib/utils';

function matchesQuery(skill: SkillSummary, query: string) {
  if (!query) {
    return true;
  }

  const haystack = [
    skill.title,
    skill.slug,
    skill.summary,
    skill.creator.firstName,
    skill.creator.lastName,
    skill.creator.email,
  ]
    .join(' ')
    .toLowerCase();

  return haystack.includes(query);
}

export function SkillsCatalog({ skills }: { skills: SkillSummary[] }) {
  const [query, setQuery] = useState('');
  const deferredQuery = useDeferredValue(query.trim().toLowerCase());

  const filteredSkills = skills.filter((skill) => matchesQuery(skill, deferredQuery));

  return (
    <div className="stack stack--lg">
      <Card>
        <div className="skills-search">
          <Input
            aria-label="Search skills"
            className="skills-search__input"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search by title, slug, summary, or author..."
          />
          <div className="inline-actions muted skills-search__meta">
            <span>{filteredSkills.length} matches</span>
            {deferredQuery ? <span>for “{query}”</span> : <span>{skills.length} total skills</span>}
          </div>
        </div>
      </Card>

      <Card>
        <div className="data-list">
          {filteredSkills.length > 0 ? (
            filteredSkills.map((skill) => (
              <article key={skill.id} className="data-list__item">
                <div className="section-heading">
                  <div className="stack stack--sm">
                    <h2>{skill.title}</h2>
                    <p>{skill.summary}</p>
                  </div>
                  <Badge tone={skill.status === 'published' ? 'success' : 'warning'}>
                    {skill.status}
                  </Badge>
                </div>
                <div className="inline-actions muted">
                  <span>
                    By {skill.creator.firstName} {skill.creator.lastName}
                  </span>
                  <span>{skill.modulesCount} modules</span>
                  <span>Updated {formatDate(skill.updatedAt)}</span>
                </div>
                <Link href={`/skills/${skill.id}`} className="muted">
                  Open skill
                </Link>
              </article>
            ))
          ) : (
            <div className="empty-state stack stack--md">
              <Badge tone="warning">No matches</Badge>
              <h2>No skills matched your search.</h2>
              <p className="muted">
                Try a broader phrase, part of the slug, a keyword from the summary, or the author
                name.
              </p>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
