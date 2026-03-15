import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiRequest } from '../../shared/api/client';
import { t, type SupportedLocale } from '../../shared/i18n';
import { Field } from '../../shared/ui/forms';

export function SupportPage({ locale }: { locale: SupportedLocale }) {
  const queryClient = useQueryClient();
  const [threadTitle, setThreadTitle] = useState('');
  const [threadBody, setThreadBody] = useState('');
  const [discussionTitle, setDiscussionTitle] = useState('');
  const [discussionBody, setDiscussionBody] = useState('');
  const [category, setCategory] = useState('general');
  const [supportReplies, setSupportReplies] = useState<Record<string, string>>({});
  const [discussionReplies, setDiscussionReplies] = useState<Record<string, string>>({});

  const supportQuery = useQuery({
    queryKey: ['support-threads'],
    queryFn: () => apiRequest<{ items: Array<{ id: string; title: string; status: string }> }>('/support/threads')
  });
  const discussionQuery = useQuery({
    queryKey: ['discussion-threads'],
    queryFn: () =>
      apiRequest<{ items: Array<{ id: string; title: string; category: string }> }>('/discussions/threads')
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

  return (
    <>
      <section className="card">
        <div className="section-header">
          <div>
            <h1>{t(locale, 'support.title')}</h1>
            <p className="muted">{t(locale, 'support.subtitle')}</p>
          </div>
          <div className="shell-highlights">
            <span className="badge badge--soft">{supportQuery.data?.items?.length ?? 0}</span>
            <span className="badge badge--soft">{discussionQuery.data?.items?.length ?? 0}</span>
          </div>
        </div>
      </section>

      <section className="dashboard-grid dashboard-grid--support">
        <section className="card">
          <h2>{t(locale, 'support.private')}</h2>
          <p className="muted">{t(locale, 'support.privateHint')}</p>
          <Field label={t(locale, 'support.threadTitle')} value={threadTitle} onChange={setThreadTitle} />
          <label className="field">
            <span>{t(locale, 'support.threadBody')}</span>
            <textarea value={threadBody} onChange={(event) => setThreadBody(event.target.value)} />
          </label>
          <div className="button-row">
            <button type="button" className="button button--primary" onClick={() => supportMutation.mutate()}>
              {t(locale, 'support.threadCreate')}
            </button>
          </div>
          <div className="stack">
            {(supportQuery.data?.items ?? []).map((thread) => (
              <article key={thread.id} className="notice">
                <strong>{thread.title}</strong>
                <span className="muted">{translateStatus(locale, thread.status)}</span>
                <textarea
                  placeholder={t(locale, 'support.replyPlaceholder')}
                  value={supportReplies[thread.id] ?? ''}
                  onChange={(event) =>
                    setSupportReplies((current) => ({ ...current, [thread.id]: event.target.value }))
                  }
                />
                <div className="button-row">
                  <button
                    type="button"
                    className="button"
                    onClick={() => supportReplyMutation.mutate({ threadID: thread.id, body: supportReplies[thread.id] ?? '' })}
                  >
                    {t(locale, 'support.replyAction')}
                  </button>
                </div>
              </article>
            ))}
          </div>
        </section>

        <section className="card">
          <h2>{t(locale, 'support.public')}</h2>
          <p className="muted">{t(locale, 'support.publicHint')}</p>
          <Field label={t(locale, 'support.threadTitle')} value={discussionTitle} onChange={setDiscussionTitle} />
          <Field label={t(locale, 'support.category')} value={category} onChange={setCategory} />
          <label className="field">
            <span>{t(locale, 'support.threadBody')}</span>
            <textarea value={discussionBody} onChange={(event) => setDiscussionBody(event.target.value)} />
          </label>
          <div className="button-row">
            <button type="button" className="button button--primary" onClick={() => discussionMutation.mutate()}>
              {t(locale, 'support.discussionCreate')}
            </button>
          </div>
          <div className="stack">
            {(discussionQuery.data?.items ?? []).map((thread) => (
              <article key={thread.id} className="notice">
                <strong>{thread.title}</strong>
                <span className="muted">{thread.category}</span>
                <textarea
                  placeholder={t(locale, 'support.replyPlaceholder')}
                  value={discussionReplies[thread.id] ?? ''}
                  onChange={(event) =>
                    setDiscussionReplies((current) => ({ ...current, [thread.id]: event.target.value }))
                  }
                />
                <div className="button-row">
                  <button
                    type="button"
                    className="button"
                    onClick={() =>
                      discussionReplyMutation.mutate({ threadID: thread.id, body: discussionReplies[thread.id] ?? '' })
                    }
                  >
                    {t(locale, 'support.replyAction')}
                  </button>
                </div>
              </article>
            ))}
          </div>
        </section>
      </section>
    </>
  );
}

function translateStatus(locale: SupportedLocale, value: string) {
  const translated = t(locale, `status.${value}`);
  return translated === `status.${value}` ? value : translated;
}
