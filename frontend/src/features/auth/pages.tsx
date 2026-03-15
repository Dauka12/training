import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import {
  ForgotPasswordForm,
  LoginForm,
  RegisterForm,
  ResetPasswordForm,
  VerifyForm
} from './forms';
import { apiRequest } from '../../shared/api/client';
import { useAuthStore, type AuthRole } from '../../shared/auth/store';
import { t, type SupportedLocale } from '../../shared/i18n';
import { SectionPage } from '../../shared/ui/forms';

export function LoginPage({ locale, isAuthenticated }: { locale: SupportedLocale; isAuthenticated?: boolean }) {
  const navigate = useNavigate();
  const setAuthenticated = useAuthStore((state) => state.setAuthenticated);
  const [message, setMessage] = useState('');
  const mutation = useMutation({
    mutationFn: (payload: Record<string, string>) =>
      apiRequest('/auth/login', {
        method: 'POST',
        body: JSON.stringify(payload)
      }),
    onSuccess: async () => {
      const me = await apiRequest<{ email: string; roles: string[]; onboarding_done: boolean }>('/me');
      setAuthenticated({
        role: (me.roles[0] as AuthRole) ?? 'user',
        email: me.email,
        onboardingDone: me.onboarding_done
      });
      navigate(me.onboarding_done ? '/today' : '/profile');
    },
    onError: (error: Error) => setMessage(error.message)
  });

  if (isAuthenticated) {
    return <SectionPage title={t(locale, 'today.title')} />;
  }

  return (
    <main className="layout">
      <section className="card">
        <LoginForm locale={locale} onSubmit={mutation.mutateAsync} pending={mutation.isPending} message={message} />
        <div className="link-stack">
          <Link to="/register">{t(locale, 'auth.links.register')}</Link>
          <Link to="/forgot-password">{t(locale, 'auth.links.forgot')}</Link>
        </div>
      </section>
    </main>
  );
}

export function RegisterPage({ locale }: { locale: SupportedLocale }) {
  const [message, setMessage] = useState('');
  const [devToken, setDevToken] = useState('');
  const mutation = useMutation({
    mutationFn: (payload: Record<string, string>) =>
      apiRequest<{ dev_verification_token?: string }>('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ ...payload, locale })
      }),
    onSuccess: (response) => {
      setMessage(t(locale, 'auth.success.registered'));
      setDevToken(response.dev_verification_token ?? '');
    },
    onError: (error: Error) => setMessage(error.message)
  });

  return (
    <main className="layout">
      <section className="card">
        <RegisterForm locale={locale} onSubmit={mutation.mutateAsync} pending={mutation.isPending} message={message} />
        {devToken ? <p className="form-message">{t(locale, 'auth.devToken')}: {devToken}</p> : null}
        <div className="link-stack">
          <Link to="/verify-email">{t(locale, 'auth.links.verify')}</Link>
          <Link to="/login">{t(locale, 'auth.links.login')}</Link>
        </div>
      </section>
    </main>
  );
}

export function VerifyPage({ locale }: { locale: SupportedLocale }) {
  const [message, setMessage] = useState('');
  const mutation = useMutation({
    mutationFn: (payload: Record<string, string>) =>
      apiRequest('/auth/verify-email', {
        method: 'POST',
        body: JSON.stringify(payload)
      }),
    onSuccess: () => setMessage(t(locale, 'auth.success.verified')),
    onError: (error: Error) => setMessage(error.message)
  });

  return (
    <main className="layout">
      <section className="card">
        <VerifyForm locale={locale} onSubmit={mutation.mutateAsync} pending={mutation.isPending} message={message} />
        <Link to="/login">{t(locale, 'auth.links.login')}</Link>
      </section>
    </main>
  );
}

export function ForgotPasswordPage({ locale }: { locale: SupportedLocale }) {
  const [message, setMessage] = useState('');
  const [devToken, setDevToken] = useState('');
  const mutation = useMutation({
    mutationFn: (payload: Record<string, string>) =>
      apiRequest<{ dev_reset_token?: string }>('/auth/forgot-password', {
        method: 'POST',
        body: JSON.stringify(payload)
      }),
    onSuccess: (response) => {
      setMessage(t(locale, 'auth.success.resetRequested'));
      setDevToken(response.dev_reset_token ?? '');
    },
    onError: (error: Error) => setMessage(error.message)
  });

  return (
    <main className="layout">
      <section className="card">
        <ForgotPasswordForm locale={locale} onSubmit={mutation.mutateAsync} pending={mutation.isPending} message={message} />
        {devToken ? <p className="form-message">{t(locale, 'auth.devToken')}: {devToken}</p> : null}
        <Link to="/reset-password">{t(locale, 'auth.reset.title')}</Link>
      </section>
    </main>
  );
}

export function ResetPasswordPage({ locale }: { locale: SupportedLocale }) {
  const [message, setMessage] = useState('');
  const mutation = useMutation({
    mutationFn: (payload: Record<string, string>) =>
      apiRequest('/auth/reset-password', {
        method: 'POST',
        body: JSON.stringify(payload)
      }),
    onSuccess: () => setMessage(t(locale, 'auth.success.passwordReset')),
    onError: (error: Error) => setMessage(error.message)
  });

  return (
    <main className="layout">
      <section className="card">
        <ResetPasswordForm locale={locale} onSubmit={mutation.mutateAsync} pending={mutation.isPending} message={message} />
        <Link to="/login">{t(locale, 'auth.links.login')}</Link>
      </section>
    </main>
  );
}
