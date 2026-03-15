import { FormEvent, useState } from 'react';
import { t, type SupportedLocale } from '../../shared/i18n';

type AuthFormProps = {
  locale: SupportedLocale;
  onSubmit: (payload: Record<string, string>) => Promise<unknown>;
  pending?: boolean;
  message?: string;
};

export function LoginForm({ locale, onSubmit, pending, message }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.login.title"
      submitKey="auth.login.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      fields={['email', 'password']}
    />
  );
}

export function RegisterForm({ locale, onSubmit, pending, message }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.register.title"
      submitKey="auth.register.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      fields={['email', 'password']}
    />
  );
}

export function VerifyForm({ locale, onSubmit, pending, message }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.verify.title"
      submitKey="auth.verify.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      fields={['token']}
    />
  );
}

export function ForgotPasswordForm({ locale, onSubmit, pending, message }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.forgot.title"
      submitKey="auth.forgot.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      fields={['email']}
    />
  );
}

export function ResetPasswordForm({ locale, onSubmit, pending, message }: AuthFormProps) {
  return (
    <AuthForm
      locale={locale}
      titleKey="auth.reset.title"
      submitKey="auth.reset.submit"
      onSubmit={onSubmit}
      pending={pending}
      message={message}
      fields={['token', 'new_password']}
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
  fields
}: AuthFormProps & { titleKey: string; submitKey: string; fields: Array<'email' | 'password' | 'token' | 'new_password'> }) {
  const [values, setValues] = useState<Record<string, string>>({});

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await onSubmit(values);
  }

  return (
    <form className="form-card" onSubmit={handleSubmit}>
      <h1>{t(locale, titleKey)}</h1>
      {fields.map((field) => (
        <label key={field} className="field">
          <span>{labelFor(locale, field)}</span>
          <input
            name={field}
            type={inputTypeFor(field)}
            value={values[field] ?? ''}
            onChange={(event) => setValues((current) => ({ ...current, [field]: event.target.value }))}
          />
        </label>
      ))}
      {message ? <p className="form-message">{message}</p> : null}
      <button type="submit" className="button button--primary" disabled={pending}>
        {pending ? t(locale, 'common.loading') : t(locale, submitKey)}
      </button>
    </form>
  );
}

function labelFor(locale: SupportedLocale, field: 'email' | 'password' | 'token' | 'new_password') {
  switch (field) {
    case 'email':
      return t(locale, 'auth.email');
    case 'password':
      return t(locale, 'auth.password');
    case 'token':
      return t(locale, 'auth.token');
    default:
      return t(locale, 'auth.newPassword');
  }
}

function inputTypeFor(field: 'email' | 'password' | 'token' | 'new_password') {
  if (field === 'email') {
    return 'email';
  }
  if (field === 'token') {
    return 'text';
  }
  return 'password';
}
