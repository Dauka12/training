import { render, screen } from '@testing-library/react';
import { afterEach, expect, test, vi } from 'vitest';
import { AppRouter } from '../../app/router';
import { t } from '../../shared/i18n';

const fetchMock = vi.fn();

afterEach(() => {
  fetchMock.mockReset();
  vi.unstubAllGlobals();
});

test('renders a media-rich exercise screen with technique, substitutions and contraindications', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/notifications')) {
      return Response.json({ items: [] });
    }
    if (url.endsWith('/catalog/exercises/20000000-0000-0000-0000-000000000001')) {
      return Response.json({
        exercise: {
          id: '20000000-0000-0000-0000-000000000001',
          slug: 'goblet-squat',
          name: 'Присед с гантелью',
          description: 'Базовое упражнение на ноги',
          technique: 'Держите корпус стабильно и направляйте колени по линии стоп.',
          movement_pattern: 'squat',
          difficulty: 'beginner',
          location_type: 'mixed',
          media_url: 'https://example.com/goblet-squat.jpg',
          contraindication_tags: ['knee_pain'],
          equipment: [{ id: 'eq-1', name: 'Гантели' }],
          substitutions: [{ id: '20000000-0000-0000-0000-000000000002', name: 'Отжимания' }]
        }
      });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  render(
    <AppRouter
      initialEntries={['/exercise/20000000-0000-0000-0000-000000000001']}
      initialAuth={{ isAuthenticated: true, role: 'user', onboardingDone: true, email: 'member@example.com' }}
    />
  );

  expect(await screen.findByRole('heading', { name: 'Присед с гантелью' })).toBeInTheDocument();
  expect(screen.getByText('Базовое упражнение на ноги')).toBeInTheDocument();
  expect(screen.getByText('Держите корпус стабильно и направляйте колени по линии стоп.')).toBeInTheDocument();
  expect(screen.getByText('knee_pain')).toBeInTheDocument();
  expect(screen.getByRole('img', { name: 'Присед с гантелью' })).toHaveAttribute('src', 'https://example.com/goblet-squat.jpg');
  expect(screen.getByRole('link', { name: 'Отжимания' })).toBeInTheDocument();
  expect(screen.getByText(t('ru', 'exercise.technique'))).toBeInTheDocument();
});
