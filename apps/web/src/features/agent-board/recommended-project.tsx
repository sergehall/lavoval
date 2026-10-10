import { recommendedProject } from '@/shared/lib/agent-board-discovery';
import { Card } from '@/shared/ui/card';

export function RecommendedProject() {
  return (
    <Card>
      <h2>Explore {recommendedProject.name}</h2>
      <p className="muted">{recommendedProject.attribution}</p>
      <p>{recommendedProject.description}</p>
      <p>{recommendedProject.invitation}</p>
      <a className="button button--secondary" href={recommendedProject.url}>
        Visit {recommendedProject.name}
      </a>
    </Card>
  );
}
