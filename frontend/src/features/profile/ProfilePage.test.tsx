import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';
import { ProfilePage } from './ProfilePage';
import { useAuthStore } from '../../shared/auth/store';
import { t } from '../../shared/i18n';

const fetchMock = vi.fn();

function renderProfilePage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <ProfilePage locale="ru" />
    </QueryClientProvider>
  );
}

afterEach(() => {
  fetchMock.mockReset();
  vi.unstubAllGlobals();
  useAuthStore.getState().clear();
});

test('loads and saves notification preferences', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/me')) {
      return Response.json({
        email: 'member@example.com',
        locale: 'ru',
        theme: 'light',
        onboarding_done: true,
        water_override_ml: 3100,
        water_target_ml: 3100
      });
    }
    if (url.endsWith('/notifications/preferences') && (!init?.method || init.method === 'GET')) {
      return Response.json({
        preferences: {
          hydration_reminder: true,
          email_enabled: false
        }
      });
    }
    if (url.endsWith('/me/preferences') && init?.method === 'PUT') {
      return Response.json({ status: 'saved' });
    }
    if (url.endsWith('/notifications/preferences') && init?.method === 'PUT') {
      return Response.json({ status: 'saved' });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  const user = userEvent.setup();
  renderProfilePage();

  expect(await screen.findByText('member@example.com', {}, { timeout: 10000 })).toBeInTheDocument();
  await waitFor(() => expect(screen.getByLabelText(t('ru', 'profile.waterOverride'))).toHaveValue(3100));

  const hydrationCheckbox = await screen.findByLabelText(t('ru', 'profile.notificationsHydration'));
  const emailCheckbox = screen.getByLabelText(t('ru', 'profile.notificationsEmail'));

  await user.click(hydrationCheckbox);
  await user.click(emailCheckbox);
  await user.click(screen.getByRole('button', { name: t('ru', 'profile.saveNotifications') }));

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/notifications/preferences',
      expect.objectContaining({
        method: 'PUT',
        credentials: 'include'
      })
    )
  );

  const notificationCall = fetchMock.mock.calls.find(
    ([input, init]) => String(input).endsWith('/notifications/preferences') && init?.method === 'PUT'
  );
  expect(notificationCall).toBeTruthy();
  expect(String(notificationCall?.[1]?.body)).toContain('"hydration_reminder":false');
  expect(String(notificationCall?.[1]?.body)).toContain('"email_enabled":true');
});

test('loads existing onboarding preferences and saves extended planning fields through a stepped flow', async () => {
  useAuthStore.getState().setAuthenticated({
    role: 'user',
    email: 'member@example.com',
    onboardingDone: false
  });

  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/me')) {
      return Response.json({
        email: 'member@example.com',
        locale: 'ru',
        theme: 'dark',
        onboarding_done: false,
        profile: {
          age: 31,
          biological_sex: 'female',
          height_cm: 167,
          current_weight_kg: 72,
          target_weight_kg: 65,
          primary_goal: 'lose_fat',
          program_duration_weeks: 10,
          experience_level: 'intermediate',
          activity_level: 'moderate',
          training_location: 'home',
          timezone: 'Asia/Qyzylorda',
          preferred_training_style: 'low_impact_strength',
          preferred_meal_style: 'simple_prep',
          hydration_preference: 'small_frequent_sips',
          injuries: ['knee_discomfort'],
          dietary_preferences: ['high_protein'],
          avoid_foods: ['lactose'],
          equipment_ids: ['10000000-0000-0000-0000-000000000001'],
          available_training_days: [{ weekday: 'tuesday', duration_min: 40 }]
        }
      });
    }
    if (url.endsWith('/catalog/equipment')) {
      return Response.json({
        items: [
          {
            id: '10000000-0000-0000-0000-000000000001',
            names: { ru: 'Гантели', kk: 'Gantel' },
            descriptions: { ru: 'Свободные веса', kk: 'Erkin salmaq' },
            category: 'weights',
            location_type: 'mixed'
          },
          {
            id: '10000000-0000-0000-0000-000000000002',
            names: { ru: 'Коврик', kk: 'Tosek' },
            descriptions: { ru: 'Для домашних сессий', kk: 'Ui zhattyguyna' },
            category: 'recovery',
            location_type: 'home'
          }
        ]
      });
    }
    if (url.endsWith('/notifications/preferences') && (!init?.method || init.method === 'GET')) {
      return Response.json({
        preferences: {
          hydration_reminder: true,
          email_enabled: false
        }
      });
    }
    if (url.endsWith('/onboarding') && init?.method === 'PUT') {
      return Response.json({ status: 'saved', water_target_ml: 2400 });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  const user = userEvent.setup();
  renderProfilePage();

  expect(await screen.findByText('member@example.com', {}, { timeout: 10000 })).toBeInTheDocument();
  expect(screen.getByRole('heading', { name: t('ru', 'profile.step.basics') })).toBeInTheDocument();
  expect(screen.queryByRole('textbox', { name: t('ru', 'profile.equipment') })).not.toBeInTheDocument();
  expect(screen.queryByRole('textbox', { name: t('ru', 'profile.days') })).not.toBeInTheDocument();

  await user.click(screen.getByRole('button', { name: t('ru', 'common.next') }));
  expect(screen.getByRole('heading', { name: t('ru', 'profile.step.goals') })).toBeInTheDocument();
  expect(screen.getByRole('combobox', { name: t('ru', 'profile.duration') })).toBeInTheDocument();
  expect(screen.getByRole('combobox', { name: t('ru', 'profile.timezone') })).toBeInTheDocument();

  await user.click(screen.getByRole('button', { name: t('ru', 'common.next') }));
  expect(screen.getByRole('heading', { name: t('ru', 'profile.step.setup') })).toBeInTheDocument();
  expect(await screen.findByRole('checkbox', { name: 'Гантели' }, { timeout: 10000 })).toBeChecked();
  await user.click(screen.getByRole('checkbox', { name: 'Коврик' }));
  await user.click(screen.getByRole('checkbox', { name: t('ru', 'weekday.thursday') }));
  await user.selectOptions(screen.getByLabelText(`${t('ru', 'weekday.thursday')} ${t('ru', 'plan.duration')}`), '60');

  await user.click(screen.getByRole('button', { name: t('ru', 'common.next') }));
  expect(screen.getByRole('heading', { name: t('ru', 'profile.step.preferences') })).toBeInTheDocument();
  expect(screen.getByRole('combobox', { name: t('ru', 'profile.trainingStyle') })).toBeInTheDocument();
  expect(screen.getByRole('combobox', { name: t('ru', 'profile.mealStyle') })).toBeInTheDocument();
  expect(screen.getByRole('combobox', { name: t('ru', 'profile.hydrationPreference') })).toBeInTheDocument();

  await user.selectOptions(screen.getByRole('combobox', { name: t('ru', 'profile.trainingStyle') }), 'strength_endurance');
  await user.selectOptions(screen.getByRole('combobox', { name: t('ru', 'profile.mealStyle') }), 'family_style');
  await user.selectOptions(screen.getByRole('combobox', { name: t('ru', 'profile.hydrationPreference') }), 'large_bottles');
  await user.click(screen.getByRole('button', { name: t('ru', 'profile.saveOnboarding') }));

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/onboarding',
      expect.objectContaining({
        method: 'PUT',
        credentials: 'include'
      })
    )
  );

  const onboardingCall = fetchMock.mock.calls.find(
    ([input, init]) => String(input).endsWith('/onboarding') && init?.method === 'PUT'
  );
  expect(onboardingCall).toBeTruthy();
  expect(String(onboardingCall?.[1]?.body)).toContain('"preferred_training_style":"strength_endurance"');
  expect(String(onboardingCall?.[1]?.body)).toContain('"preferred_meal_style":"family_style"');
  expect(String(onboardingCall?.[1]?.body)).toContain('"hydration_preference":"large_bottles"');
  expect(String(onboardingCall?.[1]?.body)).toContain('"equipment_ids":["10000000-0000-0000-0000-000000000001","10000000-0000-0000-0000-000000000002"]');
  expect(String(onboardingCall?.[1]?.body)).toContain('"weekday":"tuesday"');
  expect(String(onboardingCall?.[1]?.body)).toContain('"weekday":"thursday"');
  expect(String(onboardingCall?.[1]?.body)).toContain('"duration_min":60');
  expect(useAuthStore.getState().onboardingDone).toBe(true);
});
