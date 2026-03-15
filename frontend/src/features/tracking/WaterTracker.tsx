import { useState } from 'react';
import { t, type SupportedLocale } from '../../shared/i18n';

type WaterTrackerProps = {
  currentML: number;
  targetML: number;
  locale?: SupportedLocale;
  onQuickAdd?: (amount: number) => void;
  onCustomAdd?: (amount: number) => void;
};

export function WaterTracker({
  currentML,
  targetML,
  locale = 'ru',
  onQuickAdd,
  onCustomAdd
}: WaterTrackerProps) {
  const [customAmount, setCustomAmount] = useState('300');
  const unitML = t(locale, 'common.unitMl');

  return (
    <section className="stack">
      <p className="metric">
        <strong>{currentML}</strong> / {targetML} {unitML}
      </p>
      <div className="button-row">
        <button type="button" className="button" onClick={() => onQuickAdd?.(250)}>
          {t(locale, 'tracking.water.quick250')}
        </button>
        <button type="button" className="button" onClick={() => onQuickAdd?.(500)}>
          {t(locale, 'tracking.water.quick500')}
        </button>
      </div>
      <div className="inline-form">
        <label className="field field--inline">
          <span>{t(locale, 'tracking.water.custom')}</span>
          <input value={customAmount} onChange={(event) => setCustomAmount(event.target.value)} inputMode="numeric" />
        </label>
        <button
          type="button"
          className="button button--primary"
          onClick={() => onCustomAdd?.(Number(customAmount))}
        >
          {t(locale, 'tracking.water.add')}
        </button>
      </div>
    </section>
  );
}
