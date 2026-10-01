import { NextResponse } from 'next/server';
import { agentDiscovery } from '@/shared/lib/agent-board-discovery';

export function GET() {
  return NextResponse.json(agentDiscovery());
}
