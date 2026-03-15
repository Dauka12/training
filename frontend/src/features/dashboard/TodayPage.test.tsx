import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, expect, test, vi } from 'vitest';
import { TodayPage } from './pages';
import { t } from '../../shared/i18n';

const fetchMock = vi.fn();

function renderTodayPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <TodayPage locale="ru" />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

afterEach(() => {
  fetchMock.mockReset();
});

test('today page renders a dashboard-like overview with quick actions and hydration progress', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/dashboard/today')) {
      return Response.json({
        today_workout: { session_name: 'Верх тела A' },
        meal_status: 'followed',
        hydration: { target_ml: 2600, consumed_ml: 1200 },
        quick_actions: ['log_workout', 'log_meal', 'log_water'],
        current_week_progress: { completed_sessions: 2 },
        latest_weight_trend: 84,
        next_session: { session_name: 'Низ тела B' },
        plan_health: 'healthy',
        notifications_unread: 3
      });
    }
    if (url.endsWith('/notifications')) {
      return Response.json({
        items: [{ id: 'n1', title: 'План обновлен', type: 'plan_regenerated', read: false, createdAt: '2026-03-15T08:00:00Z', target_url: '/plan' }]
      });
    }
    return Response.json({ items: [] });
  });
  vi.stubGlobal('fetch', fetchMock);

  renderTodayPage();

  expect(await screen.findByRole('heading', { name: t('ru', 'today.heroTitle') })).toBeInTheDocument();
  expect(screen.getAllByText('Верх тела A').length).toBeGreaterThan(0);
  expect(screen.getAllByText((_, element) => element?.textContent?.includes('Низ тела B') ?? false).length).toBeGreaterThan(0);
  expect(screen.getByRole('link', { name: t('ru', 'today.logWorkout') })).toBeInTheDocument();
  expect(screen.getByRole('link', { name: t('ru', 'today.logMeal') })).toBeInTheDocument();
  expect(screen.getByText(t('ru', 'track.water.title'))).toBeInTheDocument();
  expect(screen.getAllByText(/1200/).length).toBeGreaterThan(0);
  expect(screen.getByText('План обновлен')).toBeInTheDocument();
});
