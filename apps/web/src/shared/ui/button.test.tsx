import { render, screen } from '@testing-library/react';
import { Button } from './button';

describe('Button', () => {
  it('renders children and variant class', () => {
    render(<Button variant="secondary">Save changes</Button>);
    const button = screen.getByRole('button', { name: 'Save changes' });
    expect(button).toHaveClass('button--secondary');
  });
});
