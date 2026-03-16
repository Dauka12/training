import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, expect, test, vi } from 'vitest';
import { PlanPage } from './pages';

const fetchMock = vi.fn();

function renderPlanPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <PlanPage locale="ru" />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

afterEach(() => {
  fetchMock.mockReset();
});

test('renders detailed weeks, nutrition notes, meal examples and adaptation rules', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/plans/active')) {
      return Response.json({
        plan: {
          title: 'Plan title',
          summary: 'Plan summary',
          warnings: ['Keep effort moderate'],
          nutrition: {
            daily_calories: 2200,
            protein_g: 160,
            carbs_g: 220,
            fat_g: 70,
            daily_water_ml: 2600,
            training_note: 'Use more carbs around training',
            rest_note: 'Keep meals simple on rest days',
            hydration_note: 'Drink water steadily',
            meal_examples: [
              { slot: 'breakfast', examples: ['Oats + eggs', 'Yogurt + berries'] }
            ]
          },
          schedule: [{ id: 'session-1', session_name: 'Upper A', weekday: 'monday' }],
          weeks: [
            {
              week_index: 1,
              days: [
                {
                  weekday: 'monday',
                  session_name: 'Upper A',
                  goal: 'Strength base',
                  estimated_minutes: 45,
                  warmup: ['5 minute cardio'],
                  exercises: [
                    {
                      order: 1,
                      exercise_id: 'exercise-1',
                      exercise_name: 'Dumbbell bench press',
                      sets: 3,
                      reps: '8-10',
                      rest_sec: 90,
                      effort_note: 'RPE 7',
                      notes: 'Control the tempo'
                    }
                  ],
                  cooldown: ['Chest stretch']
                }
              ]
            }
          ],
          adaptation_rules: ['If a session is missed, move it to the next valid slot']
        }
      });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  renderPlanPage();

  expect(await screen.findByText('Plan title')).toBeInTheDocument();
  expect(screen.getByTestId('plan-layout')).toBeInTheDocument();
  expect(screen.getByTestId('plan-overview')).toBeInTheDocument();
  expect(screen.getByTestId('plan-feed')).toBeInTheDocument();
  expect(screen.getByText('Use more carbs around training')).toBeInTheDocument();
  expect(screen.getByText('Oats + eggs, Yogurt + berries')).toBeInTheDocument();
  expect(screen.getByText(/Strength base/)).toBeInTheDocument();
  expect(screen.getByText('Dumbbell bench press')).toBeInTheDocument();
  expect(screen.getByText('If a session is missed, move it to the next valid slot')).toBeInTheDocument();
});
