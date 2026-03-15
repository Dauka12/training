import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiRequest } from '../../shared/api/client';
import { t, type SupportedLocale } from '../../shared/i18n';
import { preferencesStore } from '../../shared/preferences/store';
import { Checkbox, Field, parseAvailability, splitComma, TextAreaField } from '../../shared/ui/forms';

type NotificationPreferences = {
  hydration_reminder: boolean;
  email_enabled: boolean;
};

export function ProfilePage({ locale }: { locale: SupportedLocale }) {
  const queryClient = useQueryClient();
  const meQuery = useQuery({
    queryKey: ['me'],
    queryFn: () =>
      apiRequest<{
        email: string;
        locale: string;
        theme: string;
        onboarding_done: boolean;
        profile?: {
          age?: number;
          biological_sex?: string;
          height_cm?: number;
          current_weight_kg?: number;
          target_weight_kg?: number;
          primary_goal?: string;
          program_duration_weeks?: number;
          experience_level?: string;
          activity_level?: string;
          training_location?: string;
          timezone?: string;
          preferred_training_style?: string;
          preferred_meal_style?: string;
          hydration_preference?: string;
          injuries?: string[];
          dietary_preferences?: string[];
          avoid_foods?: string[];
          equipment_ids?: string[];
          available_training_days?: Array<{ weekday: string; duration_min: number }>;
        };
      }>('/me')
  });
  const notificationPreferencesQuery = useQuery({
    queryKey: ['notification-preferences'],
    queryFn: () => apiRequest<{ preferences: NotificationPreferences }>('/notifications/preferences')
  });

  const [message, setMessage] = useState('');
  const [preferences, setPreferences] = useState({ locale, theme: 'light', water_override_ml: '0' });
  const [notifications, setNotifications] = useState<NotificationPreferences>({
    hydration_reminder: true,
    email_enabled: false
  });
  const [profile, setProfile] = useState({
    age: '28',
    biological_sex: 'male',
    height_cm: '180',
    current_weight_kg: '86',
    target_weight_kg: '78',
    primary_goal: 'lose_fat',
    program_duration_weeks: '12',
    experience_level: 'beginner',
    activity_level: 'light',
    training_location: 'mixed',
    timezone: 'Asia/Qyzylorda',
    preferred_training_style: 'balanced_strength',
    preferred_meal_style: 'simple_prep',
    hydration_preference: 'regular_small_sips',
    injuries: '',
    dietary_preferences: '',
    avoid_foods: '',
    equipment_ids: '10000000-0000-0000-0000-000000000001',
    available_training_days: 'monday:60,wednesday:45,friday:60'
  });
  const themeLabel = t(locale, `theme.${preferences.theme}`);

  useEffect(() => {
    if (!meQuery.data) {
      return;
    }
    setPreferences((current) => ({
      ...current,
      locale: (meQuery.data.locale as SupportedLocale) || locale,
      theme: meQuery.data.theme || 'light'
    }));
    if (meQuery.data.profile) {
      const source = meQuery.data.profile;
      setProfile((current) => ({
        ...current,
        age: String(source.age ?? current.age),
        biological_sex: source.biological_sex ?? current.biological_sex,
        height_cm: String(source.height_cm ?? current.height_cm),
        current_weight_kg: String(source.current_weight_kg ?? current.current_weight_kg),
        target_weight_kg: String(source.target_weight_kg ?? current.target_weight_kg),
        primary_goal: source.primary_goal ?? current.primary_goal,
        program_duration_weeks: String(source.program_duration_weeks ?? current.program_duration_weeks),
        experience_level: source.experience_level ?? current.experience_level,
        activity_level: source.activity_level ?? current.activity_level,
        training_location: source.training_location ?? current.training_location,
        timezone: source.timezone ?? current.timezone,
        preferred_training_style: source.preferred_training_style ?? current.preferred_training_style,
        preferred_meal_style: source.preferred_meal_style ?? current.preferred_meal_style,
        hydration_preference: source.hydration_preference ?? current.hydration_preference,
        injuries: (source.injuries ?? []).join(', '),
        dietary_preferences: (source.dietary_preferences ?? []).join(', '),
        avoid_foods: (source.avoid_foods ?? []).join(', '),
        equipment_ids: (source.equipment_ids ?? []).join(', '),
        available_training_days:
          source.available_training_days?.map((item) => `${item.weekday}:${item.duration_min}`).join(', ') ||
          current.available_training_days
      }));
    }
  }, [locale, meQuery.data]);

  useEffect(() => {
    if (!notificationPreferencesQuery.data) {
      return;
    }
    setNotifications(notificationPreferencesQuery.data.preferences);
  }, [notificationPreferencesQuery.data]);

  const savePreferences = useMutation({
    mutationFn: () =>
      apiRequest('/me/preferences', {
        method: 'PUT',
        body: JSON.stringify({
          locale: preferences.locale,
          theme: preferences.theme,
          water_override_ml: Number(preferences.water_override_ml)
        })
      }),
    onSuccess: async () => {
      preferencesStore.getState().hydrateFromServer({
        locale: preferences.locale as 'ru' | 'kk',
        theme: preferences.theme as 'light' | 'dark'
      });
      setMessage(t(locale, 'common.save'));
      await queryClient.invalidateQueries({ queryKey: ['me'] });
    },
    onError: (error: Error) => setMessage(error.message)
  });

  const saveNotifications = useMutation({
    mutationFn: () =>
      apiRequest('/notifications/preferences', {
        method: 'PUT',
        body: JSON.stringify(notifications)
      }),
    onSuccess: async () => {
      setMessage(t(locale, 'common.save'));
      await queryClient.invalidateQueries({ queryKey: ['notification-preferences'] });
    },
    onError: (error: Error) => setMessage(error.message)
  });

  const saveOnboarding = useMutation({
    mutationFn: () =>
      apiRequest('/onboarding', {
        method: 'PUT',
        body: JSON.stringify({
          age: Number(profile.age),
          biological_sex: profile.biological_sex,
          height_cm: Number(profile.height_cm),
          current_weight_kg: Number(profile.current_weight_kg),
          target_weight_kg: Number(profile.target_weight_kg),
          primary_goal: profile.primary_goal,
          program_duration_weeks: Number(profile.program_duration_weeks),
          experience_level: profile.experience_level,
          activity_level: profile.activity_level,
          training_location: profile.training_location,
          timezone: profile.timezone,
          preferred_training_style: profile.preferred_training_style,
          preferred_meal_style: profile.preferred_meal_style,
          hydration_preference: profile.hydration_preference,
          injuries: splitComma(profile.injuries),
          dietary_preferences: splitComma(profile.dietary_preferences),
          avoid_foods: splitComma(profile.avoid_foods),
          equipment_ids: splitComma(profile.equipment_ids),
          available_training_days: parseAvailability(profile.available_training_days)
        })
      }),
    onSuccess: async () => {
      setMessage(t(locale, 'common.save'));
      await queryClient.invalidateQueries({ queryKey: ['me'] });
    },
    onError: (error: Error) => setMessage(error.message)
  });

  return (
    <>
      <section className="card">
        <div className="section-header">
          <div>
            <h1>{t(locale, 'profile.title')}</h1>
            <p className="muted">{meQuery.data?.email ?? ''}</p>
          </div>
          <div className="shell-highlights">
            <span className="badge badge--soft">{preferences.locale.toUpperCase()}</span>
            <span className="badge badge--soft">{themeLabel}</span>
            <span className="badge">{meQuery.data?.onboarding_done ? t(locale, 'status.healthy') : t(locale, 'profile.onboarding')}</span>
          </div>
        </div>
        <p className="muted">{t(locale, 'profile.summaryBody')}</p>
        {message ? <p className="form-message">{message}</p> : null}
      </section>

      <section className="dashboard-grid dashboard-grid--profile">
        <section className="card">
          <h2>{t(locale, 'profile.preferences')}</h2>
          <div className="stack">
            <label className="field">
              <span>{t(locale, 'auth.locale')}</span>
              <select
                value={preferences.locale}
                onChange={(event) =>
                  setPreferences((current) => ({
                    ...current,
                    locale: event.target.value as SupportedLocale
                  }))
                }
              >
                <option value="ru">RU</option>
                <option value="kk">KK</option>
              </select>
            </label>
            <label className="field">
              <span>{t(locale, 'profile.theme')}</span>
              <select value={preferences.theme} onChange={(event) => setPreferences((current) => ({ ...current, theme: event.target.value }))}>
                <option value="light">{t(locale, 'theme.light')}</option>
                <option value="dark">{t(locale, 'theme.dark')}</option>
              </select>
            </label>
            <Field
              label={t(locale, 'profile.waterOverride')}
              value={preferences.water_override_ml}
              onChange={(value) => setPreferences((current) => ({ ...current, water_override_ml: value }))}
            />
            <div className="button-row">
              <button type="button" className="button button--primary" onClick={() => savePreferences.mutate()}>
                {t(locale, 'profile.savePreferences')}
              </button>
            </div>
          </div>
        </section>

        <section className="card">
          <h2>{t(locale, 'profile.notifications')}</h2>
          <div className="stack">
            <Checkbox
              label={t(locale, 'profile.notificationsHydration')}
              checked={notifications.hydration_reminder}
              onChange={(checked) => setNotifications((current) => ({ ...current, hydration_reminder: checked }))}
            />
            <Checkbox
              label={t(locale, 'profile.notificationsEmail')}
              checked={notifications.email_enabled}
              onChange={(checked) => setNotifications((current) => ({ ...current, email_enabled: checked }))}
            />
            <div className="button-row">
              <button type="button" className="button button--primary" onClick={() => saveNotifications.mutate()}>
                {t(locale, 'profile.saveNotifications')}
              </button>
            </div>
          </div>
        </section>
      </section>

      <section className="card">
        <h2>{t(locale, 'profile.onboarding')}</h2>
        <p className="muted">{t(locale, 'profile.onboardingHint')}</p>
        <div className="form-grid">
          <Field label={t(locale, 'profile.age')} value={profile.age} onChange={(value) => setProfile((current) => ({ ...current, age: value }))} />
          <Field label={t(locale, 'profile.sex')} value={profile.biological_sex} onChange={(value) => setProfile((current) => ({ ...current, biological_sex: value }))} />
          <Field label={t(locale, 'profile.height')} value={profile.height_cm} onChange={(value) => setProfile((current) => ({ ...current, height_cm: value }))} />
          <Field label={t(locale, 'profile.weight')} value={profile.current_weight_kg} onChange={(value) => setProfile((current) => ({ ...current, current_weight_kg: value }))} />
          <Field label={t(locale, 'profile.targetWeight')} value={profile.target_weight_kg} onChange={(value) => setProfile((current) => ({ ...current, target_weight_kg: value }))} />
          <Field label={t(locale, 'profile.goal')} value={profile.primary_goal} onChange={(value) => setProfile((current) => ({ ...current, primary_goal: value }))} />
          <Field label={t(locale, 'profile.duration')} value={profile.program_duration_weeks} onChange={(value) => setProfile((current) => ({ ...current, program_duration_weeks: value }))} />
          <Field label={t(locale, 'profile.experience')} value={profile.experience_level} onChange={(value) => setProfile((current) => ({ ...current, experience_level: value }))} />
          <Field label={t(locale, 'profile.activity')} value={profile.activity_level} onChange={(value) => setProfile((current) => ({ ...current, activity_level: value }))} />
          <Field label={t(locale, 'profile.location')} value={profile.training_location} onChange={(value) => setProfile((current) => ({ ...current, training_location: value }))} />
          <Field label={t(locale, 'profile.timezone')} value={profile.timezone} onChange={(value) => setProfile((current) => ({ ...current, timezone: value }))} />
          <Field label={t(locale, 'profile.trainingStyle')} value={profile.preferred_training_style} onChange={(value) => setProfile((current) => ({ ...current, preferred_training_style: value }))} />
          <Field label={t(locale, 'profile.mealStyle')} value={profile.preferred_meal_style} onChange={(value) => setProfile((current) => ({ ...current, preferred_meal_style: value }))} />
          <Field label={t(locale, 'profile.hydrationPreference')} value={profile.hydration_preference} onChange={(value) => setProfile((current) => ({ ...current, hydration_preference: value }))} />
        </div>
        <div className="stack">
          <TextAreaField label={t(locale, 'profile.injuries')} value={profile.injuries} onChange={(value) => setProfile((current) => ({ ...current, injuries: value }))} />
          <TextAreaField label={t(locale, 'profile.preferencesDiet')} value={profile.dietary_preferences} onChange={(value) => setProfile((current) => ({ ...current, dietary_preferences: value }))} />
          <TextAreaField label={t(locale, 'profile.avoidFoods')} value={profile.avoid_foods} onChange={(value) => setProfile((current) => ({ ...current, avoid_foods: value }))} />
          <TextAreaField label={t(locale, 'profile.equipment')} value={profile.equipment_ids} onChange={(value) => setProfile((current) => ({ ...current, equipment_ids: value }))} />
          <TextAreaField label={t(locale, 'profile.days')} value={profile.available_training_days} onChange={(value) => setProfile((current) => ({ ...current, available_training_days: value }))} />
        </div>
        <p className="muted">{t(locale, 'profile.daysHint')}</p>
        <div className="button-row">
          <button type="button" className="button button--primary" onClick={() => saveOnboarding.mutate()}>
            {t(locale, 'profile.saveOnboarding')}
          </button>
        </div>
      </section>
    </>
  );
}
