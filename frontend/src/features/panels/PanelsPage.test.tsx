import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';
import { AdminPage, TrainerPage } from './pages';
import { t } from '../../shared/i18n';

const fetchMock = vi.fn();

function renderWithClient(node: ReactNode) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={client}>{node}</QueryClientProvider>);
}

afterEach(() => {
  fetchMock.mockReset();
  vi.unstubAllGlobals();
});

test('trainer can inspect user details, read notes, add a note and trigger plan regeneration', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/trainer/users')) {
      return Response.json({
        items: [{ email: 'member@example.com', plan_health: 'healthy', workouts: 3 }]
      });
    }
    if (url.endsWith('/trainer/users/member@example.com') && (!init?.method || init.method === 'GET')) {
      return Response.json({
        user: {
          email: 'member@example.com',
          plan_versions: 2,
          meal_logs: 4,
          water_ml: 1800,
          weekly_checkins: 2,
          assigned_trainer: 'trainer@example.com'
        }
      });
    }
    if (url.endsWith('/trainer/users/member@example.com/notes') && (!init?.method || init.method === 'GET')) {
      return Response.json({
        items: [
          {
            id: 'note-1',
            trainer_email: 'trainer@example.com',
            user_email: 'member@example.com',
            body: 'Keep the pace moderate',
            created_at: '2026-03-15T08:00:00Z'
          }
        ]
      });
    }
    if (url.endsWith('/trainer/users/member@example.com/notes') && init?.method === 'POST') {
      return Response.json({
        note: {
          id: 'note-2',
          trainer_email: 'trainer@example.com',
          user_email: 'member@example.com',
          body: 'Short recovery week',
          created_at: '2026-03-15T10:00:00Z'
        }
      });
    }
    if (url.endsWith('/trainer/users/member@example.com/regenerate-plan') && init?.method === 'POST') {
      return Response.json({
        plan: { id: 'plan-3', title: 'Refreshed plan' }
      });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  const user = userEvent.setup();
  renderWithClient(<TrainerPage locale="ru" />);

  const row = (await screen.findByText('member@example.com')).closest('article');
  expect(row).not.toBeNull();
  await user.click(within(row as HTMLElement).getByRole('button', { name: t('ru', 'trainer.details') }));

  expect((await screen.findAllByText('member@example.com')).length).toBeGreaterThanOrEqual(2);
  expect(await screen.findByText('Keep the pace moderate')).toBeInTheDocument();
  expect((await screen.findAllByText('trainer@example.com')).length).toBeGreaterThan(0);

  await user.type(screen.getByRole('textbox'), 'Short recovery week');
  await user.click(screen.getByRole('button', { name: t('ru', 'trainer.addNote') }));

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/trainer/users/member@example.com/notes',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({ body: 'Short recovery week' })
      })
    )
  );

  await user.click(screen.getByRole('button', { name: t('ru', 'trainer.regenerate') }));

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/trainer/users/member@example.com/regenerate-plan',
      expect.objectContaining({ method: 'POST', credentials: 'include' })
    )
  );
});

test('admin page shows repository-backed operational sections', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/admin/users')) {
      return Response.json({
        items: [
          {
            email: 'member@example.com',
            roles: ['user'],
            assigned_trainer_email: 'trainer@example.com',
            onboarding_done: true,
            active_plan_versions: 1
          }
        ]
      });
    }
    if (url.endsWith('/admin/trainers')) {
      return Response.json({
        items: [{ email: 'trainer@example.com', assigned_users: 1, trainer_note_count: 2 }]
      });
    }
    if (url.endsWith('/admin/support/threads')) {
      return Response.json({
        items: [{ id: 'thread-1', user_email: 'member@example.com', status: 'open', assigned_to_email: 'trainer@example.com', message_count: 2 }]
      });
    }
    if (url.endsWith('/admin/discussions/threads')) {
      return Response.json({
        items: [{ id: 'disc-1', author_email: 'member@example.com', title: 'Meal prep ideas', status: 'visible', reply_count: 1 }]
      });
    }
    if (url.endsWith('/admin/logs/notifications')) {
      return Response.json({
        items: [{ id: 'notif-1', user_email: 'member@example.com', type: 'plan_regenerated', title: 'План обновлен' }]
      });
    }
    if (url.endsWith('/admin/catalog/equipment')) {
      return Response.json({ items: [{ id: 'eq-1', names: { ru: 'Гантели' }, category: 'weights' }] });
    }
    if (url.endsWith('/admin/catalog/exercises')) {
      return Response.json({ items: [{ id: 'ex-1', names: { ru: 'Отжимания' }, slug: 'push-up' }] });
    }
    if (url.endsWith('/admin/logs/ai')) {
      return Response.json({ items: [{ id: 'ai-1', status: 'ok', provider: 'mock', plan_title: 'Весенний план' }] });
    }
    if (url.endsWith('/admin/logs/email')) {
      return Response.json({ items: [{ id: 'mail-1', subject: 'Напоминание о воде', status: 'sent' }] });
    }
    if (url.endsWith('/admin/logs/audit')) {
      return Response.json({ items: [{ id: 'audit-1', action: 'assign_trainer', actor_email: 'admin@example.com' }] });
    }
    if (url.endsWith('/admin/trainers/assign') && init?.method === 'POST') {
      return Response.json({ status: 'assigned' });
    }
    return Response.json({ items: [] });
  });
  vi.stubGlobal('fetch', fetchMock);

  renderWithClient(<AdminPage locale="ru" />);

  expect((await screen.findAllByText('member@example.com')).length).toBeGreaterThan(0);
  expect((await screen.findAllByText('trainer@example.com')).length).toBeGreaterThan(0);
  expect(await screen.findByText('Meal prep ideas')).toBeInTheDocument();
  expect(await screen.findByText('План обновлен')).toBeInTheDocument();
  expect(await screen.findByText('Гантели')).toBeInTheDocument();
  expect(await screen.findByText('Отжимания')).toBeInTheDocument();
  expect(await screen.findByText('Весенний план')).toBeInTheDocument();
  expect(await screen.findByText('Напоминание о воде')).toBeInTheDocument();
  expect(await screen.findByText('assign_trainer')).toBeInTheDocument();
});

test('admin can assign trainer from the panel', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/admin/users')) {
      return Response.json({
        items: [{ email: 'member@example.com', roles: ['user'], onboarding_done: true, active_plan_versions: 0 }]
      });
    }
    if (url.endsWith('/admin/trainers')) {
      return Response.json({
        items: [{ email: 'trainer@local.test', assigned_users: 0, trainer_note_count: 0 }]
      });
    }
    if (
      url.endsWith('/admin/support/threads') ||
      url.endsWith('/admin/discussions/threads') ||
      url.endsWith('/admin/logs/notifications') ||
      url.endsWith('/admin/catalog/equipment') ||
      url.endsWith('/admin/catalog/exercises') ||
      url.endsWith('/admin/logs/ai') ||
      url.endsWith('/admin/logs/email') ||
      url.endsWith('/admin/logs/audit')
    ) {
      return Response.json({ items: [] });
    }
    if (url.endsWith('/admin/trainers/assign') && init?.method === 'POST') {
      return Response.json({ status: 'assigned' });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  const user = userEvent.setup();
  renderWithClient(<AdminPage locale="ru" />);

  await screen.findByRole('option', { name: 'member@example.com' });
  await screen.findByRole('option', { name: 'trainer@local.test' });
  await user.selectOptions(screen.getByLabelText(t('ru', 'admin.userEmail')), 'member@example.com');
  await user.selectOptions(screen.getByLabelText(t('ru', 'admin.trainerEmail')), 'trainer@local.test');
  await user.click(screen.getByRole('button', { name: t('ru', 'admin.assignTrainer') }));

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/trainers/assign',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({
          user_email: 'member@example.com',
          trainer_email: 'trainer@local.test'
        })
      })
    )
  );
});

test('admin can moderate support and discussion threads from the panel', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/admin/users')) {
      return Response.json({ items: [] });
    }
    if (url.endsWith('/admin/trainers')) {
      return Response.json({ items: [{ email: 'trainer@example.com', assigned_users: 1, trainer_note_count: 1 }] });
    }
    if (url.endsWith('/admin/support/threads') && (!init?.method || init.method === 'GET')) {
      return Response.json({
        items: [{ id: 'thread-1', user_email: 'member@example.com', status: 'open', assigned_to_email: 'trainer@example.com', message_count: 2 }]
      });
    }
    if (url.endsWith('/admin/discussions/threads') && (!init?.method || init.method === 'GET')) {
      return Response.json({
        items: [{ id: 'discussion-1', author_email: 'member@example.com', title: 'Meal prep ideas', status: 'visible', reply_count: 1 }]
      });
    }
    if (
      url.endsWith('/admin/logs/notifications') ||
      url.endsWith('/admin/catalog/equipment') ||
      url.endsWith('/admin/catalog/exercises') ||
      url.endsWith('/admin/logs/ai') ||
      url.endsWith('/admin/logs/email') ||
      url.endsWith('/admin/logs/audit')
    ) {
      return Response.json({ items: [] });
    }
    if (url.endsWith('/admin/support/threads/thread-1/status') && init?.method === 'POST') {
      return Response.json({ status: 'saved' });
    }
    if (url.endsWith('/admin/discussions/threads/discussion-1/moderation') && init?.method === 'POST') {
      return Response.json({ status: 'saved' });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  const user = userEvent.setup();
  renderWithClient(<AdminPage locale="ru" />);

  expect(await screen.findByText('Meal prep ideas')).toBeInTheDocument();
  expect((await screen.findAllByText('member@example.com')).length).toBeGreaterThan(0);

  await user.selectOptions(screen.getByLabelText(t('ru', 'admin.supportStatus')), 'resolved');
  await user.click(screen.getByRole('button', { name: t('ru', 'admin.saveSupport') }));

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/support/threads/thread-1/status',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({ status: 'resolved', assigned_to_email: 'trainer@example.com' })
      })
    )
  );

  await user.selectOptions(screen.getByLabelText(t('ru', 'admin.discussionStatus')), 'hidden');
  await user.click(screen.getByRole('button', { name: t('ru', 'admin.saveDiscussion') }));

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/discussions/threads/discussion-1/moderation',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({ status: 'hidden' })
      })
    )
  );
});

test('admin can trigger external catalog import from the panel', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/admin/users')) {
      return Response.json({ items: [] });
    }
    if (url.endsWith('/admin/trainers')) {
      return Response.json({ items: [] });
    }
    if (
      url.endsWith('/admin/support/threads') ||
      url.endsWith('/admin/discussions/threads') ||
      url.endsWith('/admin/logs/notifications') ||
      url.endsWith('/admin/logs/ai') ||
      url.endsWith('/admin/logs/email') ||
      url.endsWith('/admin/logs/audit')
    ) {
      return Response.json({ items: [] });
    }
    if (url.endsWith('/admin/catalog/equipment')) {
      return Response.json({ items: [{ id: 'eq-1', names: { ru: 'Гантели' }, category: 'weights' }] });
    }
    if (url.endsWith('/admin/catalog/exercises')) {
      return Response.json({ items: [{ id: 'ex-1', names: { ru: 'Bear Walk' }, slug: 'bear-walk' }] });
    }
    if (url.endsWith('/admin/catalog/import/wger') && init?.method === 'POST') {
      return Response.json({ imported: { equipment: 1, exercises: 1 } });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  const user = userEvent.setup();
  renderWithClient(<AdminPage locale="ru" />);

  await screen.findByText('Гантели');
  await user.click(screen.getByRole('button', { name: t('ru', 'admin.importWger') }));

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/admin/catalog/import/wger',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
        body: JSON.stringify({ limit: 12 })
      })
    )
  );
});
