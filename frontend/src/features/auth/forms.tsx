import { FormEvent, useState } from 'react';
import { Chrome, KeyRound, LockKeyhole, Mail } from 'lucide-react';
import { t, type SupportedLocale } from '../../shared/i18n';

const googleAuthURL = `${(import.meta.env.VITE_API_BASE_URL as string | undefined) ?? 'http://localhost:8080/api/v1'}/auth/google/start`;

type AuthFormProps = {
  locale: SupportedLocale;
  onSubmit: (payload: Record<string, string>) => Promise<unknown>;
  pending?: boolean;
  message?: string;
  lead?: string;
};

export function LoginForm({ locale, onSubmit, pending, message, lead }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.login.title"
      submitKey="auth.login.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      lead={lead}
      fields={['email', 'password']}
    />
  );
}

export function RegisterForm({ locale, onSubmit, pending, message, lead }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.register.title"
      submitKey="auth.register.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      lead={lead}
      fields={['email', 'password']}
    />
  );
}

export function VerifyForm({ locale, onSubmit, pending, message, lead }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.verify.title"
      submitKey="auth.verify.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      lead={lead}
      fields={['token']}
      showGoogle={false}
    />
  );
}

export function ForgotPasswordForm({ locale, onSubmit, pending, message, lead }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.forgot.title"
      submitKey="auth.forgot.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      lead={lead}
      fields={['email']}
      showGoogle={false}
    />
  );
}

export function ResetPasswordForm({ locale, onSubmit, pending, message, lead }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.reset.title"
      submitKey="auth.reset.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      lead={lead}
      fields={['token', 'new_password']}
      showGoogle={false}
    />
  );
}

export function ChangePasswordForm({ locale, onSubmit, pending, message, lead }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.changePassword.title"
      submitKey="auth.changePassword.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      lead={lead}
      fields={['current_password', 'new_password']}
      showGoogle={false}
    />
  );
}

function AuthForm({
  locale,
  titleKey,
  submitKey,
  onSubmit,
  pending,
  message,
  lead,
  fields,
  showGoogle = true
}: AuthFormProps & { titleKey: string; submitKey: string; fields: Array<'email' | 'password' | 'token' | 'new_password' | 'current_password'>; showGoogle?: boolean }) {
  const [values, setValues] = useState<Record<string, string>>({});

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onSubmit(values);
  }

  return (
    <form className="form-card auth-form" onSubmit={handleSubmit}>
      <div className="auth-form__header">
        <h1>{t(locale, titleKey)}</h1>
        {lead ? <p className="muted form-card__lead">{lead}</p> : null}
      </div>

      {showGoogle ? (
        <>
          <a className="button button--social" href={googleAuthURL}>
            <Chrome size={18} aria-hidden="true" />
            <span>{t(locale, 'auth.google.continue')}</span>
          </a>
          <div className="auth-form__divider" aria-hidden="true">
            <span>{t(locale, 'auth.google.or')}</span>
          </div>
        </>
      ) : null}

      <div className="auth-form__fields">
        {fields.map((field) => (
          <label key={field} className="field field--with-icon">
            <span>{labelFor(locale, field)}</span>
            <div className="field-shell">
              <span className="field-shell__icon" aria-hidden="true">
                {iconFor(field)}
              </span>
              <input
                aria-label={labelFor(locale, field)}
                name={field}
                type={inputTypeFor(field)}
                value={values[field] ?? ''}
                placeholder={placeholderFor(locale, field)}
                onChange={(event) => setValues((current) => ({ ...current, [field]: event.target.value }))}
              />
            </div>
          </label>
        ))}
      </div>

      {message ? <p className="form-message">{message}</p> : null}
      <button type="submit" className="button button--primary auth-form__submit" disabled={pending}>
        {pending ? t(locale, 'common.loading') : t(locale, submitKey)}
      </button>
    </form>
  );
}

function labelFor(locale: SupportedLocale, field: 'email' | 'password' | 'token' | 'new_password' | 'current_password') {
  switch (field) {
    case 'email':
      return t(locale, 'auth.email');
    case 'password':
      return t(locale, 'auth.password');
    case 'current_password':
      return t(locale, 'auth.currentPassword');
    case 'token':
      return t(locale, 'auth.token');
    default:
      return t(locale, 'auth.newPassword');
  }
}

function inputTypeFor(field: 'email' | 'password' | 'token' | 'new_password' | 'current_password') {
  if (field === 'email') {
    return 'email';
  }
  if (field === 'token') {
    return 'text';
  }
  return 'password';
}

function placeholderFor(locale: SupportedLocale, field: 'email' | 'password' | 'token' | 'new_password' | 'current_password') {
  switch (field) {
    case 'email':
      return 'you@example.com';
    case 'password':
      return t(locale, 'auth.password');
    case 'current_password':
      return t(locale, 'auth.currentPassword');
    case 'token':
      return 'a1b2c3...';
    default:
      return t(locale, 'auth.newPassword');
  }
}

function iconFor(field: 'email' | 'password' | 'token' | 'new_password' | 'current_password') {
  switch (field) {
    case 'email':
      return <Mail size={18} />;
    case 'token':
      return <KeyRound size={18} />;
    default:
      return <LockKeyhole size={18} />;
  }
}
