import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, expect, test, vi } from 'vitest';
import { TrackPage } from './pages';
import { t } from '../../shared/i18n';

const fetchMock = vi.fn();

function renderTrackPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <TrackPage locale="ru" />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

afterEach(() => {
  fetchMock.mockReset();
});

test('track page exposes detailed workout logging controls and weekly hydration summary', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/plans/active')) {
      return Response.json({
        plan: {
          schedule: [
            { id: 'session-1', session_name: 'Верх тела A' },
            { id: 'session-2', session_name: 'Низ тела B' }
          ]
        }
      });
    }
    if (url.endsWith('/tracking/hydration/summary')) {
      return Response.json({
        target_ml: 2600,
        consumed_ml: 1200,
        adherence: 0.46,
        weekly_target_ml: 18200,
        weekly_consumed_ml: 9100,
        weekly_adherence: 0.5
      });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  renderTrackPage();

  expect(await screen.findByRole('heading', { name: t('ru', 'track.title') })).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'track.workout.session'))).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'track.workout.status'))).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'track.workout.discomfort'))).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'track.workout.difficulty'))).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'track.workout.completedAt'))).toBeInTheDocument();
  expect(screen.getByText(t('ru', 'track.hydration.weekly'))).toBeInTheDocument();
  await waitFor(() =>
    expect(
      screen.getAllByText((_, element) => (element?.textContent ?? '').includes('9100 / 18200 мл')).length
    ).toBeGreaterThan(0)
  );
});
