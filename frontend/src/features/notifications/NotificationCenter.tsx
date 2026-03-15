import { t, type SupportedLocale } from '../../shared/i18n';

export function NotificationCenter({
  locale,
  items
}: {
  locale: SupportedLocale;
  items: Array<{ id: string; title: string; type: string; read: boolean; createdAt: string; targetURL?: string }>;
}) {
  const unread = items.filter((item) => !item.read).length;

  return (
    <section className="card">
      <div className="section-header">
        <h2>{t(locale, 'notifications.title')}</h2>
        <span className="muted">
          {t(locale, unread === 1 ? 'notifications.unread.one' : 'notifications.unread.many').replace('{count}', String(unread))}
        </span>
      </div>
      {items.length === 0 ? <p className="muted">{t(locale, 'notifications.empty')}</p> : null}
      <div className="stack">
        {items.map((item) => (
          <article key={item.id} className={`notice ${item.read ? 'notice--read' : ''}`}>
            <strong>{item.title}</strong>
            <span className="muted">{item.createdAt}</span>
            {item.targetURL ? <span className="muted">{item.targetURL}</span> : null}
          </article>
        ))}
      </div>
    </section>
  );
}
