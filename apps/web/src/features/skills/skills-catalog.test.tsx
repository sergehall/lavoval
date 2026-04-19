import { render, screen } from '@testing-library/react';
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

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn() }),
  usePathname: () => '/skills',
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

const defaultProps = {
  categories: [],
  tags: [],
  activeFilter: {},
};

describe('SkillsCatalog', () => {
  it('renders all skills', () => {
    const skills = [
      makeSkill({ id: '1', title: 'Clean Architecture' }),
      makeSkill({ id: '2', title: 'Test-Driven Development' }),
    ];

    render(<SkillsCatalog skills={skills} {...defaultProps} />);

    expect(screen.getByText('Clean Architecture')).toBeInTheDocument();
    expect(screen.getByText('Test-Driven Development')).toBeInTheDocument();
  });

  it('shows skill count in results header', () => {
    const skills = [
      makeSkill({ id: '1', title: 'Clean Architecture' }),
      makeSkill({ id: '2', title: 'Test-Driven Development' }),
    ];

    render(<SkillsCatalog skills={skills} {...defaultProps} />);

    expect(screen.getByText('2 skills')).toBeInTheDocument();
  });

  it('shows singular skill count', () => {
    const skills = [makeSkill({ id: '1' })];

    render(<SkillsCatalog skills={skills} {...defaultProps} />);

    expect(screen.getByText('1 skill')).toBeInTheDocument();
  });

  it('shows empty state when no skills and no filters', () => {
    render(<SkillsCatalog skills={[]} {...defaultProps} />);

    expect(screen.getByText(/No published skills yet/)).toBeInTheDocument();
  });

  it('shows filtered empty state when activeFilter has values', () => {
    render(
      <SkillsCatalog
        skills={[]}
        categories={[]}
        tags={[]}
        activeFilter={{ difficulty: 'senior' }}
      />,
    );

    expect(screen.getByText(/No skills matched your filters/)).toBeInTheDocument();
  });

  it('shows search query in sidebar counter', () => {
    const skills = [makeSkill({ id: '1', title: 'Clean Architecture' })];

    render(
      <SkillsCatalog skills={skills} categories={[]} tags={[]} activeFilter={{ q: 'clean' }} />,
    );

    expect(screen.getByText(/matching/)).toBeInTheDocument();
  });

  it('renders View skill links for each skill', () => {
    const skills = [makeSkill({ id: 'abc', title: 'Clean Architecture' })];

    render(<SkillsCatalog skills={skills} {...defaultProps} />);

    const link = screen.getByRole('link', { name: /view skill/i });
    expect(link).toHaveAttribute('href', '/skills/abc');
  });
});
