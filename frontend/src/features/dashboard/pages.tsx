import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';
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

      <section className="stats-grid stats-grid--hero">
        <CardStat title={t(locale, 'today.planHealth')} value={planHealth} />
        <CardStat title={t(locale, 'today.hydrationCard')} value={hydrationProgress} />
        <CardStat title={t(locale, 'today.notifications')} value={String(dashboard.data?.notifications_unread ?? 0)} />
      </section>

      <section className="dashboard-grid dashboard-grid--today">
        <section className="card card--panel">
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

        <section className="card card--panel">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'today.hydrationCard')}</h2>
              <p className="muted">{t(locale, 'track.water.title')}</p>
            </div>
            <span className="badge badge--soft">{hydrationProgress}</span>
          </div>
          <WaterLogger locale={locale} current={dashboard.data?.hydration.consumed_ml ?? 0} target={dashboard.data?.hydration.target_ml ?? 0} />
        </section>

        <section className="card card--panel">
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
  const plan = planQuery.data?.plan;

  return (
    <div className="page-stack page-stack--plan">
      <section className="card page-intro">
        <div className="section-header">
          <div>
            <h1>{t(locale, 'plan.title')}</h1>
            <p className="muted">{t(locale, 'landing.dashboard.body')}</p>
          </div>
          <button type="button" className="button button--primary" onClick={() => mutation.mutate()}>
            {t(locale, 'plan.generate')}
          </button>
        </div>
      </section>
      {planQuery.isError ? <section className="card card--panel"><p className="muted">{t(locale, 'plan.empty')}</p></section> : null}
      {plan ? (
        <section className="plan-layout" data-testid="plan-layout">
          <aside className="plan-layout__aside" data-testid="plan-overview">
            <section className="card card--panel plan-panel">
              <div className="stack">
                <span className="badge badge--soft">{t(locale, 'plan.title')}</span>
                <strong>{plan.title}</strong>
                <p className="muted">{plan.summary}</p>
              </div>
            </section>

            <section className="card card--panel plan-panel">
              <div className="section-header">
                <h2>{t(locale, 'plan.nutrition')}</h2>
                <span className="badge badge--soft">{plan.nutrition.daily_water_ml} {t(locale, 'common.unitMl')}</span>
              </div>
              <div className="stats-grid stats-grid--dense">
                <CardStat title={t(locale, 'plan.calories')} value={String(plan.nutrition.daily_calories)} />
                <CardStat title={t(locale, 'plan.protein')} value={String(plan.nutrition.protein_g)} />
                <CardStat title={t(locale, 'plan.carbs')} value={String(plan.nutrition.carbs_g)} />
                <CardStat title={t(locale, 'plan.fats')} value={String(plan.nutrition.fat_g)} />
              </div>
              <article className="notice">
                <strong>{t(locale, 'plan.trainingNote')}</strong>
                <span>{plan.nutrition.training_note}</span>
              </article>
              <article className="notice">
                <strong>{t(locale, 'plan.restNote')}</strong>
                <span>{plan.nutrition.rest_note}</span>
              </article>
              <article className="notice">
                <strong>{t(locale, 'plan.hydrationNote')}</strong>
                <span>{plan.nutrition.hydration_note}</span>
              </article>
            </section>

            {plan.schedule.length > 0 ? (
              <section className="card card--panel plan-panel">
                <h2>{t(locale, 'today.nextWorkout')}</h2>
                <div className="stack">
                  {plan.schedule.map((session) => (
                    <article key={session.id} className="notice notice--subtle session-preview">
                      <strong>{session.session_name}</strong>
                      <span className="muted">{translateWeekday(locale, session.weekday)}</span>
                    </article>
                  ))}
                </div>
              </section>
            ) : null}

            {(plan.warnings.length > 0 || plan.adaptation_rules.length > 0) ? (
              <section className="card card--panel plan-panel">
                {plan.warnings.length > 0 ? <h2>{t(locale, 'plan.warnings')}</h2> : null}
                {plan.warnings.map((warning) => (
                  <article key={warning} className="notice">
                    <span>{warning}</span>
                  </article>
                ))}
                {plan.adaptation_rules.length > 0 ? <h2>{t(locale, 'plan.adaptation')}</h2> : null}
                {plan.adaptation_rules.map((rule) => (
                  <article key={rule} className="notice notice--subtle">
                    <span>{rule}</span>
                  </article>
                ))}
              </section>
            ) : null}
          </aside>

          <div className="plan-layout__main" data-testid="plan-feed">
            {plan.weeks.length > 0 ? (
              <section className="card card--panel plan-panel">
                <div className="section-header">
                  <div>
                    <h2>{t(locale, 'plan.structure')}</h2>
                    <p className="muted">{plan.summary}</p>
                  </div>
                </div>
                <div className="stack">
                  {plan.weeks.map((week) => (
                    <article key={week.week_index} className="card card--nested week-card">
                      <h3>{t(locale, 'plan.week')} {week.week_index}</h3>
                      <div className="stack">
                        {week.days.map((day) => (
                          <section key={`${week.week_index}-${day.weekday}-${day.session_name}`} className="notice week-card__day">
                            <div className="section-header">
                              <div>
                                <strong>{translateWeekday(locale, day.weekday)}: {day.session_name}</strong>
                                <p className="muted">{t(locale, 'plan.goal')}: {day.goal}</p>
                              </div>
                              <span className="badge badge--soft">
                                {day.estimated_minutes} {t(locale, 'plan.minutes')}
                              </span>
                            </div>
                            {day.warmup.length > 0 ? <p>{t(locale, 'plan.warmup')}: {day.warmup.join(', ')}</p> : null}
                            <div className="exercise-list">
                              {day.exercises.map((exercise) => (
                                <article key={`${day.session_name}-${exercise.exercise_id}-${exercise.order}`} className="notice notice--subtle exercise-card">
                                  <div className="section-header">
                                    <strong>
                                      <Link to={`/exercise/${exercise.exercise_id}`}>{exercise.exercise_name || exercise.exercise_id}</Link>
                                    </strong>
                                    <span className="badge badge--soft">{exercise.sets} x {exercise.reps}</span>
                                  </div>
                                  {exercise.effort_note ? <span>{exercise.effort_note}</span> : null}
                                  {exercise.notes ? <span className="muted">{exercise.notes}</span> : null}
                                </article>
                              ))}
                            </div>
                            {day.cooldown.length > 0 ? <p>{t(locale, 'plan.cooldown')}: {day.cooldown.join(', ')}</p> : null}
                          </section>
                        ))}
                      </div>
                    </article>
                  ))}
                </div>
              </section>
            ) : null}

            {plan.nutrition.meal_examples.length > 0 ? (
              <section className="card card--panel plan-panel">
                <h2>{t(locale, 'plan.mealExamples')}</h2>
                <div className="stack">
                  {plan.nutrition.meal_examples.map((item) => (
                    <article key={`${item.slot}-${item.examples.join('-')}`} className="notice">
                      <strong>{translateMealSlot(locale, item.slot)}</strong>
                      <span>{item.examples.join(', ')}</span>
                    </article>
                  ))}
                </div>
              </section>
            ) : null}
          </div>
        </section>
      ) : null}
    </div>
  );
}

export function ExercisePage({ locale }: { locale: SupportedLocale }) {
  const { exerciseID = '' } = useParams();
  const exerciseQuery = useQuery({
    queryKey: ['exercise-detail', exerciseID],
    enabled: Boolean(exerciseID),
    queryFn: () =>
      apiRequest<{
        exercise: {
          id: string;
          slug: string;
          name: string;
          description: string;
          technique: string;
          movement_pattern: string;
          difficulty: string;
          location_type: string;
          media_url: string;
          contraindication_tags: string[];
          equipment: Array<{ id: string; name: string }>;
          substitutions: Array<{ id: string; name: string }>;
        };
      }>(`/catalog/exercises/${exerciseID}`)
  });

  if (exerciseQuery.isLoading) {
    return <SectionPage title={t(locale, 'common.loading')} />;
  }

  const exercise = exerciseQuery.data?.exercise;
  if (!exercise) {
    return <SectionPage title={t(locale, 'common.empty')} />;
  }

  return (
    <div className="page-stack page-stack--exercise">
      <section className="card page-intro">
        <div className="section-header">
          <div>
            <h1>{exercise.name}</h1>
            <p className="muted">{exercise.description}</p>
          </div>
          <div className="shell-highlights">
            <span className="badge badge--soft">{exercise.movement_pattern}</span>
            <span className="badge badge--soft">{exercise.difficulty}</span>
          </div>
        </div>
      </section>

      <section className="plan-layout">
        <aside className="plan-layout__aside">
          <section className="card card--panel plan-panel">
            {exercise.media_url ? <img className="exercise-media" src={exercise.media_url} alt={exercise.name} /> : null}
            <div className="stats-grid stats-grid--dense">
              <CardStat title={t(locale, 'exercise.location')} value={exercise.location_type} />
              <CardStat title={t(locale, 'exercise.equipment')} value={String(exercise.equipment.length)} />
            </div>
          </section>
        </aside>

        <div className="plan-layout__main">
          <section className="card card--panel plan-panel">
            <h2>{t(locale, 'exercise.technique')}</h2>
            <p>{exercise.technique}</p>
          </section>

          {exercise.contraindication_tags.length > 0 ? (
            <section className="card card--panel plan-panel">
              <h2>{t(locale, 'exercise.contraindications')}</h2>
              <div className="shell-highlights">
                {exercise.contraindication_tags.map((tag) => (
                  <span key={tag} className="badge badge--soft">{tag}</span>
                ))}
              </div>
            </section>
          ) : null}

          {exercise.equipment.length > 0 ? (
            <section className="card card--panel plan-panel">
              <h2>{t(locale, 'exercise.equipment')}</h2>
              <div className="stack">
                {exercise.equipment.map((item) => (
                  <article key={item.id} className="notice notice--subtle">
                    <strong>{item.name}</strong>
                  </article>
                ))}
              </div>
            </section>
          ) : null}

          {exercise.substitutions.length > 0 ? (
            <section className="card card--panel plan-panel">
              <h2>{t(locale, 'exercise.substitutions')}</h2>
              <div className="stack">
                {exercise.substitutions.map((item) => (
                  <article key={item.id} className="notice notice--subtle">
                    <Link to={`/exercise/${item.id}`}>{item.name}</Link>
                  </article>
                ))}
              </div>
            </section>
          ) : null}
        </div>
      </section>
    </div>
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
    <div className="page-stack page-stack--track">
      <section className="card page-intro">
        <h1>{t(locale, 'track.title')}</h1>
        {message ? <p className="form-message">{message}</p> : null}
      </section>

      <section className="tracker-layout" data-testid="tracker-layout">
        <div className="tracker-layout__primary" data-testid="tracker-primary">
        <section className="card card--panel tracker-card">
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

        <section className="card card--panel tracker-card">
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
        </div>

        <aside className="tracker-layout__secondary" data-testid="tracker-secondary">
        <section className="card card--panel tracker-card">
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

        <section className="card card--panel tracker-card">
          <h2>{t(locale, 'today.progressCard')}</h2>
          <div className="stats-grid stats-grid--dense">
            <CardStat title={t(locale, 'track.water.title')} value={`${hydrationSummary.data?.consumed_ml ?? 0} / ${hydrationSummary.data?.target_ml ?? 0} ${t(locale, 'common.unitMl')}`} />
            <CardStat title={t(locale, 'today.weekProgress')} value={String(plan.data?.plan.schedule.length ?? 0)} />
            <CardStat title={t(locale, 'today.planHealth')} value={`${Math.round((hydrationSummary.data?.weekly_adherence ?? 0) * 100)}%`} />
          </div>
          <div className="stack">
            {(plan.data?.plan.schedule ?? []).map((item) => (
              <article key={item.id} className="notice notice--subtle session-preview">
                <strong>{item.session_name}</strong>
              </article>
            ))}
          </div>
        </section>
        </aside>
      </section>
    </div>
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
    <div className="page-stack page-stack--progress">
      <section className="card page-intro">
        <h1>{t(locale, 'progress.title')}</h1>
        <p className="muted">{t(locale, 'today.heroBody')}</p>
      </section>
      <section className="progress-layout">
        <section className="card card--panel">
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

        <aside className="card card--panel">
          <div className="stack">
            <article className="notice">
              <strong>{t(locale, 'progress.weight')}</strong>
              <span className="muted">Следите за реальным изменением веса раз в неделю, без ежедневной паники.</span>
            </article>
            <article className="notice">
              <strong>{t(locale, 'progress.energy')}</strong>
              <span className="muted">Энергия помогает понять, стоит ли адаптировать нагрузку уже сейчас.</span>
            </article>
            <article className="notice">
              <strong>{t(locale, 'today.planHealth')}</strong>
              <span className="muted">Если изменились дни, оборудование или ограничения, лучше сразу зафиксировать это здесь.</span>
            </article>
          </div>
        </aside>
      </section>
    </div>
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
