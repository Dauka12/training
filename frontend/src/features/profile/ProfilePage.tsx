import { useEffect, useMemo, useState } from 'react';
import { CalendarDays, ChevronLeft, ChevronRight, CircleCheckBig, Droplets, Dumbbell, Sparkles } from 'lucide-react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiRequest } from '../../shared/api/client';
import { useAuthStore } from '../../shared/auth/store';
import { t, type SupportedLocale } from '../../shared/i18n';
import { preferencesStore } from '../../shared/preferences/store';
import { Checkbox, Field, SelectField, TextAreaField, splitComma } from '../../shared/ui/forms';

type NotificationPreferences = {
  hydration_reminder: boolean;
  email_enabled: boolean;
};

type AvailabilityDay = {
  weekday: string;
  duration_min: number;
};

type EquipmentCatalogItem = {
  id: string;
  names?: Record<string, string>;
  descriptions?: Record<string, string>;
  category: string;
  location_type: string;
  active?: boolean;
};

type ProfileFormState = {
  age: string;
  biological_sex: string;
  height_cm: string;
  current_weight_kg: string;
  target_weight_kg: string;
  primary_goal: string;
  program_duration_weeks: string;
  experience_level: string;
  activity_level: string;
  training_location: string;
  timezone: string;
  preferred_training_style: string;
  preferred_meal_style: string;
  hydration_preference: string;
  injuries: string;
  dietary_preferences: string;
  avoid_foods: string;
  equipment_ids: string[];
  available_training_days: AvailabilityDay[];
};

const weekdayOrder = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'] as const;
const durationOptions = [30, 45, 60, 75, 90];
const programDurationOptions = [6, 8, 12, 16, 20, 24];
const timezoneValues = ['Asia/Qyzylorda', 'Asia/Almaty', 'Asia/Aqtobe', 'Asia/Aqtau', 'UTC'];
const onboardingStepIDs = ['basics', 'goals', 'setup', 'preferences'] as const;
type OnboardingStepID = (typeof onboardingStepIDs)[number];

export function ProfilePage({ locale }: { locale: SupportedLocale }) {
  const queryClient = useQueryClient();
  const setAuthenticated = useAuthStore((state) => state.setAuthenticated);
  const authRole = useAuthStore((state) => state.role);
  const meQuery = useQuery({
    queryKey: ['me'],
    queryFn: () =>
      apiRequest<{
        email: string;
        locale: string;
        theme: string;
        onboarding_done: boolean;
        water_target_ml?: number;
        water_override_ml?: number;
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
          available_training_days?: AvailabilityDay[];
        };
      }>('/me')
  });
  const notificationPreferencesQuery = useQuery({
    queryKey: ['notification-preferences'],
    queryFn: () => apiRequest<{ preferences: NotificationPreferences }>('/notifications/preferences')
  });
  const equipmentQuery = useQuery({
    queryKey: ['catalog-equipment'],
    queryFn: () => apiRequest<{ items: EquipmentCatalogItem[] }>('/catalog/equipment')
  });

  const [message, setMessage] = useState('');
  const [activeStep, setActiveStep] = useState(0);
  const [preferences, setPreferences] = useState({ locale, theme: 'light', water_override_ml: '0' });
  const [notifications, setNotifications] = useState<NotificationPreferences>({
    hydration_reminder: true,
    email_enabled: false
  });
  const [profile, setProfile] = useState<ProfileFormState>({
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
    equipment_ids: ['10000000-0000-0000-0000-000000000001'],
    available_training_days: [
      { weekday: 'monday', duration_min: 60 },
      { weekday: 'wednesday', duration_min: 45 },
      { weekday: 'friday', duration_min: 60 }
    ]
  });
  const stepCards = useMemo(
    () => [
      {
        id: 'basics' as OnboardingStepID,
        icon: <Sparkles size={18} />,
        title: t(locale, 'profile.step.basics'),
        body: t(locale, 'profile.step.basicsHint')
      },
      {
        id: 'goals' as OnboardingStepID,
        icon: <CircleCheckBig size={18} />,
        title: t(locale, 'profile.step.goals'),
        body: t(locale, 'profile.step.goalsHint')
      },
      {
        id: 'setup' as OnboardingStepID,
        icon: <CalendarDays size={18} />,
        title: t(locale, 'profile.step.setup'),
        body: t(locale, 'profile.step.setupHint')
      },
      {
        id: 'preferences' as OnboardingStepID,
        icon: <Droplets size={18} />,
        title: t(locale, 'profile.step.preferences'),
        body: t(locale, 'profile.step.preferencesHint')
      }
    ],
    [locale]
  );
  const themeLabel = t(locale, `theme.${preferences.theme}`);
  const selectedDayCount = profile.available_training_days.length;
  const selectedEquipmentCount = profile.equipment_ids.length;
  const activeStepCard = stepCards[activeStep];
  const isLastStep = activeStep === stepCards.length - 1;

  useEffect(() => {
    if (!meQuery.data) {
      return;
    }
    const waterOverride = meQuery.data.water_override_ml ?? meQuery.data.water_target_ml ?? 0;
    setPreferences((current) => ({
      ...current,
      locale: (meQuery.data.locale as SupportedLocale) || locale,
      theme: meQuery.data.theme || 'light',
      water_override_ml: String(waterOverride)
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
        equipment_ids: source.equipment_ids ?? current.equipment_ids,
        available_training_days: normalizeAvailability(source.available_training_days ?? current.available_training_days)
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
          equipment_ids: profile.equipment_ids,
          available_training_days: normalizeAvailability(profile.available_training_days)
        })
      }),
    onSuccess: async () => {
      setAuthenticated({
        role: authRole,
        email: meQuery.data?.email ?? '',
        onboardingDone: true
      });
      setMessage(t(locale, 'common.save'));
      await queryClient.invalidateQueries({ queryKey: ['me'] });
    },
    onError: (error: Error) => setMessage(error.message)
  });

  const sexOptions = ensureCurrentOption(profile.biological_sex, [
    { value: 'male', label: t(locale, 'profile.option.sex.male') },
    { value: 'female', label: t(locale, 'profile.option.sex.female') }
  ]);
  const goalOptions = ensureCurrentOption(profile.primary_goal, [
    { value: 'lose_fat', label: t(locale, 'profile.option.goal.lose_fat') },
    { value: 'gain_muscle', label: t(locale, 'profile.option.goal.gain_muscle') },
    { value: 'maintain', label: t(locale, 'profile.option.goal.maintain') }
  ]);
  const experienceOptions = ensureCurrentOption(profile.experience_level, [
    { value: 'beginner', label: t(locale, 'profile.option.experience.beginner') },
    { value: 'intermediate', label: t(locale, 'profile.option.experience.intermediate') },
    { value: 'advanced', label: t(locale, 'profile.option.experience.advanced') }
  ]);
  const activityOptions = ensureCurrentOption(profile.activity_level, [
    { value: 'light', label: t(locale, 'profile.option.activity.light') },
    { value: 'moderate', label: t(locale, 'profile.option.activity.moderate') },
    { value: 'high', label: t(locale, 'profile.option.activity.high') }
  ]);
  const locationOptions = ensureCurrentOption(profile.training_location, [
    { value: 'home', label: t(locale, 'profile.option.location.home') },
    { value: 'gym', label: t(locale, 'profile.option.location.gym') },
    { value: 'mixed', label: t(locale, 'profile.option.location.mixed') }
  ]);
  const durationSelectOptions = ensureCurrentOption(
    profile.program_duration_weeks,
    programDurationOptions.map((value) => ({
      value: String(value),
      label: `${value} ${t(locale, 'profile.durationUnit')}`
    }))
  );
  const timezoneOptions = ensureCurrentOption(
    profile.timezone,
    timezoneValues.map((value) => ({
      value,
      label: t(locale, `profile.option.timezone.${value}`)
    }))
  );
  const trainingStyleOptions = ensureCurrentOption(profile.preferred_training_style, [
    { value: 'balanced_strength', label: t(locale, 'profile.option.training.balanced_strength') },
    { value: 'strength_endurance', label: t(locale, 'profile.option.training.strength_endurance') },
    { value: 'low_impact_strength', label: t(locale, 'profile.option.training.low_impact_strength') },
    { value: 'mobility_friendly', label: t(locale, 'profile.option.training.mobility_friendly') },
    { value: 'short_sessions', label: t(locale, 'profile.option.training.short_sessions') }
  ]);
  const mealStyleOptions = ensureCurrentOption(profile.preferred_meal_style, [
    { value: 'simple_prep', label: t(locale, 'profile.option.meal.simple_prep') },
    { value: 'family_style', label: t(locale, 'profile.option.meal.family_style') },
    { value: 'high_protein', label: t(locale, 'profile.option.meal.high_protein') },
    { value: 'budget_friendly', label: t(locale, 'profile.option.meal.budget_friendly') },
    { value: 'vegetarian_friendly', label: t(locale, 'profile.option.meal.vegetarian_friendly') }
  ]);
  const hydrationOptions = ensureCurrentOption(profile.hydration_preference, [
    { value: 'regular_small_sips', label: t(locale, 'profile.option.hydration.regular_small_sips') },
    { value: 'large_bottles', label: t(locale, 'profile.option.hydration.large_bottles') },
    { value: 'reminders_needed', label: t(locale, 'profile.option.hydration.reminders_needed') },
    { value: 'around_workouts', label: t(locale, 'profile.option.hydration.around_workouts') },
    { value: 'flexible', label: t(locale, 'profile.option.hydration.flexible') }
  ]);

  function nextStep() {
    setActiveStep((current) => Math.min(current + 1, stepCards.length - 1));
  }

  function previousStep() {
    setActiveStep((current) => Math.max(current - 1, 0));
  }

  return (
    <>
      <section className="card card--panel">
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
        <section className="card card--panel">
          <h2>{t(locale, 'profile.preferences')}</h2>
          <div className="stack">
            <SelectField
              label={t(locale, 'auth.locale')}
              value={preferences.locale}
              onChange={(value) =>
                setPreferences((current) => ({
                  ...current,
                  locale: value as SupportedLocale
                }))
              }
              options={[
                { value: 'ru', label: 'RU' },
                { value: 'kk', label: 'KK' }
              ]}
            />
            <SelectField
              label={t(locale, 'profile.theme')}
              value={preferences.theme}
              onChange={(value) => setPreferences((current) => ({ ...current, theme: value }))}
              options={[
                { value: 'light', label: t(locale, 'theme.light') },
                { value: 'dark', label: t(locale, 'theme.dark') }
              ]}
            />
            <Field
              label={t(locale, 'profile.waterOverride')}
              value={preferences.water_override_ml}
              onChange={(value) => setPreferences((current) => ({ ...current, water_override_ml: value }))}
              type="number"
              inputMode="numeric"
              hint={t(locale, 'profile.hint.waterOverride')}
            />
            <div className="button-row">
              <button type="button" className="button button--primary" onClick={() => savePreferences.mutate()}>
                {t(locale, 'profile.savePreferences')}
              </button>
            </div>
          </div>
        </section>

        <section className="card card--panel">
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

      <section id="onboarding-section" className="card card--panel onboarding-panel">
        <div className="section-header">
          <div>
            <h2>{t(locale, 'profile.onboarding')}</h2>
            <p className="muted">{t(locale, 'profile.onboardingHint')}</p>
          </div>
          <div className="shell-highlights">
            <span className="badge badge--soft">{selectedEquipmentCount}</span>
            <span className="badge badge--soft">{selectedDayCount}</span>
          </div>
        </div>

        <div className="onboarding-layout">
          <aside className="onboarding-steps" aria-label={t(locale, 'profile.onboarding')}>
            {stepCards.map((step, index) => (
              <button
                key={step.id}
                type="button"
                className={`step-card${index === activeStep ? ' step-card--active' : ''}${index < activeStep ? ' step-card--done' : ''}`}
                onClick={() => setActiveStep(index)}
              >
                <span className="step-card__icon">{step.icon}</span>
                <span className="step-card__copy">
                  <strong>{step.title}</strong>
                  <span className="muted">{step.body}</span>
                </span>
              </button>
            ))}
          </aside>

          <section className="form-section onboarding-stage">
            <div className="form-section__header">
              <div className="eyebrow-row">
                <span className="badge badge--soft">
                  {activeStep + 1} / {stepCards.length}
                </span>
                <span className="muted">{activeStepCard.body}</span>
              </div>
              <h3>{activeStepCard.title}</h3>
            </div>

            {activeStepCard.id === 'basics' ? (
              <div className="form-grid">
                <Field label={t(locale, 'profile.age')} value={profile.age} onChange={(value) => setProfile((current) => ({ ...current, age: value }))} type="number" inputMode="numeric" />
                <SelectField label={t(locale, 'profile.sex')} value={profile.biological_sex} onChange={(value) => setProfile((current) => ({ ...current, biological_sex: value }))} options={sexOptions} />
                <Field label={t(locale, 'profile.height')} value={profile.height_cm} onChange={(value) => setProfile((current) => ({ ...current, height_cm: value }))} type="number" inputMode="numeric" />
                <Field label={t(locale, 'profile.weight')} value={profile.current_weight_kg} onChange={(value) => setProfile((current) => ({ ...current, current_weight_kg: value }))} type="number" inputMode="numeric" />
                <Field label={t(locale, 'profile.targetWeight')} value={profile.target_weight_kg} onChange={(value) => setProfile((current) => ({ ...current, target_weight_kg: value }))} type="number" inputMode="numeric" />
              </div>
            ) : null}

            {activeStepCard.id === 'goals' ? (
              <div className="form-grid">
                <SelectField label={t(locale, 'profile.goal')} value={profile.primary_goal} onChange={(value) => setProfile((current) => ({ ...current, primary_goal: value }))} options={goalOptions} />
                <SelectField label={t(locale, 'profile.duration')} value={profile.program_duration_weeks} onChange={(value) => setProfile((current) => ({ ...current, program_duration_weeks: value }))} options={durationSelectOptions} hint={t(locale, 'profile.hint.duration')} />
                <SelectField label={t(locale, 'profile.experience')} value={profile.experience_level} onChange={(value) => setProfile((current) => ({ ...current, experience_level: value }))} options={experienceOptions} />
                <SelectField label={t(locale, 'profile.activity')} value={profile.activity_level} onChange={(value) => setProfile((current) => ({ ...current, activity_level: value }))} options={activityOptions} />
                <SelectField label={t(locale, 'profile.location')} value={profile.training_location} onChange={(value) => setProfile((current) => ({ ...current, training_location: value }))} options={locationOptions} />
                <SelectField label={t(locale, 'profile.timezone')} value={profile.timezone} onChange={(value) => setProfile((current) => ({ ...current, timezone: value }))} options={timezoneOptions} hint={t(locale, 'profile.hint.timezone')} />
              </div>
            ) : null}

            {activeStepCard.id === 'setup' ? (
              <div className="setup-grid">
                <section className="form-subsection">
                  <div className="section-header">
                    <div>
                      <h4>{t(locale, 'profile.section.equipment')}</h4>
                      <p className="muted">{t(locale, 'profile.equipmentHelp')}</p>
                    </div>
                    <span className="badge badge--soft">
                      <Dumbbell size={16} aria-hidden="true" />
                      {selectedEquipmentCount}
                    </span>
                  </div>
                  {equipmentQuery.isLoading ? <p className="muted">{t(locale, 'common.loading')}</p> : null}
                  {equipmentQuery.isError ? <p className="muted">{t(locale, 'common.empty')}</p> : null}
                  <div className="choice-grid">
                    {(equipmentQuery.data?.items ?? []).map((item) => {
                      const title = localizedCatalogText(locale, item.names, item.id);
                      const description = localizedCatalogText(locale, item.descriptions, item.category);
                      const selected = profile.equipment_ids.includes(item.id);

                      return (
                        <label key={item.id} className={`choice-card${selected ? ' choice-card--selected' : ''}`}>
                          <input
                            aria-label={title}
                            type="checkbox"
                            checked={selected}
                            onChange={() =>
                              setProfile((current) => ({
                                ...current,
                                equipment_ids: toggleEquipmentSelection(current.equipment_ids, item.id)
                              }))
                            }
                          />
                          <div className="choice-card__copy">
                            <div className="choice-card__title-row">
                              <strong>{title}</strong>
                              <span className="badge badge--soft">{item.category}</span>
                            </div>
                            <span className="muted">{description}</span>
                          </div>
                        </label>
                      );
                    })}
                  </div>
                </section>

                <section className="form-subsection">
                  <div className="section-header">
                    <div>
                      <h4>{t(locale, 'profile.section.schedule')}</h4>
                      <p className="muted">{t(locale, 'profile.scheduleHelp')}</p>
                    </div>
                    <span className="badge badge--soft">
                      <CalendarDays size={16} aria-hidden="true" />
                      {selectedDayCount}
                    </span>
                  </div>
                  <div className="availability-grid">
                    {weekdayOrder.map((weekday) => {
                      const currentDay = profile.available_training_days.find((item) => item.weekday === weekday);
                      const weekdayLabel = translateWeekday(locale, weekday);

                      return (
                        <div key={weekday} className={`availability-row${currentDay ? ' availability-row--active' : ''}`}>
                          <label className="checkbox availability-row__toggle">
                            <input
                              aria-label={weekdayLabel}
                              type="checkbox"
                              checked={Boolean(currentDay)}
                              onChange={() =>
                                setProfile((current) => ({
                                  ...current,
                                  available_training_days: toggleAvailabilityDay(current.available_training_days, weekday)
                                }))
                              }
                            />
                            <span>{weekdayLabel}</span>
                          </label>
                          <SelectField
                            label={`${weekdayLabel} ${t(locale, 'plan.duration')}`}
                            value={String(currentDay?.duration_min ?? 45)}
                            onChange={(value) =>
                              setProfile((current) => ({
                                ...current,
                                available_training_days: updateAvailabilityDuration(current.available_training_days, weekday, Number(value))
                              }))
                            }
                            disabled={!currentDay}
                            options={durationOptions.map((duration) => ({
                              value: String(duration),
                              label: `${duration} ${t(locale, 'plan.minutes')}`
                            }))}
                          />
                        </div>
                      );
                    })}
                  </div>
                  <p className="muted">{t(locale, 'profile.daysHint')}</p>
                </section>
              </div>
            ) : null}

            {activeStepCard.id === 'preferences' ? (
              <div className="stack">
                <div className="form-grid">
                  <SelectField label={t(locale, 'profile.trainingStyle')} value={profile.preferred_training_style} onChange={(value) => setProfile((current) => ({ ...current, preferred_training_style: value }))} options={trainingStyleOptions} hint={t(locale, 'profile.hint.trainingStyle')} />
                  <SelectField label={t(locale, 'profile.mealStyle')} value={profile.preferred_meal_style} onChange={(value) => setProfile((current) => ({ ...current, preferred_meal_style: value }))} options={mealStyleOptions} hint={t(locale, 'profile.hint.mealStyle')} />
                  <SelectField label={t(locale, 'profile.hydrationPreference')} value={profile.hydration_preference} onChange={(value) => setProfile((current) => ({ ...current, hydration_preference: value }))} options={hydrationOptions} hint={t(locale, 'profile.hint.hydration')} />
                </div>
                <div className="form-grid form-grid--single">
                  <TextAreaField label={t(locale, 'profile.injuries')} value={profile.injuries} onChange={(value) => setProfile((current) => ({ ...current, injuries: value }))} />
                  <TextAreaField label={t(locale, 'profile.preferencesDiet')} value={profile.dietary_preferences} onChange={(value) => setProfile((current) => ({ ...current, dietary_preferences: value }))} />
                  <TextAreaField label={t(locale, 'profile.avoidFoods')} value={profile.avoid_foods} onChange={(value) => setProfile((current) => ({ ...current, avoid_foods: value }))} />
                </div>
              </div>
            ) : null}

            <div className="onboarding-actions">
              <button type="button" className="button button--ghost" onClick={previousStep} disabled={activeStep === 0}>
                <ChevronLeft size={18} aria-hidden="true" />
                <span>{t(locale, 'common.previous')}</span>
              </button>
              {!isLastStep ? (
                <button type="button" className="button button--primary" onClick={nextStep}>
                  <span>{t(locale, 'common.next')}</span>
                  <ChevronRight size={18} aria-hidden="true" />
                </button>
              ) : (
                <button type="button" className="button button--primary" onClick={() => saveOnboarding.mutate()}>
                  <CircleCheckBig size={18} aria-hidden="true" />
                  <span>{t(locale, 'profile.saveOnboarding')}</span>
                </button>
              )}
            </div>
          </section>
        </div>
      </section>
    </>
  );
}

function ensureCurrentOption(value: string, options: Array<{ value: string; label: string }>) {
  if (options.some((option) => option.value === value)) {
    return options;
  }
  return [...options, { value, label: value }];
}

function translateWeekday(locale: SupportedLocale, weekday: string) {
  const key = `weekday.${weekday}`;
  const translated = t(locale, key);
  return translated === key ? weekday : translated;
}

function normalizeAvailability(days: AvailabilityDay[]) {
  return weekdayOrder
    .map((weekday) => days.find((item) => item.weekday === weekday))
    .filter((item): item is AvailabilityDay => Boolean(item))
    .map((item) => ({
      weekday: item.weekday,
      duration_min: item.duration_min || 45
    }));
}

function toggleAvailabilityDay(days: AvailabilityDay[], weekday: string) {
  if (days.some((item) => item.weekday === weekday)) {
    return normalizeAvailability(days.filter((item) => item.weekday !== weekday));
  }
  return normalizeAvailability([...days, { weekday, duration_min: 45 }]);
}

function updateAvailabilityDuration(days: AvailabilityDay[], weekday: string, durationMin: number) {
  return normalizeAvailability(
    days.map((item) =>
      item.weekday === weekday
        ? {
            ...item,
            duration_min: durationMin || 45
          }
        : item
    )
  );
}

function toggleEquipmentSelection(ids: string[], id: string) {
  if (ids.includes(id)) {
    return ids.filter((item) => item !== id);
  }
  return [...ids, id];
}

function localizedCatalogText(locale: SupportedLocale, values: Record<string, string> | undefined, fallback: string) {
  return values?.[locale] ?? values?.ru ?? values?.kk ?? fallback;
}
