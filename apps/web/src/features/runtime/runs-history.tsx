'use client';

import type { Route } from 'next';
import Link from 'next/link';
import { useDeferredValue, useMemo, useState } from 'react';
import type { SkillRun } from '@lavoval/contracts/runtime';
import { Badge } from '@/shared/ui/badge';
import { Button } from '@/shared/ui/button';
import { Card } from '@/shared/ui/card';
import { Input } from '@/shared/ui/input';
import { formatDate } from '@/shared/lib/utils';

function toneForStatus(status: SkillRun['status']) {
  switch (status) {
    case 'completed':
      return 'success';
    case 'failed':
      return 'warning';
    case 'running':
      return 'warning';
    default:
      return 'neutral';
  }
}

const statusOptions: Array<{ value: 'all' | SkillRun['status']; label: string }> = [
  { value: 'all', label: 'All' },
  { value: 'completed', label: 'Completed' },
  { value: 'running', label: 'Running' },
  { value: 'failed', label: 'Failed' },
  { value: 'queued', label: 'Queued' },
];

function matchesQuery(run: SkillRun, query: string) {
  if (!query) {
    return true;
  }

  const haystack = [
    run.id,
    run.skill.title,
    run.skill.slug,
    run.skill.entrypoint,
    extractPreview(run.output),
  ]
    .join(' ')
    .toLowerCase();

  return haystack.includes(query);
}

function extractPreview(output: SkillRun['output']) {
  if (!output || Object.keys(output).length === 0) {
    return 'No output yet.';
  }

  if (typeof output.result === 'string') {
    return trimPreview(output.result);
  }

  if (Array.isArray(output.keywords)) {
    return trimPreview(output.keywords.join(', '));
  }

  return trimPreview(JSON.stringify(output));
}

function trimPreview(value: string, maxLength = 120) {
  return value.length > maxLength ? `${value.slice(0, maxLength).trim()}...` : value;
}

export function RunsHistory({
  runs,
  detailBasePath = '/account/runs',
}: {
  runs: SkillRun[];
  detailBasePath?: '/account/runs' | '/admin/runs';
}) {
  const [statusFilter, setStatusFilter] = useState<'all' | SkillRun['status']>('all');
  const [query, setQuery] = useState('');
  const deferredQuery = useDeferredValue(query.trim().toLowerCase());

  const filteredRuns = useMemo(
    () =>
      runs.filter(
        (run) =>
          (statusFilter === 'all' || run.status === statusFilter) && matchesQuery(run, deferredQuery),
      ),
    [runs, statusFilter, deferredQuery],
  );

  return (
    <div className="stack stack--lg">
      <Card>
        <div className="skills-search">
          <Input
            aria-label="Search runs"
            className="skills-search__input"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search by skill title, slug, entrypoint, run id, or output preview..."
          />
          <div className="toolbar">
            {statusOptions.map((option) => (
              <Button
                key={option.value}
                type="button"
                variant={statusFilter === option.value ? 'secondary' : 'ghost'}
                onClick={() => setStatusFilter(option.value)}
              >
                {option.label}
              </Button>
            ))}
          </div>
          <div className="inline-actions muted skills-search__meta">
            <span>{filteredRuns.length} matches</span>
            <span>{runs.length} total runs</span>
          </div>
        </div>
      </Card>

      <Card>
        <div className="data-list">
          {filteredRuns.length > 0 ? (
            filteredRuns.map((run) => (
              <article key={run.id} className="data-list__item">
                <div className="section-heading">
                  <div className="stack stack--sm">
                    <h2>{run.skill.title}</h2>
                    <p className="muted">
                      `{run.skill.slug}` via <strong>{run.skill.entrypoint}</strong>
                    </p>
                  </div>
                  <Badge tone={toneForStatus(run.status)}>{run.status}</Badge>
                </div>
                <div className="inline-actions muted">
                  <span>Created {formatDate(run.createdAt)}</span>
                  <span>{run.meta.inputKeysCount} input fields</span>
                  <span>{run.meta.outputKeysCount} output fields</span>
                  <span>Run {run.id.slice(0, 8)}</span>
                  {run.meta.durationMs ? <span>{run.meta.durationMs} ms</span> : null}
                </div>
                <p>{extractPreview(run.output)}</p>
                <Link href={`${detailBasePath}/${run.id}` as Route} className="muted">
                  Open run details
                </Link>
              </article>
            ))
          ) : (
            <div className="empty-state stack stack--md">
              <Badge tone="warning">No runs yet</Badge>
              <h2>You have not executed any skills yet.</h2>
              <p className="muted">
                Open a public skill, run it from the detail page, and your history will start
                showing up here.
              </p>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
