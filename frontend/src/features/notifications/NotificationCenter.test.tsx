import { render, screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import { NotificationCenter } from './NotificationCenter';

test('renders unread notification count and items', () => {
  render(
    <NotificationCenter
      locale="ru"
      items={[
        { id: '1', title: 'План обновлен', type: 'plan_regenerated', read: false, createdAt: '2026-03-14' },
        { id: '2', title: 'Ответ в поддержке', type: 'support_reply', read: true, createdAt: '2026-03-14' }
      ]}
    />
  );

  expect(screen.getByText('1 непрочитанное')).toBeInTheDocument();
  expect(screen.getByText('План обновлен')).toBeInTheDocument();
  expect(screen.getByText('Ответ в поддержке')).toBeInTheDocument();
});
