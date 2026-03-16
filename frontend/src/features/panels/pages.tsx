import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiRequest } from '../../shared/api/client';
import { t, type SupportedLocale } from '../../shared/i18n';
import { EmptyState, Field, SelectField, TextAreaField, splitComma } from '../../shared/ui/forms';

type TrainerUser = { email: string; plan_health: string; workouts: number };
type TrainerUserDetail = {
  email: string;
  plan_versions: number;
  meal_logs: number;
  water_ml: number;
  weekly_checkins?: number;
  assigned_trainer?: string;
};
type TrainerNote = { id: string; trainer_email: string; user_email: string; body: string; created_at: string };
type AdminExercise = { id: string; names?: Record<string, string>; slug: string };
type ExerciseDetail = {
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

export function TrainerPage({ locale }: { locale: SupportedLocale }) {
  const queryClient = useQueryClient();
  const [selectedEmail, setSelectedEmail] = useState('');
  const [noteBody, setNoteBody] = useState('');

  const users = useQuery({
    queryKey: ['trainer-users'],
    queryFn: () => apiRequest<{ items: TrainerUser[] }>('/trainer/users')
  });

  const userDetail = useQuery({
    queryKey: ['trainer-user-detail', selectedEmail],
    enabled: Boolean(selectedEmail),
    queryFn: () => apiRequest<{ user: TrainerUserDetail }>(`/trainer/users/${selectedEmail}`)
  });

  const notes = useQuery({
    queryKey: ['trainer-user-notes', selectedEmail],
    enabled: Boolean(selectedEmail),
    queryFn: () => apiRequest<{ items: TrainerNote[] }>(`/trainer/users/${selectedEmail}/notes`)
  });

  const regenerate = useMutation({
    mutationFn: () => apiRequest(`/trainer/users/${selectedEmail}/regenerate-plan`, { method: 'POST' }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['trainer-users'] });
      await queryClient.invalidateQueries({ queryKey: ['trainer-user-detail', selectedEmail] });
    }
  });

  const addNote = useMutation({
    mutationFn: () =>
      apiRequest(`/trainer/users/${selectedEmail}/notes`, {
        method: 'POST',
        body: JSON.stringify({ body: noteBody })
      }),
    onSuccess: async () => {
      setNoteBody('');
      await queryClient.invalidateQueries({ queryKey: ['trainer-user-notes', selectedEmail] });
    }
  });

  return (
    <div className="page-stack page-stack--ops">
      <section className="card card--panel ops-hero">
        <div className="section-header">
          <div>
            <h1>{t(locale, 'trainer.title')}</h1>
            <p className="muted">{t(locale, 'shell.subtitle')}</p>
          </div>
        </div>
        <div className="stats-grid stats-grid--dense">
          <article className="stat">
            <span className="muted">{t(locale, 'admin.users')}</span>
            <strong>{users.data?.items?.length ?? 0}</strong>
          </article>
          <article className="stat">
            <span className="muted">{t(locale, 'trainer.notes')}</span>
            <strong>{notes.data?.items?.length ?? 0}</strong>
          </article>
          <article className="stat">
            <span className="muted">{t(locale, 'trainer.selectedUser')}</span>
            <strong>{selectedEmail || '-'}</strong>
          </article>
        </div>
      </section>

      <section className="ops-grid ops-grid--trainer">
        <section className="card card--panel ops-panel">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'admin.users')}</h2>
              <p className="muted">{t(locale, 'shell.quickNavigation')}</p>
            </div>
          </div>
          <div className="stack">
            {(users.data?.items ?? []).map((item) => (
              <article key={item.email} className={`notice notice--subtle ops-list-item${selectedEmail === item.email ? ' ops-list-item--active' : ''}`}>
                <div className="section-header">
                  <div>
                    <strong>{item.email}</strong>
                    <p className="muted">{translateStatus(locale, item.plan_health)}</p>
                  </div>
                  <span className="badge badge--soft">{item.workouts}</span>
                </div>
                <button type="button" className="button button--ghost" onClick={() => setSelectedEmail(item.email)}>
                  {t(locale, 'trainer.details')}
                </button>
              </article>
            ))}
            {!users.data?.items?.length ? <EmptyState locale={locale} /> : null}
          </div>
        </section>

        <section className="card card--panel ops-panel ops-panel--detail">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'trainer.selectedUser')}</h2>
              <p className="muted">{selectedEmail || t(locale, 'common.empty')}</p>
            </div>
            <button type="button" className="button button--primary" onClick={() => regenerate.mutate()} disabled={!selectedEmail}>
              {t(locale, 'trainer.regenerate')}
            </button>
          </div>
          {selectedEmail && userDetail.data?.user ? (
            <div className="stats-grid">
              <article className="stat">
                <span className="muted">{t(locale, 'plan.title')}</span>
                <strong>{userDetail.data.user.plan_versions}</strong>
              </article>
              <article className="stat">
                <span className="muted">{t(locale, 'track.meal.title')}</span>
                <strong>{userDetail.data.user.meal_logs}</strong>
              </article>
              <article className="stat">
                <span className="muted">{t(locale, 'track.water.title')}</span>
                <strong>{userDetail.data.user.water_ml}</strong>
              </article>
              <article className="stat">
                <span className="muted">{t(locale, 'trainer.checkins')}</span>
                <strong>{userDetail.data.user.weekly_checkins ?? 0}</strong>
              </article>
              <article className="stat">
                <span className="muted">{t(locale, 'trainer.assignedTrainer')}</span>
                <strong>{userDetail.data.user.assigned_trainer || '-'}</strong>
              </article>
            </div>
          ) : (
            <EmptyState locale={locale} />
          )}
        </section>

        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'trainer.notes')}</h2>
          <label className="field">
            <span>{t(locale, 'common.note')}</span>
            <textarea value={noteBody} onChange={(event) => setNoteBody(event.target.value)} />
          </label>
          <div className="button-row">
            <button type="button" className="button button--primary" onClick={() => addNote.mutate()} disabled={!selectedEmail || !noteBody.trim()}>
              {t(locale, 'trainer.addNote')}
            </button>
          </div>
          <div className="stack">
            {(notes.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.trainer_email}</strong>
                <span>{item.body}</span>
                <span className="muted">{item.created_at}</span>
              </article>
            ))}
            {!notes.data?.items?.length ? <EmptyState locale={locale} /> : null}
          </div>
        </section>
      </section>
    </div>
  );
}

export function AdminPage({ locale }: { locale: SupportedLocale }) {
  const queryClient = useQueryClient();
  const [equipmentName, setEquipmentName] = useState('');
  const [exerciseName, setExerciseName] = useState('');
  const [selectedExerciseID, setSelectedExerciseID] = useState('');
  const [exerciseMediaURL, setExerciseMediaURL] = useState('');
  const [exerciseTechnique, setExerciseTechnique] = useState('');
  const [exerciseContraindications, setExerciseContraindications] = useState('');
  const [exerciseSubstitutions, setExerciseSubstitutions] = useState('');
  const [memberEmail, setMemberEmail] = useState('');
  const [trainerEmail, setTrainerEmail] = useState('trainer@local.test');
  const [supportStatuses, setSupportStatuses] = useState<Record<string, string>>({});
  const [supportAssignees, setSupportAssignees] = useState<Record<string, string>>({});
  const [discussionStatuses, setDiscussionStatuses] = useState<Record<string, string>>({});
  const [previewData, setPreviewData] = useState<{
    counts: { equipment: number; exercises: number };
    equipment: Array<{ source_id?: string; name?: string }>;
    exercises: Array<{ source_id?: string; name_en?: string; slug?: string; media_url?: string }>;
  } | null>(null);

  const adminUsers = useQuery({
    queryKey: ['admin-users'],
    queryFn: () =>
      apiRequest<{
        items: Array<{
          email: string;
          roles: string[];
          assigned_trainer_email?: string;
          onboarding_done: boolean;
          active_plan_versions: number;
        }>;
      }>('/admin/users')
  });

  const trainers = useQuery({
    queryKey: ['admin-trainers'],
    queryFn: () =>
      apiRequest<{ items: Array<{ email: string; assigned_users: number; trainer_note_count: number }> }>('/admin/trainers')
  });

  const trainerOptions = [
    { value: '', label: t(locale, 'admin.selectTrainer') },
    ...(trainers.data?.items ?? []).map((item) => ({ value: item.email, label: item.email }))
  ];

  const supportThreads = useQuery({
    queryKey: ['admin-support-threads'],
    queryFn: () =>
      apiRequest<{ items: Array<{ id: string; user_email: string; status: string; assigned_to_email?: string; message_count: number }> }>(
        '/admin/support/threads'
      )
  });

  const discussions = useQuery({
    queryKey: ['admin-discussion-threads'],
    queryFn: () =>
      apiRequest<{ items: Array<{ id: string; author_email: string; title: string; status: string; reply_count: number }> }>(
        '/admin/discussions/threads'
      )
  });

  const notificationLogs = useQuery({
    queryKey: ['admin-notification-logs'],
    queryFn: () =>
      apiRequest<{ items: Array<{ id: string; user_email: string; type: string; title: string }> }>('/admin/logs/notifications')
  });

  const equipment = useQuery({
    queryKey: ['admin-equipment'],
    queryFn: () =>
      apiRequest<{ items: Array<{ id: string; names?: Record<string, string>; category: string }> }>('/admin/catalog/equipment')
  });

  const exercises = useQuery({
    queryKey: ['admin-exercises'],
    queryFn: () =>
      apiRequest<{ items: AdminExercise[] }>('/admin/catalog/exercises')
  });

  const exerciseDetail = useQuery({
    queryKey: ['admin-exercise-detail', selectedExerciseID],
    enabled: Boolean(selectedExerciseID),
    queryFn: () => apiRequest<{ exercise: ExerciseDetail }>(`/catalog/exercises/${selectedExerciseID}`)
  });

  const aiLogs = useQuery({
    queryKey: ['admin-ai-logs'],
    queryFn: () =>
      apiRequest<{ items: Array<{ id: string; status: string; provider: string; plan_title?: string }> }>('/admin/logs/ai')
  });

  const emailLogs = useQuery({
    queryKey: ['admin-email-logs'],
    queryFn: () => apiRequest<{ items: Array<{ id: string; subject: string; status: string }> }>('/admin/logs/email')
  });

  const auditLogs = useQuery({
    queryKey: ['admin-audit-logs'],
    queryFn: () => apiRequest<{ items: Array<{ id: string; action: string; actor_email: string }> }>('/admin/logs/audit')
  });

  const addEquipment = useMutation({
    mutationFn: () =>
      apiRequest('/admin/catalog/equipment', {
        method: 'POST',
        body: JSON.stringify({
          names: { ru: equipmentName, kk: equipmentName },
          descriptions: { ru: equipmentName, kk: equipmentName },
          category: 'custom',
          location_type: 'mixed',
          media_url: '',
          active: true
        })
      }),
    onSuccess: async () => {
      setEquipmentName('');
      await queryClient.invalidateQueries({ queryKey: ['admin-equipment'] });
    }
  });

  const addExercise = useMutation({
    mutationFn: () =>
      apiRequest('/admin/catalog/exercises', {
        method: 'POST',
        body: JSON.stringify({
          slug: exerciseName.toLowerCase().replace(/\s+/g, '-'),
          names: { ru: exerciseName, kk: exerciseName },
          descriptions: { ru: exerciseName, kk: exerciseName },
          technique: { ru: exerciseName, kk: exerciseName },
          movement_pattern: 'custom',
          difficulty: 'beginner',
          location_type: 'mixed',
          equipment_ids: [],
          active: true
        })
      }),
    onSuccess: async () => {
      setExerciseName('');
      await queryClient.invalidateQueries({ queryKey: ['admin-exercises'] });
    }
  });

  const assignTrainer = useMutation({
    mutationFn: () =>
      apiRequest('/admin/trainers/assign', {
        method: 'POST',
        body: JSON.stringify({
          user_email: memberEmail,
          trainer_email: trainerEmail
        })
      }),
    onSuccess: async () => {
      setMemberEmail('');
      await queryClient.invalidateQueries({ queryKey: ['admin-audit-logs'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-users'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-trainers'] });
    }
  });

  const moderateSupport = useMutation({
    mutationFn: ({ threadID, status, assignedToEmail }: { threadID: string; status: string; assignedToEmail: string }) =>
      apiRequest(`/admin/support/threads/${threadID}/status`, {
        method: 'POST',
        body: JSON.stringify({
          status,
          assigned_to_email: assignedToEmail
        })
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['admin-support-threads'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-audit-logs'] });
    }
  });

  const moderateDiscussion = useMutation({
    mutationFn: ({ threadID, status }: { threadID: string; status: string }) =>
      apiRequest(`/admin/discussions/threads/${threadID}/moderation`, {
        method: 'POST',
        body: JSON.stringify({ status })
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['admin-discussion-threads'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-audit-logs'] });
    }
  });

  const importCatalog = useMutation({
    mutationFn: () =>
      apiRequest('/admin/catalog/import/wger', {
        method: 'POST',
        body: JSON.stringify({ limit: 12 })
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['admin-equipment'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-exercises'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-audit-logs'] });
    }
  });

  const previewCatalog = useMutation({
    mutationFn: () =>
      apiRequest<{
        preview: {
          counts: { equipment: number; exercises: number };
          equipment: Array<{ source_id?: string; name?: string }>;
          exercises: Array<{ source_id?: string; name_en?: string; slug?: string; media_url?: string }>;
        };
      }>('/admin/catalog/import/wger/preview', {
        method: 'POST',
        body: JSON.stringify({ limit: 6 })
      }),
    onSuccess: (response) => setPreviewData(response.preview)
  });

  const saveExerciseMeta = useMutation({
    mutationFn: () =>
      apiRequest(`/admin/catalog/exercises/${selectedExerciseID}`, {
        method: 'PUT',
        body: JSON.stringify({
          media_url: exerciseMediaURL,
          technique: { ru: exerciseTechnique },
          contraindication_tags: splitComma(exerciseContraindications),
          substitution_ids: splitComma(exerciseSubstitutions)
        })
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['admin-exercises'] });
      await queryClient.invalidateQueries({ queryKey: ['admin-exercise-detail', selectedExerciseID] });
      await queryClient.invalidateQueries({ queryKey: ['admin-audit-logs'] });
    }
  });

  useEffect(() => {
    if (!selectedExerciseID && exercises.data?.items?.length) {
      setSelectedExerciseID(exercises.data.items[0].id);
    }
  }, [exercises.data, selectedExerciseID]);

  useEffect(() => {
    if (!exerciseDetail.data?.exercise) {
      return;
    }
    setExerciseMediaURL(exerciseDetail.data.exercise.media_url ?? '');
    setExerciseTechnique(exerciseDetail.data.exercise.technique ?? '');
    setExerciseContraindications(exerciseDetail.data.exercise.contraindication_tags.join(', '));
    setExerciseSubstitutions(exerciseDetail.data.exercise.substitutions.map((item) => item.id).join(', '));
  }, [exerciseDetail.data]);

  return (
    <div className="page-stack page-stack--ops">
      <section className="card card--panel ops-hero">
        <h1>{t(locale, 'admin.title')}</h1>
        <p className="muted">{t(locale, 'admin.opsOverview')}</p>
        <div className="stats-grid">
          <article className="stat">
            <span className="muted">{t(locale, 'admin.users')}</span>
            <strong>{adminUsers.data?.items?.length ?? 0}</strong>
          </article>
          <article className="stat">
            <span className="muted">{t(locale, 'admin.trainers')}</span>
            <strong>{trainers.data?.items?.length ?? 0}</strong>
          </article>
          <article className="stat">
            <span className="muted">{t(locale, 'admin.supportThreads')}</span>
            <strong>{supportThreads.data?.items?.length ?? 0}</strong>
          </article>
          <article className="stat">
            <span className="muted">{t(locale, 'admin.discussions')}</span>
            <strong>{discussions.data?.items?.length ?? 0}</strong>
          </article>
        </div>
      </section>

      <section className="ops-grid ops-grid--admin">
        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.assignment')}</h2>
          <div className="form-grid">
            <SelectField
              label={t(locale, 'admin.userEmail')}
              value={memberEmail}
              onChange={setMemberEmail}
              options={[
                { value: '', label: t(locale, 'admin.selectUser') },
                ...(adminUsers.data?.items ?? [])
                  .filter((item) => item.roles.includes('user'))
                  .map((item) => ({ value: item.email, label: item.email }))
              ]}
            />
            <SelectField
              label={t(locale, 'admin.trainerEmail')}
              value={trainerEmail}
              onChange={setTrainerEmail}
              options={[
                { value: '', label: t(locale, 'admin.selectTrainer') },
                ...(trainers.data?.items ?? []).map((item) => ({ value: item.email, label: item.email }))
              ]}
            />
          </div>
          <div className="button-row">
            <button
              type="button"
              className="button button--primary"
              onClick={() => assignTrainer.mutate()}
              disabled={!memberEmail || !trainerEmail}
            >
              {t(locale, 'admin.assignTrainer')}
            </button>
          </div>
        </section>

        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.users')}</h2>
          <div className="stack">
            {(adminUsers.data?.items ?? []).map((item) => (
              <article key={item.email} className="notice">
                <strong>{item.email}</strong>
                <span className="muted">{item.roles.join(', ')}</span>
                <span className="muted">
                  {t(locale, 'admin.activePlans')}: {item.active_plan_versions}
                </span>
                <span className="muted">{item.assigned_trainer_email || '-'}</span>
              </article>
            ))}
            {!adminUsers.data?.items?.length ? <EmptyState locale={locale} /> : null}
          </div>
        </section>

        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.trainers')}</h2>
          <div className="stack">
            {(trainers.data?.items ?? []).map((item) => (
              <article key={item.email} className="notice">
                <strong>{item.email}</strong>
                <span className="muted">
                  {t(locale, 'admin.assignedUsers')}: {item.assigned_users}
                </span>
                <span className="muted">
                  {t(locale, 'trainer.notes')}: {item.trainer_note_count}
                </span>
              </article>
            ))}
            {!trainers.data?.items?.length ? <EmptyState locale={locale} /> : null}
          </div>
        </section>
      </section>

      <section className="ops-grid ops-grid--admin">
        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.equipment')}</h2>
          <p className="muted">{t(locale, 'admin.importHint')}</p>
          <div className="button-row">
            <button type="button" className="button" onClick={() => previewCatalog.mutate()}>
              {t(locale, 'admin.previewWger')}
            </button>
            <button type="button" className="button button--ghost" onClick={() => importCatalog.mutate()}>
              {t(locale, 'admin.importWger')}
            </button>
          </div>
          {previewData ? (
            <section className="notice notice--subtle">
              <strong>{t(locale, 'admin.previewSamples')}</strong>
              <div className="stats-grid stats-grid--dense">
                <article className="stat">
                  <span className="muted">{t(locale, 'admin.equipment')}</span>
                  <strong>{previewData.counts.equipment}</strong>
                </article>
                <article className="stat">
                  <span className="muted">{t(locale, 'admin.exercises')}</span>
                  <strong>{previewData.counts.exercises}</strong>
                </article>
              </div>
              <div className="stack">
                {previewData.exercises.map((item) => (
                  <article key={item.source_id ?? item.slug} className="notice notice--subtle">
                    <strong>{item.name_en ?? item.slug}</strong>
                    <span className="muted">{item.media_url || '-'}</span>
                  </article>
                ))}
              </div>
            </section>
          ) : null}
          <Field label={t(locale, 'admin.addEquipment')} value={equipmentName} onChange={setEquipmentName} />
          <div className="button-row">
            <button
              type="button"
              className="button button--primary"
              onClick={() => addEquipment.mutate()}
              disabled={!equipmentName.trim()}
            >
              {t(locale, 'admin.addEquipment')}
            </button>
          </div>
          <div className="stack">
            {(equipment.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.names?.ru ?? item.id}</strong>
                <span className="muted">{item.category}</span>
              </article>
            ))}
          </div>
        </section>

        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.exercises')}</h2>
          <Field label={t(locale, 'admin.addExercise')} value={exerciseName} onChange={setExerciseName} />
          <div className="button-row">
            <button
              type="button"
              className="button button--primary"
              onClick={() => addExercise.mutate()}
              disabled={!exerciseName.trim()}
            >
              {t(locale, 'admin.addExercise')}
            </button>
          </div>
          <div className="stack">
            {(exercises.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.names?.ru ?? item.slug}</strong>
                <span className="muted">{item.slug}</span>
              </article>
            ))}
          </div>

          <section className="ops-editor">
            <SelectField
              label={t(locale, 'admin.exerciseEditor')}
              value={selectedExerciseID}
              onChange={setSelectedExerciseID}
              options={[
                { value: '', label: t(locale, 'common.empty') },
                ...(exercises.data?.items ?? []).map((item) => ({
                  value: item.id,
                  label: item.names?.ru ?? item.slug
                }))
              ]}
            />
            <Field label={t(locale, 'admin.exerciseMedia')} value={exerciseMediaURL} onChange={setExerciseMediaURL} />
            <TextAreaField label={t(locale, 'admin.exerciseTechnique')} value={exerciseTechnique} onChange={setExerciseTechnique} />
            <Field
              label={t(locale, 'admin.exerciseContraindications')}
              value={exerciseContraindications}
              onChange={setExerciseContraindications}
            />
            <Field
              label={t(locale, 'admin.exerciseSubstitutions')}
              value={exerciseSubstitutions}
              onChange={setExerciseSubstitutions}
            />
            <div className="button-row">
              <button
                type="button"
                className="button button--primary"
                onClick={() => saveExerciseMeta.mutate()}
                disabled={!selectedExerciseID}
              >
                {t(locale, 'admin.saveExerciseMeta')}
              </button>
            </div>
            {exerciseDetail.data?.exercise ? (
              <article className="notice notice--subtle">
                <strong>{exerciseDetail.data.exercise.name}</strong>
                <span className="muted">{exerciseDetail.data.exercise.description}</span>
              </article>
            ) : null}
          </section>
        </section>
      </section>

      <section className="ops-grid ops-grid--admin">
        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.supportThreads')}</h2>
          <div className="stack">
            {(supportThreads.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.user_email}</strong>
                <span className="muted">{translateStatus(locale, item.status)}</span>
                <span className="muted">
                  {t(locale, 'admin.messages')}: {item.message_count}
                </span>
                <span className="muted">{item.assigned_to_email || '-'}</span>
                <div className="form-grid form-grid--compact">
                  <SelectField
                    label={t(locale, 'admin.supportStatus')}
                    value={supportStatuses[item.id] ?? item.status}
                    onChange={(value) => setSupportStatuses((current) => ({ ...current, [item.id]: value }))}
                    options={[
                      { value: 'open', label: translateStatus(locale, 'open') },
                      { value: 'in_progress', label: translateStatus(locale, 'in_progress') },
                      { value: 'resolved', label: translateStatus(locale, 'resolved') },
                      { value: 'closed', label: translateStatus(locale, 'closed') }
                    ]}
                  />
                  <SelectField
                    label={t(locale, 'admin.assignee')}
                    value={supportAssignees[item.id] ?? item.assigned_to_email ?? ''}
                    onChange={(value) => setSupportAssignees((current) => ({ ...current, [item.id]: value }))}
                    options={trainerOptions}
                  />
                </div>
                <button
                  type="button"
                  className="button"
                  onClick={() =>
                    moderateSupport.mutate({
                      threadID: item.id,
                      status: supportStatuses[item.id] ?? item.status,
                      assignedToEmail: supportAssignees[item.id] ?? item.assigned_to_email ?? ''
                    })
                  }
                >
                  {t(locale, 'admin.saveSupport')}
                </button>
              </article>
            ))}
            {!supportThreads.data?.items?.length ? <EmptyState locale={locale} /> : null}
          </div>
        </section>

        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.discussions')}</h2>
          <div className="stack">
            {(discussions.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.title}</strong>
                <span className="muted">{item.author_email}</span>
                <span className="muted">
                  {t(locale, 'admin.replies')}: {item.reply_count}
                </span>
                <span className="muted">{translateStatus(locale, item.status)}</span>
                <SelectField
                  label={t(locale, 'admin.discussionStatus')}
                  value={discussionStatuses[item.id] ?? item.status}
                  onChange={(value) => setDiscussionStatuses((current) => ({ ...current, [item.id]: value }))}
                  options={[
                    { value: 'visible', label: translateStatus(locale, 'visible') },
                    { value: 'hidden', label: translateStatus(locale, 'hidden') },
                    { value: 'flagged', label: translateStatus(locale, 'flagged') },
                    { value: 'archived', label: translateStatus(locale, 'archived') }
                  ]}
                />
                <button
                  type="button"
                  className="button"
                  onClick={() =>
                    moderateDiscussion.mutate({
                      threadID: item.id,
                      status: discussionStatuses[item.id] ?? item.status
                    })
                  }
                >
                  {t(locale, 'admin.saveDiscussion')}
                </button>
              </article>
            ))}
            {!discussions.data?.items?.length ? <EmptyState locale={locale} /> : null}
          </div>
        </section>
      </section>

      <section className="ops-grid ops-grid--admin">
        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.notificationLogs')}</h2>
          <div className="stack">
            {(notificationLogs.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.title}</strong>
                <span className="muted">{item.user_email}</span>
                <span className="muted">{item.type}</span>
              </article>
            ))}
            {!notificationLogs.data?.items?.length ? <EmptyState locale={locale} /> : null}
          </div>
        </section>

        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.aiLogs')}</h2>
          <div className="stack">
            {(aiLogs.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.plan_title ?? item.id}</strong>
                <span className="muted">{item.status}</span>
                <span className="muted">{item.provider}</span>
              </article>
            ))}
          </div>
        </section>

        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.emailLogs')}</h2>
          <div className="stack">
            {(emailLogs.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.subject}</strong>
                <span className="muted">{item.status}</span>
              </article>
            ))}
          </div>
        </section>

        <section className="card card--panel ops-panel">
          <h2>{t(locale, 'admin.auditLogs')}</h2>
          <div className="stack">
            {(auditLogs.data?.items ?? []).map((item) => (
              <article key={item.id} className="notice">
                <strong>{item.action}</strong>
                <span className="muted">{item.actor_email}</span>
              </article>
            ))}
          </div>
        </section>
      </section>
    </div>
  );
}

function translateStatus(locale: SupportedLocale, value: string) {
  const translated = t(locale, `status.${value}`);
  return translated === `status.${value}` ? value : translated;
}
