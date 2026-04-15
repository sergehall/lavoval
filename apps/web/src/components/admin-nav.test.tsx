import { render, screen } from '@testing-library/react';
import { AdminNav } from './admin-nav';

vi.mock('next/navigation', () => ({
  usePathname: vi.fn(),
}));

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

import { usePathname } from 'next/navigation';

function setPathname(path: string) {
  vi.mocked(usePathname).mockReturnValue(path);
}

describe('AdminNav', () => {
  it('renders all nav items', () => {
    setPathname('/admin');
    render(<AdminNav />);

    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('Skills')).toBeInTheDocument();
    expect(screen.getByText('Runs')).toBeInTheDocument();
    expect(screen.getByText('Enrollments')).toBeInTheDocument();
  });

  it('renders the Governance badge', () => {
    setPathname('/admin');
    render(<AdminNav />);

    expect(screen.getByText('Governance')).toBeInTheDocument();
  });

  it('marks Dashboard active on exact /admin match', () => {
    setPathname('/admin');
    render(<AdminNav />);

    const dashboardLink = screen.getByRole('link', { name: 'Dashboard' });
    expect(dashboardLink).toHaveClass('admin-nav__link--active');
  });

  it('does not mark Dashboard active on a sub-route like /admin/users', () => {
    setPathname('/admin/users');
    render(<AdminNav />);

    const dashboardLink = screen.getByRole('link', { name: 'Dashboard' });
    expect(dashboardLink).not.toHaveClass('admin-nav__link--active');
  });

  it('marks Users active when pathname is /admin/users', () => {
    setPathname('/admin/users');
    render(<AdminNav />);

    const usersLink = screen.getByRole('link', { name: 'Users' });
    expect(usersLink).toHaveClass('admin-nav__link--active');
  });

  it('marks Users active on a nested user detail page', () => {
    setPathname('/admin/users/abc-123');
    render(<AdminNav />);

    const usersLink = screen.getByRole('link', { name: 'Users' });
    expect(usersLink).toHaveClass('admin-nav__link--active');
  });

  it('marks Skills active when pathname is /admin/skills', () => {
    setPathname('/admin/skills');
    render(<AdminNav />);

    const skillsLink = screen.getByRole('link', { name: 'Skills' });
    expect(skillsLink).toHaveClass('admin-nav__link--active');
  });

  it('marks Enrollments active and not others on /admin/enrollments', () => {
    setPathname('/admin/enrollments');
    render(<AdminNav />);

    expect(screen.getByRole('link', { name: 'Enrollments' })).toHaveClass(
      'admin-nav__link--active',
    );
    expect(screen.getByRole('link', { name: 'Dashboard' })).not.toHaveClass(
      'admin-nav__link--active',
    );
    expect(screen.getByRole('link', { name: 'Users' })).not.toHaveClass('admin-nav__link--active');
  });

  it('has correct href attributes on all links', () => {
    setPathname('/admin');
    render(<AdminNav />);

    expect(screen.getByRole('link', { name: 'Dashboard' })).toHaveAttribute('href', '/admin');
    expect(screen.getByRole('link', { name: 'Users' })).toHaveAttribute('href', '/admin/users');
    expect(screen.getByRole('link', { name: 'Skills' })).toHaveAttribute('href', '/admin/skills');
    expect(screen.getByRole('link', { name: 'Runs' })).toHaveAttribute('href', '/admin/runs');
    expect(screen.getByRole('link', { name: 'Enrollments' })).toHaveAttribute(
      'href',
      '/admin/enrollments',
    );
  });
});
