import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { NotificationCenter } from '../notifications/NotificationCenter';
import { WaterTracker } from '../tracking/WaterTracker';
import { apiRequest } from '../../shared/api/client';
import { t, type SupportedLocale } from '../../shared/i18n';
import { CardStat, Checkbox, Field, SectionPage, SelectField, TextAreaField } from '../../shared/ui/forms';

export function TodayPage({ locale }: { locale: SupportedLocale }) {
  const dashboard = useQuery({
    queryKey: ['dashboard'],
    queryFn: () =>
      apiRequest<{
        today_workout?: { session_name?: string };
        meal_status: string;
        hydration: { target_ml: number; consumed_ml: number };
        quick_actions: string[];
        current_week_progress: { completed_sessions: number };
        latest_weight_trend: number;
        next_session?: { session_name?: string };
        plan_health: string;
        notifications_unread: number;
      }>('/dashboard/today')
  });
  const notifications = useQuery({
    queryKey: ['notifications'],
    queryFn: () =>
      apiRequest<{ items: Array<{ id: string; title: string; type: string; read: boolean; createdAt: string; target_url?: string }> }>('/notifications')
  });

  if (dashboard.isLoading) {
    return <SectionPage title={t(locale, 'common.loading')} />;
  }

  const planHealth = translateStatus(locale, String(dashboard.data?.plan_health ?? '-'));
  const mealStatus = translateStatus(locale, String(dashboard.data?.meal_status ?? 'not_logged'));
  const workoutName = dashboard.data?.today_workout?.session_name ?? t(locale, 'plan.empty');
  const nextSession = dashboard.data?.next_session?.session_name ?? '-';
  const hydrationProgress = `${dashboard.data?.hydration.consumed_ml ?? 0} / ${dashboard.data?.hydration.target_ml ?? 0} ${t(locale, 'common.unitMl')}`;

  return (
    <>
      <section className="card dashboard-hero">
        <div className="dashboard-hero__copy">
          <p className="eyebrow">{t(locale, 'today.title')}</p>
          <h1>{t(locale, 'today.heroTitle')}</h1>
          <p className="muted">{t(locale, 'today.heroBody')}</p>
        </div>
        <article className="notice notice--accent dashboard-focus">
          <span className="muted">{t(locale, 'today.workoutCard')}</span>
          <strong>{workoutName}</strong>
          <span className="muted">{t(locale, 'today.nextWorkout')}: {nextSession}</span>
        </article>
      </section>

      <section className="dashboard-grid dashboard-grid--today">
        <section className="card">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'today.quickActions')}</h2>
              <p className="muted">{workoutName}</p>
            </div>
            <span className="badge badge--soft">{planHealth}</span>
          </div>
          <div className="stats-grid stats-grid--dense">
            <CardStat title={t(locale, 'today.planHealth')} value={planHealth} />
            <CardStat title={t(locale, 'today.mealStatus')} value={mealStatus} />
            <CardStat title={t(locale, 'today.weekProgress')} value={String(dashboard.data?.current_week_progress.completed_sessions ?? 0)} />
            <CardStat title={t(locale, 'today.weightTrend')} value={String(dashboard.data?.latest_weight_trend ?? '-')} />
            <CardStat title={t(locale, 'today.notifications')} value={String(dashboard.data?.notifications_unread ?? 0)} />
            <CardStat title={t(locale, 'today.secondaryCard')} value={nextSession} />
          </div>
          <div className="button-row button-row--dashboard">
            <Link className="button button--primary" to="/track">{t(locale, 'today.logWorkout')}</Link>
            <Link className="button button--ghost" to="/track">{t(locale, 'today.logMeal')}</Link>
            <Link className="button" to="/profile">{t(locale, 'today.adjustPlan')}</Link>
          </div>
        </section>

        <section className="card">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'today.hydrationCard')}</h2>
              <p className="muted">{t(locale, 'track.water.title')}</p>
            </div>
            <span className="badge badge--soft">{hydrationProgress}</span>
          </div>
          <WaterLogger locale={locale} current={dashboard.data?.hydration.consumed_ml ?? 0} target={dashboard.data?.hydration.target_ml ?? 0} />
        </section>

        <section className="card">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'today.progressCard')}</h2>
              <p className="muted">{t(locale, 'today.notificationCard')}</p>
            </div>
          </div>
          <div className="stack">
            <article className="notice">
              <strong>{t(locale, 'today.planHealth')}</strong>
              <span>{planHealth}</span>
            </article>
            <article className="notice">
              <strong>{t(locale, 'today.mealStatus')}</strong>
              <span>{mealStatus}</span>
            </article>
            <article className="notice">
              <strong>{t(locale, 'today.nextWorkout')}</strong>
              <span>{nextSession}</span>
            </article>
          </div>
        </section>

        <NotificationCenter
          locale={locale}
          items={(notifications.data?.items ?? []).map((item) => ({ ...item, targetURL: item.target_url }))}
        />
      </section>
    </>
  );
}

export function PlanPage({ locale }: { locale: SupportedLocale }) {
  const queryClient = useQueryClient();
  const planQuery = useQuery({
    queryKey: ['active-plan'],
    queryFn: () =>
      apiRequest<{
        plan: {
          title: string;
          summary: string;
          warnings: string[];
          nutrition: {
            daily_calories: number;
            protein_g: number;
            carbs_g: number;
            fat_g: number;
            daily_water_ml: number;
            training_note: string;
            rest_note: string;
            hydration_note: string;
            meal_examples: Array<{ slot: string; examples: string[] }>;
          };
          schedule: Array<{ id: string; session_name: string; weekday: string }>;
          weeks: Array<{
            week_index: number;
            days: Array<{
              weekday: string;
              session_name: string;
              goal: string;
              estimated_minutes: number;
              warmup: string[];
              exercises: Array<{
                order: number;
                exercise_id: string;
                exercise_name: string;
                sets: number;
                reps: string;
                rest_sec: number;
                effort_note: string;
                notes: string;
              }>;
              cooldown: string[];
            }>;
          }>;
          adaptation_rules: string[];
        };
      }>('/plans/active')
  });
  const mutation = useMutation({
    mutationFn: () =>
      apiRequest('/plans/generate', {
        method: 'POST',
        body: JSON.stringify({ generation_type: 'manual' })
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['active-plan'] });
      await queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    }
  });

  return (
    <section className="card">
      <div className="section-header">
        <h1>{t(locale, 'plan.title')}</h1>
        <button type="button" className="button button--primary" onClick={() => mutation.mutate()}>
          {t(locale, 'plan.generate')}
        </button>
      </div>
      {planQuery.isError ? <p className="muted">{t(locale, 'plan.empty')}</p> : null}
      {planQuery.data ? (
        <div className="stack">
          <strong>{planQuery.data.plan.title}</strong>
          <p>{planQuery.data.plan.summary}</p>
          {planQuery.data.plan.warnings.length > 0 ? <h2>{t(locale, 'plan.warnings')}</h2> : null}
          {planQuery.data.plan.warnings.map((warning) => (
            <article key={warning} className="notice">
              <span>{warning}</span>
            </article>
          ))}

          <h2>{t(locale, 'plan.nutrition')}</h2>
          <div className="stats-grid">
            <CardStat title={t(locale, 'plan.calories')} value={String(planQuery.data.plan.nutrition.daily_calories)} />
            <CardStat title={t(locale, 'plan.protein')} value={String(planQuery.data.plan.nutrition.protein_g)} />
            <CardStat title={t(locale, 'plan.carbs')} value={String(planQuery.data.plan.nutrition.carbs_g)} />
            <CardStat title={t(locale, 'plan.fats')} value={String(planQuery.data.plan.nutrition.fat_g)} />
            <CardStat title={t(locale, 'plan.water')} value={String(planQuery.data.plan.nutrition.daily_water_ml)} />
          </div>
          <article className="notice">
            <strong>{t(locale, 'plan.trainingNote')}</strong>
            <span>{planQuery.data.plan.nutrition.training_note}</span>
          </article>
          <article className="notice">
            <strong>{t(locale, 'plan.restNote')}</strong>
            <span>{planQuery.data.plan.nutrition.rest_note}</span>
          </article>
          <article className="notice">
            <strong>{t(locale, 'plan.hydrationNote')}</strong>
            <span>{planQuery.data.plan.nutrition.hydration_note}</span>
          </article>

          {planQuery.data.plan.nutrition.meal_examples.length > 0 ? <h2>{t(locale, 'plan.mealExamples')}</h2> : null}
          {planQuery.data.plan.nutrition.meal_examples.map((item) => (
            <article key={`${item.slot}-${item.examples.join('-')}`} className="notice">
              <strong>{translateMealSlot(locale, item.slot)}</strong>
              <span>{item.examples.join(', ')}</span>
            </article>
          ))}

          {planQuery.data.plan.weeks.length > 0 ? <h2>{t(locale, 'plan.structure')}</h2> : null}
          {planQuery.data.plan.weeks.map((week) => (
            <article key={week.week_index} className="card card--nested">
              <h3>{t(locale, 'plan.week')} {week.week_index}</h3>
              <div className="stack">
                {week.days.map((day) => (
                  <section key={`${week.week_index}-${day.weekday}-${day.session_name}`} className="notice">
                    <strong>{translateWeekday(locale, day.weekday)}: {day.session_name}</strong>
                    <p>{t(locale, 'plan.goal')}: {day.goal}</p>
                    <p>{t(locale, 'plan.duration')}: {day.estimated_minutes} {t(locale, 'plan.minutes')}</p>
                    {day.warmup.length > 0 ? <p>{t(locale, 'plan.warmup')}: {day.warmup.join(', ')}</p> : null}
                    <div className="stack">
                      {day.exercises.map((exercise) => (
                        <article key={`${day.session_name}-${exercise.exercise_id}-${exercise.order}`} className="notice notice--subtle">
                          <strong>{exercise.exercise_name || exercise.exercise_id}</strong>
                          <span>{exercise.sets} x {exercise.reps}</span>
                          {exercise.effort_note ? <span>{exercise.effort_note}</span> : null}
                          {exercise.notes ? <span>{exercise.notes}</span> : null}
                        </article>
                      ))}
                    </div>
                    {day.cooldown.length > 0 ? <p>{t(locale, 'plan.cooldown')}: {day.cooldown.join(', ')}</p> : null}
                  </section>
                ))}
              </div>
            </article>
          ))}

          {planQuery.data.plan.adaptation_rules.length > 0 ? <h2>{t(locale, 'plan.adaptation')}</h2> : null}
          {planQuery.data.plan.adaptation_rules.map((rule) => (
            <article key={rule} className="notice">
              <span>{rule}</span>
            </article>
          ))}
        </div>
      ) : null}
    </section>
  );
}

export function TrackPage({ locale }: { locale: SupportedLocale }) {
  const queryClient = useQueryClient();
  const plan = useQuery({
    queryKey: ['active-plan'],
    queryFn: () =>
      apiRequest<{ plan: { schedule: Array<{ id: string; session_name: string }> } }>('/plans/active')
  });
  const hydrationSummary = useQuery({
    queryKey: ['hydration-summary'],
    queryFn: () =>
      apiRequest<{
        target_ml: number;
        consumed_ml: number;
        adherence: number;
        weekly_target_ml: number;
        weekly_consumed_ml: number;
        weekly_adherence: number;
      }>('/tracking/hydration/summary')
  });
  const [mealStatus, setMealStatus] = useState('followed');
  const [mealNote, setMealNote] = useState('');
  const [selectedScheduleID, setSelectedScheduleID] = useState('');
  const [workoutStatus, setWorkoutStatus] = useState('done');
  const [discomfortFlag, setDiscomfortFlag] = useState(false);
  const [difficulty, setDifficulty] = useState('6');
  const [workoutNote, setWorkoutNote] = useState('');
  const [completionTime, setCompletionTime] = useState(() => toLocalDateTimeValue(new Date()));
  const [message, setMessage] = useState('');

  useEffect(() => {
    if (!selectedScheduleID && plan.data?.plan.schedule?.length) {
      setSelectedScheduleID(plan.data.plan.schedule[0].id);
    }
  }, [plan.data, selectedScheduleID]);

  const mealMutation = useMutation({
    mutationFn: () =>
      apiRequest('/tracking/meals', {
        method: 'POST',
        body: JSON.stringify({ status: mealStatus, note: mealNote })
      }),
    onSuccess: async () => {
      setMessage(t(locale, 'common.save'));
      await queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    }
  });

  const workoutMutation = useMutation({
    mutationFn: () =>
      apiRequest('/tracking/workouts/log', {
        method: 'POST',
        body: JSON.stringify({
          schedule_id: selectedScheduleID,
          status: workoutStatus,
          discomfort_flag: discomfortFlag,
          difficulty: Number(difficulty) || 6,
          note: workoutNote,
          completion_time: toRFC3339(completionTime)
        })
      }),
    onSuccess: async () => {
      setMessage(t(locale, 'common.save'));
      await queryClient.invalidateQueries({ queryKey: ['active-plan'] });
      await queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      await queryClient.invalidateQueries({ queryKey: ['hydration-summary'] });
    }
  });

  const waterMutation = useMutation({
    mutationFn: (amount: number) =>
      apiRequest('/tracking/water', {
        method: 'POST',
        body: JSON.stringify({ amount_ml: amount })
      }),
    onSuccess: async () => {
      setMessage(t(locale, 'common.save'));
      await queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      await queryClient.invalidateQueries({ queryKey: ['hydration-summary'] });
    }
  });

  return (
    <>
      <section className="card">
        <h1>{t(locale, 'track.title')}</h1>
        {message ? <p className="form-message">{message}</p> : null}
      </section>

      <section className="card">
        <div className="section-header">
          <div>
            <h2>{t(locale, 'track.water.title')}</h2>
            <p className="muted">{t(locale, 'track.hydration.weekly')}</p>
          </div>
          <span className="badge badge--soft">
            {(hydrationSummary.data?.weekly_consumed_ml ?? 0)} / {(hydrationSummary.data?.weekly_target_ml ?? 0)} {t(locale, 'common.unitMl')}
          </span>
        </div>
        <WaterTracker
          locale={locale}
          currentML={hydrationSummary.data?.consumed_ml ?? 0}
          targetML={hydrationSummary.data?.target_ml ?? 0}
          onQuickAdd={(amount) => waterMutation.mutate(amount)}
          onCustomAdd={(amount) => waterMutation.mutate(amount)}
        />
      </section>

      <section className="card">
        <h2>{t(locale, 'track.meal.title')}</h2>
        <SelectField
          label={t(locale, 'common.status')}
          value={mealStatus}
          onChange={setMealStatus}
          options={[
            { value: 'followed', label: t(locale, 'track.meal.followed') },
            { value: 'partially_followed', label: t(locale, 'track.meal.partial') },
            { value: 'off_plan', label: t(locale, 'track.meal.off') }
          ]}
        />
        <TextAreaField label={t(locale, 'common.note')} value={mealNote} onChange={setMealNote} placeholder={t(locale, 'common.note')} />
        <div className="button-row">
          <button type="button" className="button button--primary" onClick={() => mealMutation.mutate()}>
            {t(locale, 'common.save')}
          </button>
        </div>
      </section>

      <section className="card">
        <h2>{t(locale, 'track.workout.title')}</h2>
        <div className="form-grid">
          <SelectField
            label={t(locale, 'track.workout.session')}
            value={selectedScheduleID}
            onChange={setSelectedScheduleID}
            options={(plan.data?.plan.schedule ?? []).map((item) => ({
              value: item.id,
              label: item.session_name
            }))}
          />
          <SelectField
            label={t(locale, 'track.workout.status')}
            value={workoutStatus}
            onChange={setWorkoutStatus}
            options={[
              { value: 'done', label: t(locale, 'track.workout.done') },
              { value: 'partially_done', label: t(locale, 'track.workout.partial') },
              { value: 'skipped', label: t(locale, 'track.workout.skipped') },
              { value: 'rescheduled', label: t(locale, 'track.workout.rescheduled') }
            ]}
          />
          <Field label={t(locale, 'track.workout.difficulty')} value={difficulty} onChange={setDifficulty} />
          <label className="field">
            <span>{t(locale, 'track.workout.completedAt')}</span>
            <input
              aria-label={t(locale, 'track.workout.completedAt')}
              type="datetime-local"
              value={completionTime}
              onChange={(event) => setCompletionTime(event.target.value)}
            />
          </label>
        </div>
        <div className="stack">
          <Checkbox label={t(locale, 'track.workout.discomfort')} checked={discomfortFlag} onChange={setDiscomfortFlag} />
          <TextAreaField label={t(locale, 'track.workout.note')} value={workoutNote} onChange={setWorkoutNote} placeholder={t(locale, 'track.workout.note')} />
        </div>
        <div className="button-row">
          <button type="button" className="button button--primary" onClick={() => workoutMutation.mutate()} disabled={!selectedScheduleID}>
            {t(locale, 'common.save')}
          </button>
        </div>
      </section>
    </>
  );
}

export function ProgressPage({ locale }: { locale: SupportedLocale }) {
  const [weight, setWeight] = useState('80');
  const [energy, setEnergy] = useState('3');
  const [availabilityChanged, setAvailabilityChanged] = useState(false);
  const [equipmentChanged, setEquipmentChanged] = useState(false);
  const [injuryChanged, setInjuryChanged] = useState(false);
  const [message, setMessage] = useState('');

  const mutation = useMutation({
    mutationFn: () =>
      apiRequest('/checkins/weekly', {
        method: 'POST',
        body: JSON.stringify({
          weight_kg: Number(weight),
          energy_level: Number(energy),
          availability_changed: availabilityChanged,
          equipment_changed: equipmentChanged,
          injury_changed: injuryChanged,
          note: ''
        })
      }),
    onSuccess: () => setMessage(t(locale, 'common.save')),
    onError: (error: Error) => setMessage(error.message)
  });

  return (
    <section className="card">
      <h1>{t(locale, 'progress.title')}</h1>
      <div className="stack">
        <Field label={t(locale, 'progress.weight')} value={weight} onChange={setWeight} />
        <Field label={t(locale, 'progress.energy')} value={energy} onChange={setEnergy} />
        <Checkbox label={t(locale, 'progress.availabilityChanged')} checked={availabilityChanged} onChange={setAvailabilityChanged} />
        <Checkbox label={t(locale, 'progress.equipmentChanged')} checked={equipmentChanged} onChange={setEquipmentChanged} />
        <Checkbox label={t(locale, 'progress.injuryChanged')} checked={injuryChanged} onChange={setInjuryChanged} />
        {message ? <p className="form-message">{message}</p> : null}
        <button type="button" className="button button--primary" onClick={() => mutation.mutate()}>
          {t(locale, 'progress.submit')}
        </button>
      </div>
    </section>
  );
}

function WaterLogger({ locale, current, target }: { locale: SupportedLocale; current: number; target: number }) {
  const queryClient = useQueryClient();
  const mutation = useMutation({
    mutationFn: (amount: number) =>
      apiRequest('/tracking/water', {
        method: 'POST',
        body: JSON.stringify({ amount_ml: amount })
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    }
  });

  return (
    <WaterTracker
      locale={locale}
      currentML={current}
      targetML={target}
      onQuickAdd={(amount) => mutation.mutate(amount)}
      onCustomAdd={(amount) => mutation.mutate(amount)}
    />
  );
}

function translateMealSlot(locale: SupportedLocale, slot: string) {
  const normalized = slot.trim().toLowerCase();
  const key = `plan.slot.${normalized}`;
  const translated = t(locale, key);
  return translated === key ? slot : translated;
}

function translateWeekday(locale: SupportedLocale, weekday: string) {
  const normalized = weekday.trim().toLowerCase();
  const key = `weekday.${normalized}`;
  const translated = t(locale, key);
  return translated === key ? weekday : translated;
}

function translateStatus(locale: SupportedLocale, value: string) {
  const normalized = value.trim().toLowerCase();
  const key = `status.${normalized}`;
  const translated = t(locale, key);
  return translated === key ? value : translated;
}

function toLocalDateTimeValue(date: Date) {
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 16);
}

function toRFC3339(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return new Date().toISOString();
  }
  return parsed.toISOString();
}
