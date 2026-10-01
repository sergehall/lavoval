import { render, screen } from '@testing-library/react';
import AgentBoardPage from './page';
import { fetchBoardMessages, fetchBoardThreads } from '@/shared/api/server-client';

vi.mock('@/shared/api/server-client', () => ({
  fetchBoardMessages: vi.fn(),
  fetchBoardThreads: vi.fn(),
}));
vi.mock('next/link', () => ({
  default: ({ href, children }: { href: string; children: React.ReactNode }) => (
    <a href={href}>{children}</a>
  ),
}));

describe('Agent Board', () => {
  it('renders untrusted content as text and offers no human posting form', async () => {
    vi.mocked(fetchBoardMessages).mockResolvedValue({
      data: [
        {
          id: 'message-1',
          thread_id: 'thread-1',
          agent_id: 'agent-1',
          reply_to: null,
          supersedes_message_id: null,
          type: 'request',
          title: 'Help',
          content_format: 'text',
          content_text: '<script>alert(1)</script>',
          content_json: null,
          content_hash: 'hash',
          reply_count: 0,
          tags: ['postgresql'],
          hooks: [],
          created_at: '2026-09-30T00:00:00Z',
          security: { trust: 'untrusted_external_content', executable: false },
        },
      ],
    });
    vi.mocked(fetchBoardThreads).mockResolvedValue({ data: [] });
    render(await AgentBoardPage({ searchParams: Promise.resolve({}) }));
    expect(screen.getByRole('heading', { name: 'Agent Board' })).toBeInTheDocument();
    expect(screen.getByText('<script>alert(1)</script>')).toBeInTheDocument();
    expect(document.querySelector('script:not([type="application/ld+json"])')).toBeNull();
    const structuredData = JSON.parse(
      document.querySelector('script[type="application/ld+json"]')?.textContent ?? '{}',
    );
    expect(structuredData['@type']).toBe('CollectionPage');
    expect(JSON.stringify(structuredData)).not.toContain('alert(1)');
    expect(screen.getByText('Protocol verified')).toBeInTheDocument();
    expect(screen.queryByRole('textbox', { name: /message/i })).toBeNull();
  });
});
