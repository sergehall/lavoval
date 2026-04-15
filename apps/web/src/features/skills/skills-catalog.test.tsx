import { render, screen, fireEvent } from '@testing-library/react';
import type { SkillSummary } from '@lavoval/registry';
import { SkillsCatalog } from './skills-catalog';

vi.mock('next/link', () => ({
  default: ({
    href,
    children,
    className,
  }: {
    href: string;
    children: React.ReactNode;
    className?: string;
  }) => (
    <a href={href} className={className}>
      {children}
    </a>
  ),
}));

function makeSkill(overrides: Partial<SkillSummary> = {}): SkillSummary {
  return {
    id: '1',
    slug: 'test-skill',
    title: 'Test Skill',
    summary: 'A comprehensive test skill for the catalog',
    provider: 'internal',
    entrypoint: 'echo',
    config: {},
    status: 'published',
    visibility: 'public',
    creator: { id: 'c1', email: 'creator@test.com', firstName: 'Ada', lastName: 'Lovelace' },
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    modulesCount: 2,
    ...overrides,
  };
}

describe('SkillsCatalog', () => {
  it('renders all skills when no query is set', () => {
    const skills = [
      makeSkill({ id: '1', title: 'Clean Architecture' }),
      makeSkill({ id: '2', title: 'Test-Driven Development' }),
    ];

    render(<SkillsCatalog skills={skills} />);

    expect(screen.getByText('Clean Architecture')).toBeInTheDocument();
    expect(screen.getByText('Test-Driven Development')).toBeInTheDocument();
    expect(screen.getByText('2 matches')).toBeInTheDocument();
  });

  it('filters skills by title query', () => {
    const skills = [
      makeSkill({ id: '1', title: 'Clean Architecture' }),
      makeSkill({ id: '2', title: 'Test-Driven Development' }),
    ];

    render(<SkillsCatalog skills={skills} />);

    fireEvent.change(screen.getByRole('textbox', { name: /search skills/i }), {
      target: { value: 'clean' },
    });

    expect(screen.getByText('Clean Architecture')).toBeInTheDocument();
    expect(screen.queryByText('Test-Driven Development')).not.toBeInTheDocument();
    expect(screen.getByText('1 matches')).toBeInTheDocument();
  });

  it('shows empty state when no skills match the query', () => {
    const skills = [makeSkill({ id: '1', title: 'Clean Architecture' })];

    render(<SkillsCatalog skills={skills} />);

    fireEvent.change(screen.getByRole('textbox', { name: /search skills/i }), {
      target: { value: 'nonexistent-keyword-xyz' },
    });

    expect(screen.getByText('No matches')).toBeInTheDocument();
    expect(screen.queryByText('Clean Architecture')).not.toBeInTheDocument();
  });

  it('renders with an initial query pre-applied', () => {
    const skills = [
      makeSkill({ id: '1', title: 'Clean Architecture' }),
      makeSkill({ id: '2', title: 'Test-Driven Development' }),
    ];

    render(<SkillsCatalog skills={skills} initialQuery="clean" />);

    expect(screen.getByText('Clean Architecture')).toBeInTheDocument();
    expect(screen.queryByText('Test-Driven Development')).not.toBeInTheDocument();
  });

  it('shows skills count in the header when no query', () => {
    const skills = [makeSkill({ id: '1' }), makeSkill({ id: '2' }), makeSkill({ id: '3' })];

    render(<SkillsCatalog skills={skills} />);

    expect(screen.getByText('3 public skill offers')).toBeInTheDocument();
  });
});
