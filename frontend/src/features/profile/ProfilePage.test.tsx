import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';
import { ProfilePage } from './ProfilePage';
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
  expect(screen.getByDisplayValue('3100')).toBeInTheDocument();

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

test('loads existing onboarding preferences and saves extended planning fields', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/me')) {
      return Response.json({
        email: 'member@example.com',
        locale: 'ru',
        theme: 'dark',
        onboarding_done: true,
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

  expect(await screen.findByDisplayValue('31', {}, { timeout: 10000 })).toBeInTheDocument();
  expect(screen.getByDisplayValue('low_impact_strength')).toBeInTheDocument();
  expect(screen.getByDisplayValue('simple_prep')).toBeInTheDocument();
  expect(screen.getByDisplayValue('small_frequent_sips')).toBeInTheDocument();

  await user.clear(screen.getByLabelText(t('ru', 'profile.trainingStyle')));
  await user.type(screen.getByLabelText(t('ru', 'profile.trainingStyle')), 'strength_endurance');
  await user.clear(screen.getByLabelText(t('ru', 'profile.mealStyle')));
  await user.type(screen.getByLabelText(t('ru', 'profile.mealStyle')), 'family_style');
  await user.clear(screen.getByLabelText(t('ru', 'profile.hydrationPreference')));
  await user.type(screen.getByLabelText(t('ru', 'profile.hydrationPreference')), 'large_bottles');
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
});
