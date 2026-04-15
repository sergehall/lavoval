import { render, screen } from '@testing-library/react';
import { Badge } from './badge';

describe('Badge', () => {
  it('renders children', () => {
    render(<Badge>published</Badge>);
    expect(screen.getByText('published')).toBeInTheDocument();
  });

  it('applies success tone class', () => {
    const { container } = render(<Badge tone="success">active</Badge>);
    expect(container.firstChild).toHaveClass('badge--success');
  });

  it('applies warning tone class', () => {
    const { container } = render(<Badge tone="warning">draft</Badge>);
    expect(container.firstChild).toHaveClass('badge--warning');
  });

  it('defaults to neutral tone', () => {
    const { container } = render(<Badge>neutral</Badge>);
    expect(container.firstChild).toHaveClass('badge--neutral');
  });
});
