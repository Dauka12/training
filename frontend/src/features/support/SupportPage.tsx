import { useMemo, useState } from 'react';
import { LifeBuoy, MessageSquareText, Newspaper, Sparkles } from 'lucide-react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiRequest } from '../../shared/api/client';
import { t, type SupportedLocale } from '../../shared/i18n';
import { Field, SelectField, TextAreaField } from '../../shared/ui/forms';

type SupportThread = { id: string; title: string; status: string };
type DiscussionThread = { id: string; title: string; category: string };

export function SupportPage({ locale }: { locale: SupportedLocale }) {
  const queryClient = useQueryClient();
  const [threadTitle, setThreadTitle] = useState('');
  const [threadBody, setThreadBody] = useState('');
  const [discussionTitle, setDiscussionTitle] = useState('');
  const [discussionBody, setDiscussionBody] = useState('');
  const [category, setCategory] = useState('general');
  const [feedTab, setFeedTab] = useState<'private' | 'community'>('private');
  const [supportReplies, setSupportReplies] = useState<Record<string, string>>({});
  const [discussionReplies, setDiscussionReplies] = useState<Record<string, string>>({});

  const supportQuery = useQuery({
    queryKey: ['support-threads'],
    queryFn: () => apiRequest<{ items: SupportThread[] }>('/support/threads')
  });
  const discussionQuery = useQuery({
    queryKey: ['discussion-threads'],
    queryFn: () => apiRequest<{ items: DiscussionThread[] }>('/discussions/threads')
  });

  const supportMutation = useMutation({
    mutationFn: () =>
      apiRequest('/support/threads', {
        method: 'POST',
        body: JSON.stringify({ title: threadTitle, body: threadBody })
      }),
    onSuccess: async () => {
      setThreadTitle('');
      setThreadBody('');
      await queryClient.invalidateQueries({ queryKey: ['support-threads'] });
    }
  });

  const discussionMutation = useMutation({
    mutationFn: () =>
      apiRequest('/discussions/threads', {
        method: 'POST',
        body: JSON.stringify({ title: discussionTitle, body: discussionBody, category })
      }),
    onSuccess: async () => {
      setDiscussionTitle('');
      setDiscussionBody('');
      await queryClient.invalidateQueries({ queryKey: ['discussion-threads'] });
    }
  });

  const supportReplyMutation = useMutation({
    mutationFn: ({ threadID, body }: { threadID: string; body: string }) =>
      apiRequest(`/support/threads/${threadID}/messages`, {
        method: 'POST',
        body: JSON.stringify({ body })
      }),
    onSuccess: async (_data, variables) => {
      setSupportReplies((current) => ({ ...current, [variables.threadID]: '' }));
      await queryClient.invalidateQueries({ queryKey: ['support-threads'] });
    }
  });

  const discussionReplyMutation = useMutation({
    mutationFn: ({ threadID, body }: { threadID: string; body: string }) =>
      apiRequest(`/discussions/threads/${threadID}/replies`, {
        method: 'POST',
        body: JSON.stringify({ body })
      }),
    onSuccess: async (_data, variables) => {
      setDiscussionReplies((current) => ({ ...current, [variables.threadID]: '' }));
      await queryClient.invalidateQueries({ queryKey: ['discussion-threads'] });
    }
  });

  const categoryOptions = useMemo(
    () => [
      { value: 'general', label: t(locale, 'support.category.general') },
      { value: 'nutrition', label: t(locale, 'support.category.nutrition') },
      { value: 'workouts', label: t(locale, 'support.category.workouts') },
      { value: 'motivation', label: t(locale, 'support.category.motivation') },
      { value: 'equipment', label: t(locale, 'support.category.equipment') }
    ],
    [locale]
  );

  return (
    <div className="page-stack page-stack--support">
      <section className="card card--panel support-hero">
        <div className="section-header">
          <div>
            <h1>{t(locale, 'support.feedTitle')}</h1>
            <p className="muted">{t(locale, 'support.feedBody')}</p>
          </div>
          <div className="shell-highlights">
            <span className="badge badge--soft">{supportQuery.data?.items?.length ?? 0}</span>
            <span className="badge badge--soft">{discussionQuery.data?.items?.length ?? 0}</span>
          </div>
        </div>
      </section>

      <section className="support-layout">
        <section className="card card--panel support-composer">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'support.composeTitle')}</h2>
              <p className="muted">{t(locale, 'support.subtitle')}</p>
            </div>
          </div>
          <div className="segmented">
            <button
              type="button"
              className={`segmented__item${feedTab === 'private' ? ' segmented__item--active' : ''}`}
              onClick={() => setFeedTab('private')}
            >
              <LifeBuoy size={16} aria-hidden="true" />
              <span>{t(locale, 'support.feed.privateTab')}</span>
            </button>
            <button
              type="button"
              className={`segmented__item${feedTab === 'community' ? ' segmented__item--active' : ''}`}
              onClick={() => setFeedTab('community')}
            >
              <Newspaper size={16} aria-hidden="true" />
              <span>{t(locale, 'support.feed.communityTab')}</span>
            </button>
          </div>

          {feedTab === 'private' ? (
            <div className="stack">
              <Field label={t(locale, 'support.threadTitle')} value={threadTitle} onChange={setThreadTitle} />
              <TextAreaField label={t(locale, 'support.threadBody')} value={threadBody} onChange={setThreadBody} />
              <button
                type="button"
                className="button button--primary"
                onClick={() => supportMutation.mutate()}
                disabled={!threadTitle.trim() || !threadBody.trim()}
              >
                {t(locale, 'support.threadCreate')}
              </button>
            </div>
          ) : (
            <div className="stack">
              <Field label={t(locale, 'support.threadTitle')} value={discussionTitle} onChange={setDiscussionTitle} />
              <SelectField label={t(locale, 'support.category')} value={category} onChange={setCategory} options={categoryOptions} />
              <TextAreaField label={t(locale, 'support.threadBody')} value={discussionBody} onChange={setDiscussionBody} />
              <button
                type="button"
                className="button button--primary"
                onClick={() => discussionMutation.mutate()}
                disabled={!discussionTitle.trim() || !discussionBody.trim()}
              >
                {t(locale, 'support.discussionCreate')}
              </button>
            </div>
          )}
        </section>

        <section className="card card--panel support-feed">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'support.timelineTitle')}</h2>
              <p className="muted">
                {feedTab === 'private' ? t(locale, 'support.privateHint') : t(locale, 'support.publicHint')}
              </p>
            </div>
          </div>
          <div className="stack">
            {feedTab === 'private'
              ? (supportQuery.data?.items ?? []).map((thread) => (
                  <article key={thread.id} className="feed-card">
                    <div className="feed-card__header">
                      <div>
                        <strong>{thread.title}</strong>
                        <p className="muted">{translateStatus(locale, thread.status)}</p>
                      </div>
                      <span className="badge badge--soft">{t(locale, 'support.feed.privateTab')}</span>
                    </div>
                    <TextAreaField
                      label={t(locale, 'support.replyPlaceholder')}
                      value={supportReplies[thread.id] ?? ''}
                      onChange={(value) => setSupportReplies((current) => ({ ...current, [thread.id]: value }))}
                      placeholder={t(locale, 'support.replyPlaceholder')}
                    />
                    <button
                      type="button"
                      className="button button--ghost"
                      disabled={!String(supportReplies[thread.id] ?? '').trim()}
                      onClick={() => supportReplyMutation.mutate({ threadID: thread.id, body: supportReplies[thread.id] ?? '' })}
                    >
                      {t(locale, 'support.replyAction')}
                    </button>
                  </article>
                ))
              : (discussionQuery.data?.items ?? []).map((thread) => (
                  <article key={thread.id} className="feed-card">
                    <div className="feed-card__header">
                      <div>
                        <strong>{thread.title}</strong>
                        <p className="muted">{thread.category}</p>
                      </div>
                      <span className="badge badge--soft">{t(locale, 'support.feed.communityTab')}</span>
                    </div>
                    <TextAreaField
                      label={t(locale, 'support.replyPlaceholder')}
                      value={discussionReplies[thread.id] ?? ''}
                      onChange={(value) => setDiscussionReplies((current) => ({ ...current, [thread.id]: value }))}
                      placeholder={t(locale, 'support.replyPlaceholder')}
                    />
                    <button
                      type="button"
                      className="button button--ghost"
                      disabled={!String(discussionReplies[thread.id] ?? '').trim()}
                      onClick={() => discussionReplyMutation.mutate({ threadID: thread.id, body: discussionReplies[thread.id] ?? '' })}
                    >
                      {t(locale, 'support.replyAction')}
                    </button>
                  </article>
                ))}
          </div>
        </section>

        <aside className="card card--panel support-guidance">
          <div className="section-header">
            <div>
              <h2>{t(locale, 'support.guidelinesTitle')}</h2>
              <p className="muted">{t(locale, 'support.guidelinesBody')}</p>
            </div>
            <Sparkles size={18} aria-hidden="true" />
          </div>
          <div className="stack">
            <article className="notice notice--subtle">
              <MessageSquareText size={18} aria-hidden="true" />
              <span>{t(locale, 'support.guideline.one')}</span>
            </article>
            <article className="notice notice--subtle">
              <LifeBuoy size={18} aria-hidden="true" />
              <span>{t(locale, 'support.guideline.two')}</span>
            </article>
            <article className="notice notice--subtle">
              <Newspaper size={18} aria-hidden="true" />
              <span>{t(locale, 'support.guideline.three')}</span>
            </article>
          </div>
        </aside>
      </section>
    </div>
  );
}

function translateStatus(locale: SupportedLocale, value: string) {
  const translated = t(locale, `status.${value}`);
  return translated === `status.${value}` ? value : translated;
}
