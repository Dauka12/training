import type { HTMLInputTypeAttribute } from 'react';
import { ChevronDown } from 'lucide-react';
import { t, type SupportedLocale } from '../i18n';

export function SectionPage({ title }: { title: string }) {
  return (
    <section className="card">
      <h1>{title}</h1>
    </section>
  );
}

export function CardStat({ title, value }: { title: string; value: string }) {
  return (
    <article className="stat">
      <span className="muted">{title}</span>
      <strong>{value}</strong>
    </article>
  );
}

export function Field({
  label,
  value,
  onChange,
  type = 'text',
  placeholder,
  inputMode,
  min,
  max,
  step,
  autoComplete
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: HTMLInputTypeAttribute;
  placeholder?: string;
  inputMode?: 'none' | 'text' | 'tel' | 'url' | 'email' | 'numeric' | 'decimal' | 'search';
  min?: number | string;
  max?: number | string;
  step?: number | string;
  autoComplete?: string;
}) {
  return (
    <label className="field">
      <span>{label}</span>
      <input
        aria-label={label}
        type={type}
        value={value}
        placeholder={placeholder}
        inputMode={inputMode}
        min={min}
        max={max}
        step={step}
        autoComplete={autoComplete}
        onChange={(event) => onChange(event.target.value)}
      />
    </label>
  );
}

export function SelectField({
  label,
  value,
  onChange,
  options,
  disabled = false
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: Array<{ value: string; label: string }>;
  disabled?: boolean;
}) {
  return (
    <label className="field">
      <span>{label}</span>
      <div className="select-shell">
        <select aria-label={label} value={value} onChange={(event) => onChange(event.target.value)} disabled={disabled}>
          {options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        <ChevronDown aria-hidden="true" className="select-shell__icon" size={18} />
      </div>
    </label>
  );
}

export function TextAreaField({
  label,
  value,
  onChange,
  placeholder
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}) {
  return (
    <label className="field">
      <span>{label}</span>
      <textarea aria-label={label} value={value} placeholder={placeholder} onChange={(event) => onChange(event.target.value)} />
    </label>
  );
}

export function Checkbox({
  label,
  checked,
  onChange
}: {
  label: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <label className="checkbox">
      <input type="checkbox" checked={checked} onChange={(event) => onChange(event.target.checked)} />
      <span>{label}</span>
    </label>
  );
}

export function StatusSelect({
  value,
  onChange,
  options
}: {
  value: string;
  onChange: (value: string) => void;
  options: Array<{ value: string; label: string }>;
}) {
  return (
    <div className="select-shell">
      <select value={value} onChange={(event) => onChange(event.target.value)}>
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      <ChevronDown aria-hidden="true" className="select-shell__icon" size={18} />
    </div>
  );
}

export function EmptyState({ locale }: { locale: SupportedLocale }) {
  return <p className="muted">{t(locale, 'common.empty')}</p>;
}

export function splitComma(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean);
}

export function parseAvailability(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item) => {
      const [weekday, duration] = item.split(':');
      return { weekday, duration_min: Number(duration) || 45 };
    });
}
