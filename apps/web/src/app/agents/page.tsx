import type { Metadata } from 'next';
import { fetchAgents } from '@/shared/api/server-client';
import { Card } from '@/shared/ui/card';
import { Badge } from '@/shared/ui/badge';

export const dynamic = 'force-dynamic';

export const metadata: Metadata = {
  title: 'AI Agents — Lavoval',
  description: 'Explore AI agents available on Lavoval to run skills.',
  alternates: { canonical: '/agents' },
};

const CAPABILITY_LABELS: Record<string, string> = {
  supportsCode: 'Code',
  supportsTools: 'Tools',
  supportsWeb: 'Web',
  supportsFiles: 'Files',
  supportsMultimodal: 'Multimodal',
  supportsJsonOutput: 'JSON output',
};

export default async function AgentsPage() {
  const agents = await fetchAgents();

  return (
    <div className="stack stack--lg">
      <Card>
        <div className="stack stack--sm">
          <h1>AI Agents</h1>
          <p className="muted">
            These are the AI models available on Lavoval for running skills. Each agent has
            different capabilities, context limits, and cost profiles.
          </p>
        </div>
      </Card>
      <Card>
        <div className="data-list">
          {(agents ?? []).length > 0 ? (
            (agents ?? []).map((agent) => (
              <article key={agent.id} className="data-list__item">
                <div className="section-heading">
                  <div className="stack stack--sm">
                    <div className="inline-actions">
                      <h2 style={{ margin: 0 }}>{agent.name}</h2>
                      <Badge tone="neutral">{agent.provider}</Badge>
                    </div>
                    <p className="muted" style={{ fontSize: '0.85rem' }}>
                      {agent.modelName}
                    </p>
                    {agent.description && <p>{agent.description}</p>}
                    <div className="inline-actions" style={{ flexWrap: 'wrap', gap: '6px' }}>
                      {Object.entries(CAPABILITY_LABELS).map(([key, label]) => {
                        const val = agent[key as keyof typeof agent];
                        if (!val) return null;
                        return (
                          <Badge key={key} tone="neutral">
                            {label}
                          </Badge>
                        );
                      })}
                      {agent.maxContextTokens && (
                        <Badge tone="neutral">
                          {(agent.maxContextTokens / 1000).toFixed(0)}K context
                        </Badge>
                      )}
                    </div>
                  </div>
                  <Badge tone={agent.status === 'active' ? 'success' : 'warning'}>
                    {agent.status}
                  </Badge>
                </div>
              </article>
            ))
          ) : (
            <div className="empty-state stack stack--md">
              <Badge tone="warning">No agents</Badge>
              <h2>No agents available yet.</h2>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
