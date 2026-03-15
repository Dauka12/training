import { render, screen } from '@testing-library/react';
import { expect, test, vi } from 'vitest';
import { WaterTracker } from './WaterTracker';

test('shows quick add buttons and summary', () => {
  const onQuickAdd = vi.fn();

  render(<WaterTracker currentML={500} targetML={2000} onQuickAdd={onQuickAdd} />);

  expect(screen.getByRole('button', { name: '+250 мл' })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: '+500 мл' })).toBeInTheDocument();
  expect(screen.getByText('500')).toBeInTheDocument();
  expect(screen.getByText('/ 2000 мл')).toBeInTheDocument();
});
