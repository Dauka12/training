import { useState, type ReactNode } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import {
  ChangePasswordForm,
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
        onboardingDone: me.onboarding_done,
        mustChangePassword: (me as { must_change_password?: boolean }).must_change_password ?? false
      });
      if ((me as { must_change_password?: boolean }).must_change_password) {
        navigate('/change-password');
        return;
      }
      navigate(me.onboarding_done ? '/today' : '/profile');
    },
    onError: (error: Error) => setMessage(error.message)
  });

  if (isAuthenticated) {
    return <SectionPage title={t(locale, 'today.title')} />;
  }

  return (
    <AuthScaffold
      locale={locale}
      footer={
        <>
          <Link to="/register">{t(locale, 'auth.links.register')}</Link>
          <Link to="/forgot-password">{t(locale, 'auth.links.forgot')}</Link>
        </>
      }
    >
      <LoginForm
        locale={locale}
        onSubmit={mutation.mutateAsync}
        pending={mutation.isPending}
        message={message}
        lead={t(locale, 'auth.login.lead')}
      />
    </AuthScaffold>
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
    <AuthScaffold
      locale={locale}
      devToken={devToken}
      footer={
        <>
          <Link to="/verify-email">{t(locale, 'auth.links.verify')}</Link>
          <Link to="/login">{t(locale, 'auth.links.login')}</Link>
        </>
      }
    >
      <RegisterForm
        locale={locale}
        onSubmit={mutation.mutateAsync}
        pending={mutation.isPending}
        message={message}
        lead={t(locale, 'auth.register.lead')}
      />
    </AuthScaffold>
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
    <AuthScaffold locale={locale} footer={<Link to="/login">{t(locale, 'auth.links.login')}</Link>}>
      <VerifyForm
        locale={locale}
        onSubmit={mutation.mutateAsync}
        pending={mutation.isPending}
        message={message}
        lead={t(locale, 'auth.verify.lead')}
      />
    </AuthScaffold>
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
    <AuthScaffold
      locale={locale}
      devToken={devToken}
      footer={<Link to="/reset-password">{t(locale, 'auth.reset.title')}</Link>}
    >
      <ForgotPasswordForm
        locale={locale}
        onSubmit={mutation.mutateAsync}
        pending={mutation.isPending}
        message={message}
        lead={t(locale, 'auth.forgot.lead')}
      />
    </AuthScaffold>
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
    <AuthScaffold locale={locale} footer={<Link to="/login">{t(locale, 'auth.links.login')}</Link>}>
      <ResetPasswordForm
        locale={locale}
        onSubmit={mutation.mutateAsync}
        pending={mutation.isPending}
        message={message}
        lead={t(locale, 'auth.reset.lead')}
      />
    </AuthScaffold>
  );
}

export function ChangePasswordPage({ locale }: { locale: SupportedLocale }) {
  const navigate = useNavigate();
  const setAuthenticated = useAuthStore((state) => state.setAuthenticated);
  const authRole = useAuthStore((state) => state.role);
  const authEmail = useAuthStore((state) => state.email);
  const onboardingDone = useAuthStore((state) => state.onboardingDone);
  const [message, setMessage] = useState('');
  const mutation = useMutation({
    mutationFn: (payload: Record<string, string>) =>
      apiRequest('/auth/change-password', {
        method: 'POST',
        body: JSON.stringify(payload)
      }),
    onSuccess: () => {
      setAuthenticated({
        role: authRole,
        email: authEmail,
        onboardingDone,
        mustChangePassword: false
      });
      setMessage(t(locale, 'auth.success.passwordChanged'));
      navigate(authRole === 'admin' ? '/admin' : onboardingDone ? '/today' : '/profile');
    },
    onError: (error: Error) => setMessage(error.message)
  });

  return (
    <AuthScaffold locale={locale} footer={<Link to="/login">{t(locale, 'auth.links.login')}</Link>}>
      <ChangePasswordForm
        locale={locale}
        onSubmit={mutation.mutateAsync}
        pending={mutation.isPending}
        message={message}
        lead={t(locale, 'auth.changePassword.lead')}
      />
    </AuthScaffold>
  );
}

function AuthScaffold({
  locale,
  children,
  footer,
  devToken
}: {
  locale: SupportedLocale;
  children: ReactNode;
  footer: ReactNode;
  devToken?: string;
}) {
  const hydrationPreview = `2600 ${t(locale, 'common.unitMl')}`;

  return (
    <main className="auth-layout">
      <section className="card auth-panel auth-panel--brand">
        <div className="auth-panel__brand">
          <span className="badge">{t(locale, 'landing.free')}</span>
          <strong>{t(locale, 'brand.name')}</strong>
        </div>
        <div className="stack">
          <h2>{t(locale, 'landing.dashboard.title')}</h2>
          <p className="muted">{t(locale, 'landing.dashboard.body')}</p>
        </div>
        <div className="auth-metric-grid">
          <article className="stat stat--compact">
            <span className="muted">{t(locale, 'landing.preview.health')}</span>
            <strong>{t(locale, 'status.healthy')}</strong>
          </article>
          <article className="stat stat--compact">
            <span className="muted">{t(locale, 'landing.preview.hydration')}</span>
            <strong>{hydrationPreview}</strong>
          </article>
          <article className="stat stat--compact">
            <span className="muted">{t(locale, 'landing.preview.notifications')}</span>
            <strong>3</strong>
          </article>
        </div>
        <div className="stack auth-points">
          <article className="notice">
            <strong>{t(locale, 'landing.benefit.plan')}</strong>
            <span className="muted">{t(locale, 'landing.how.two')}</span>
          </article>
          <article className="notice">
            <strong>{t(locale, 'landing.benefit.track')}</strong>
            <span className="muted">{t(locale, 'landing.how.three')}</span>
          </article>
          <article className="notice">
            <strong>{t(locale, 'landing.benefit.support')}</strong>
            <span className="muted">{t(locale, 'landing.free.body')}</span>
          </article>
        </div>
      </section>

      <section className="card auth-panel auth-panel--form">
        <div className="auth-panel__form">{children}</div>
        {devToken ? <p className="form-message">{t(locale, 'auth.devToken')}: {devToken}</p> : null}
        <div className="auth-link-grid">{footer}</div>
      </section>
    </main>
  );
}
