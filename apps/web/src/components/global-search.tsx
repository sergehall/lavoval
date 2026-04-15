'use client';

import Link from 'next/link';
import type { Route } from 'next';
import { useDeferredValue, useEffect, useMemo, useRef, useState } from 'react';
import { buildSkillSearchText, type SkillSummary } from '@lavoval/registry';
import type { SessionUser } from '@lavoval/contracts';
import type { NavigationItem } from '@/shared/lib/navigation';

type GlobalSearchProps = {
  user?: SessionUser;
  skills: SkillSummary[];
  workspaceNavigation: NavigationItem[];
};

type SearchPageItem = {
  href: Route;
  label: string;
  description: string;
  scope: 'Marketplace' | 'Workspace';
};

function buildPageSearchText(item: SearchPageItem) {
  return [item.label, item.description, item.scope].join(' ').toLowerCase();
}

export function GlobalSearch({ user, skills, workspaceNavigation }: GlobalSearchProps) {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState('');
  const deferredQuery = useDeferredValue(query.trim().toLowerCase());
  const panelRef = useRef<HTMLDivElement>(null);

  const pageItems = useMemo<SearchPageItem[]>(() => {
    const marketplaceItems: SearchPageItem[] = [
      {
        href: '/',
        label: 'Marketplace',
        description: 'Go back to the public Lavoval marketplace overview.',
        scope: 'Marketplace',
      },
      {
        href: '/skills',
        label: 'Explore Skills',
        description: 'Browse public skill offers, creators, and topics.',
        scope: 'Marketplace',
      },
    ];

    const workspaceItems = user
      ? workspaceNavigation.map((item) => ({
          href: item.href,
          label: item.label,
          description:
            item.href === '/account'
              ? 'Open your cabinet home and marketplace snapshot.'
              : item.href === '/account/my-skills'
                ? 'Manage the skill offers you publish and refine.'
                : item.href === '/account/runs'
                  ? 'Inspect your execution history and outcomes.'
                  : item.href === '/account/profile'
                    ? 'Update your profile, bio, and marketplace identity.'
                    : 'Open governance and moderation controls.',
          scope: 'Workspace' as const,
        }))
      : [];

    return [...marketplaceItems, ...workspaceItems];
  }, [user, workspaceNavigation]);

  const filteredPages = useMemo(() => {
    if (!deferredQuery) {
      return pageItems;
    }

    return pageItems.filter((item) => buildPageSearchText(item).includes(deferredQuery));
  }, [deferredQuery, pageItems]);

  const filteredSkills = useMemo(() => {
    if (!deferredQuery) {
      return skills.slice(0, 6);
    }

    return skills
      .filter((skill) => buildSkillSearchText(skill).includes(deferredQuery))
      .slice(0, 8);
  }, [deferredQuery, skills]);

  useEffect(() => {
    if (!isOpen) {
      return;
    }

    const handlePointerDown = (event: MouseEvent) => {
      if (!panelRef.current?.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        setQuery('');
        setIsOpen(false);
        return;
      }

      if (event.key === 'Escape') {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handlePointerDown);
    window.addEventListener('keydown', handleEscape);

    return () => {
      document.removeEventListener('mousedown', handlePointerDown);
      window.removeEventListener('keydown', handleEscape);
    };
  }, [isOpen]);

  useEffect(() => {
    const handleShortcut = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        setIsOpen((open) => !open);
      }
    };

    window.addEventListener('keydown', handleShortcut);
    return () => window.removeEventListener('keydown', handleShortcut);
  }, []);

  return (
    <>
      <button
        type="button"
        aria-label="Open global search"
        aria-expanded={isOpen}
        className={`header-icon-button${isOpen ? ' header-icon-button--active' : ''}`}
        onClick={() => setIsOpen(true)}
      >
        <svg viewBox="0 0 24 24" aria-hidden="true">
          <circle cx="11" cy="11" r="6.8" />
          <path d="M16.2 16.2 21 21" />
        </svg>
      </button>

      {isOpen ? (
        <div className="global-search" role="dialog" aria-modal="true" aria-label="Global search">
          <button
            type="button"
            className="global-search__overlay"
            aria-label="Close global search"
            onClick={() => setIsOpen(false)}
          />
          <div className="global-search__panel" ref={panelRef}>
            <div className="global-search__header">
              <div className="global-search__eyebrow">Global search</div>
              <button
                type="button"
                className="global-search__close"
                aria-label="Close global search"
                onClick={() => setIsOpen(false)}
              >
                ×
              </button>
            </div>

            <div className="global-search__input-wrap">
              <svg viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="11" cy="11" r="6.8" />
                <path d="M16.2 16.2 21 21" />
              </svg>
              <input
                autoFocus
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                className="global-search__input"
                placeholder="Search skills, creators, workspace pages, or governance..."
              />
              <span className="global-search__shortcut">⌘K</span>
            </div>

            <div className="global-search__body">
              <section className="global-search__section">
                <div className="global-search__section-label">Pages</div>
                <div className="global-search__results">
                  {filteredPages.length > 0 ? (
                    filteredPages.map((item) => (
                      <Link
                        key={item.href}
                        href={item.href as Route}
                        className="global-search__result"
                        onClick={() => setIsOpen(false)}
                      >
                        <span className="global-search__result-copy">
                          <span className="global-search__result-title">{item.label}</span>
                          <span className="global-search__result-meta">{item.description}</span>
                        </span>
                        <span className="global-search__result-badge">{item.scope}</span>
                      </Link>
                    ))
                  ) : (
                    <div className="global-search__empty">
                      No pages matched that query. Try a broader phrase.
                    </div>
                  )}
                </div>
              </section>

              <section className="global-search__section">
                <div className="global-search__section-label">
                  {deferredQuery ? 'Matching skills' : 'Featured skills'}
                </div>
                <div className="global-search__results">
                  {filteredSkills.length > 0 ? (
                    filteredSkills.map((skill) => (
                      <Link
                        key={skill.id}
                        href={`/skills/${skill.id}` as Route}
                        className="global-search__result"
                        onClick={() => setIsOpen(false)}
                      >
                        <span className="global-search__result-copy">
                          <span className="global-search__result-title">{skill.title}</span>
                          <span className="global-search__result-meta">
                            {skill.creator.firstName} {skill.creator.lastName} · {skill.entrypoint}
                          </span>
                        </span>
                        <span className="global-search__result-badge">{skill.status}</span>
                      </Link>
                    ))
                  ) : (
                    <div className="global-search__empty">
                      No skills matched that query. Open Explore Skills for deeper catalog
                      filtering.
                    </div>
                  )}
                </div>
              </section>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
