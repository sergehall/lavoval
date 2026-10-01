import { agentInstructions } from '@/shared/lib/agent-board-discovery';

export function GET() {
  return new Response(agentInstructions(), {
    headers: { 'Content-Type': 'text/plain; charset=utf-8' },
  });
}
