import type { ReactElement } from 'react';
import { useMemo } from 'react';
import { BrowserRouter, Link, MemoryRouter, NavLink, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Activity, BadgeHelp, ChartNoAxesCombined, CircleUserRound, ClipboardList, House, LayoutDashboard, ShieldCheck, Users } from 'lucide-react';
import {
  ForgotPasswordPage,
  LoginPage,
  RegisterPage,
  ResetPasswordPage,
  VerifyPage
} from '../features/auth/pages';
import { PlanPage, ProgressPage, TodayPage, TrackPage } from '../features/dashboard/pages';
import { AdminPage, TrainerPage } from '../features/panels/pages';
import { ProfilePage } from '../features/profile/ProfilePage';
import { SupportPage } from '../features/support/SupportPage';
import { apiRequest } from '../shared/api/client';
import { useAuthStore, type AuthRole, type AuthState } from '../shared/auth/store';
import { t, type SupportedLocale } from '../shared/i18n';
import { SectionPage } from '../shared/ui/forms';

type AuthSnapshot = Pick<AuthState, 'isAuthenticated' | 'role'> & { email?: string; onboardingDone?: boolean };
type NavItem = { to: string; label: string; icon: ReactElement };

export function AppRouter({
  initialEntries,
  initialAuth,
  locale = 'ru'
}: {
  initialEntries?: string[];
  initialAuth?: AuthSnapshot;
  locale?: SupportedLocale;
}) {
  const client = useMemo(() => new QueryClient({ defaultOptions: { queries: { retry: false } } }), []);

  return (
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={initialEntries ?? ['/']}>
        <AppRoutes locale={locale} authOverride={initialAuth} />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

const browserQueryClient = new QueryClient({
  defaultOptions: {
    queries: { retry: false, staleTime: 10_000 }
  }
});

export function BrowserAppRouter({ locale }: { locale: SupportedLocale }) {
  return (
    <QueryClientProvider client={browserQueryClient}>
      <BrowserRouter>
        <AppRoutes locale={locale} />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

function AppRoutes({ locale, authOverride }: { locale: SupportedLocale; authOverride?: AuthSnapshot }) {
  const auth = useAuthStore();
  const effectiveAuth = authOverride ?? auth;

  return (
    <Routes>
      <Route path="/" element={<LandingPage locale={locale} />} />
      <Route path="/login" element={<LoginPage locale={locale} isAuthenticated={effectiveAuth.isAuthenticated} />} />
      <Route path="/register" element={<RegisterPage locale={locale} />} />
      <Route path="/verify-email" element={<VerifyPage locale={locale} />} />
      <Route path="/forgot-password" element={<ForgotPasswordPage locale={locale} />} />
      <Route path="/reset-password" element={<ResetPasswordPage locale={locale} />} />
      <Route path="/today" element={<Protected locale={locale} auth={effectiveAuth}><RequiresOnboarding auth={effectiveAuth}><Shell locale={locale} auth={effectiveAuth}><TodayPage locale={locale} /></Shell></RequiresOnboarding></Protected>} />
      <Route path="/plan" element={<Protected locale={locale} auth={effectiveAuth}><RequiresOnboarding auth={effectiveAuth}><Shell locale={locale} auth={effectiveAuth}><PlanPage locale={locale} /></Shell></RequiresOnboarding></Protected>} />
      <Route path="/track" element={<Protected locale={locale} auth={effectiveAuth}><RequiresOnboarding auth={effectiveAuth}><Shell locale={locale} auth={effectiveAuth}><TrackPage locale={locale} /></Shell></RequiresOnboarding></Protected>} />
      <Route path="/progress" element={<Protected locale={locale} auth={effectiveAuth}><RequiresOnboarding auth={effectiveAuth}><Shell locale={locale} auth={effectiveAuth}><ProgressPage locale={locale} /></Shell></RequiresOnboarding></Protected>} />
      <Route path="/profile" element={<Protected locale={locale} auth={effectiveAuth}><Shell locale={locale} auth={effectiveAuth}><ProfilePage locale={locale} /></Shell></Protected>} />
      <Route path="/support" element={<Protected locale={locale} auth={effectiveAuth}><Shell locale={locale} auth={effectiveAuth}><SupportPage locale={locale} /></Shell></Protected>} />
      <Route path="/trainer" element={<Protected locale={locale} auth={effectiveAuth}><Shell locale={locale} auth={effectiveAuth}><RoleGate locale={locale} auth={effectiveAuth} roles={['trainer', 'admin']}><TrainerPage locale={locale} /></RoleGate></Shell></Protected>} />
      <Route path="/admin" element={<Protected locale={locale} auth={effectiveAuth}><Shell locale={locale} auth={effectiveAuth}><RoleGate locale={locale} auth={effectiveAuth} roles={['admin']}><AdminPage locale={locale} /></RoleGate></Shell></Protected>} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

function Protected({ auth, locale, children }: { auth: AuthSnapshot; locale: SupportedLocale; children: ReactElement }) {
  if (!auth.isAuthenticated) {
    return <LoginPage locale={locale} />;
  }
  return children;
}

function RequiresOnboarding({ auth, children }: { auth: AuthSnapshot; children: ReactElement }) {
  if (auth.role === 'user' && auth.onboardingDone === false) {
    return <Navigate to="/profile" replace />;
  }
  return children;
}

function RoleGate({
  auth,
  roles,
  locale,
  children
}: {
  auth: AuthSnapshot;
  roles: AuthRole[];
  locale: SupportedLocale;
  children: ReactElement;
}) {
  if (!roles.includes(auth.role)) {
    return <SectionPage title={t(locale, 'common.forbidden')} />;
  }
  return children;
}

function Shell({ locale, auth, children }: { locale: SupportedLocale; auth: AuthSnapshot; children: ReactElement }) {
  const clear = useAuthStore((state) => state.clear);
  const navigate = useNavigate();
  const location = useLocation();
  const navItems = navigationItems(locale, auth.role);
  const onboardingLocked = auth.role === 'user' && auth.onboardingDone === false;
  const onboardingBlockedPaths = new Set(['/today', '/plan', '/track', '/progress']);
  const currentPage = navItems.find((item) => item.to === location.pathname) ?? {
    label: t(locale, 'shell.workspace'),
    to: location.pathname,
    icon: <LayoutDashboard size={18} />
  };
  const dockItems = mobileDockItems(navItems, currentPage);
  const quickItems = navItems.filter((item) =>
    ['/today', '/plan', '/track', '/progress', '/support'].includes(item.to)
  );

  async function handleLogout() {
    await apiRequest('/auth/logout', { method: 'POST' }).catch(() => undefined);
    clear();
    navigate('/login');
  }

  function openOnboarding() {
    if (location.pathname !== '/profile') {
      navigate('/profile');
      setTimeout(() => {
        document.getElementById('onboarding-section')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }, 0);
      return;
    }

    document.getElementById('onboarding-section')?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  return (
    <main className="workspace-shell" data-testid="workspace-shell">
      <aside className="workspace-shell__sidebar" data-testid="workspace-sidebar">
        <section className="card shell-card shell-card--brand rail-brand workspace-rail__brand">
          <div className="shell-brand">
            <span className="shell-brand__mark" aria-hidden="true">
              <LayoutDashboard size={18} />
            </span>
            <div>
              <strong>{t(locale, 'brand.name')}</strong>
              <p className="muted">{t(locale, 'shell.workspace')}</p>
            </div>
          </div>
          <p className="muted">{t(locale, 'shell.subtitle')}</p>
        </section>

        <nav className="card shell-card shell-nav workspace-rail__nav" aria-label={t(locale, 'shell.navigation')}>
          {navItems.map((item) => {
            const locked = onboardingLocked && onboardingBlockedPaths.has(item.to);

            if (locked) {
              return (
                <button
                  key={item.to}
                  type="button"
                  className="nav-link nav-link--rail nav-link--button"
                  onClick={openOnboarding}
                >
                  <span className="nav-link__icon" aria-hidden="true">{item.icon}</span>
                  <span className="nav-link__content">
                    <span className="nav-link__label">{item.label}</span>
                    <span className="nav-link__meta">{t(locale, 'profile.onboarding')}</span>
                  </span>
                </button>
              );
            }

            return (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) => `nav-link nav-link--rail${isActive ? ' nav-link--active' : ''}`}
              >
                <span className="nav-link__icon" aria-hidden="true">{item.icon}</span>
                <span className="nav-link__content">
                  <span className="nav-link__label">{item.label}</span>
                  <span className="nav-link__meta">{item.to === currentPage.to ? roleLabel(locale, auth.role) : t(locale, 'shell.workspace')}</span>
                </span>
              </NavLink>
            );
          })}
        </nav>

        <section className="card shell-card shell-account rail-account workspace-rail__account">
          <span className="muted shell-account__label">{t(locale, 'shell.account')}</span>
          <strong>{auth.email || t(locale, 'brand.name')}</strong>
          <span className="muted">{roleLabel(locale, auth.role)}</span>
          <button type="button" className="button button--ghost button--wide" onClick={() => void handleLogout()}>
            {t(locale, 'common.logout')}
          </button>
        </section>
      </aside>

      <section className="workspace-shell__main">
        <header className="card workspace-shell__header" data-testid="workspace-header">
          <div className="workspace-shell__headline">
            <p className="eyebrow">{t(locale, 'shell.workspace')}</p>
            <p className="workspace-shell__title">{currentPage.label}</p>
            <p className="muted">{t(locale, 'shell.subtitle')}</p>
          </div>
          <div className="workspace-shell__toolbar">
            <div className="workspace-shell__meta">
              <span className="badge badge--soft">{roleLabel(locale, auth.role)}</span>
              <span className="badge">{auth.email || t(locale, 'brand.name')}</span>
            </div>
            <nav className="workspace-shell__chips" aria-label={t(locale, 'shell.quickNavigation')}>
              {quickItems.map((item) => {
                const locked = onboardingLocked && onboardingBlockedPaths.has(item.to);

                if (locked) {
                  return (
                    <button
                      key={item.to}
                      type="button"
                      className="workspace-chip"
                      onClick={openOnboarding}
                    >
                      <span className="workspace-chip__icon" aria-hidden="true">{item.icon}</span>
                      <span>{item.label}</span>
                    </button>
                  );
                }

                return (
                  <NavLink
                    key={item.to}
                    to={item.to}
                    className={({ isActive }) => `workspace-chip${isActive ? ' workspace-chip--active' : ''}`}
                  >
                    <span className="workspace-chip__icon" aria-hidden="true">{item.icon}</span>
                    <span>{item.label}</span>
                  </NavLink>
                );
              })}
            </nav>
          </div>
        </header>
        <div className="workspace-shell__content">{children}</div>
        <nav className="mobile-dock" aria-label={`${t(locale, 'shell.quickNavigation')} mobile`}>
          {dockItems.map((item) => {
            const locked = onboardingLocked && onboardingBlockedPaths.has(item.to);

            if (locked) {
              return (
                <button
                  key={item.to}
                  type="button"
                  className="mobile-dock__link"
                  onClick={openOnboarding}
                >
                  <span className="mobile-dock__icon" aria-hidden="true">{item.icon}</span>
                  <span>{item.label}</span>
                </button>
              );
            }

            return (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) => `mobile-dock__link${isActive ? ' mobile-dock__link--active' : ''}`}
              >
                <span className="mobile-dock__icon" aria-hidden="true">{item.icon}</span>
                <span>{item.label}</span>
              </NavLink>
            );
          })}
        </nav>
      </section>
    </main>
  );
}

function LandingPage({ locale }: { locale: SupportedLocale }) {
  const hydrationPreview = `2600 ${t(locale, 'common.unitMl')}`;

  return (
    <main className="layout layout--landing">
      <section className="card hero hero--landing">
        <div className="hero__copy">
          <span className="badge">{t(locale, 'landing.free')}</span>
          <h1>{t(locale, 'landing.title')}</h1>
          <p>{t(locale, 'landing.subtitle')}</p>
          <div className="button-row">
            <Link className="button button--primary" to="/register">{t(locale, 'landing.cta')}</Link>
            <Link className="button button--ghost" to="/login">{t(locale, 'landing.cta.secondary')}</Link>
          </div>
        </div>
        <div className="hero__preview">
          <article className="notice notice--accent">
            <span className="muted">{t(locale, 'landing.preview.workout')}</span>
            <strong>{t(locale, 'today.title')}</strong>
            <span className="muted">{t(locale, 'landing.dashboard.body')}</span>
          </article>
          <div className="stats-grid">
            <article className="stat">
              <span className="muted">{t(locale, 'landing.preview.health')}</span>
              <strong>{t(locale, 'status.healthy')}</strong>
            </article>
            <article className="stat">
              <span className="muted">{t(locale, 'landing.preview.hydration')}</span>
              <strong>{hydrationPreview}</strong>
            </article>
            <article className="stat">
              <span className="muted">{t(locale, 'landing.preview.notifications')}</span>
              <strong>3</strong>
            </article>
          </div>
        </div>
      </section>

      <section className="card card--accent">
        <div className="section-header">
          <div>
            <h2>{t(locale, 'landing.free.title')}</h2>
            <p className="muted">{t(locale, 'landing.free.body')}</p>
          </div>
          <span className="badge">{t(locale, 'landing.free')}</span>
        </div>
      </section>

      <section className="card">
        <h2>{t(locale, 'landing.benefits.title')}</h2>
        <div className="feature-grid">
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
            <span className="muted">{t(locale, 'landing.dashboard.body')}</span>
          </article>
        </div>
      </section>

      <section className="landing-grid">
        <section className="card">
          <h2>{t(locale, 'landing.how.title')}</h2>
          <ul className="list">
            <li>{t(locale, 'landing.how.one')}</li>
            <li>{t(locale, 'landing.how.two')}</li>
            <li>{t(locale, 'landing.how.three')}</li>
          </ul>
        </section>

        <section className="card">
          <h2>{t(locale, 'landing.dashboard.title')}</h2>
          <p className="muted">{t(locale, 'landing.dashboard.body')}</p>
          <div className="button-row">
            <Link className="button" to="/login">{t(locale, 'auth.login.title')}</Link>
            <Link className="button button--primary" to="/register">{t(locale, 'landing.cta')}</Link>
          </div>
        </section>
      </section>

      <section className="card">
        <h2>{t(locale, 'landing.faq.title')}</h2>
        <p><strong>{t(locale, 'landing.faq.free.q')}</strong></p>
        <p>{t(locale, 'landing.faq.free.a')}</p>
      </section>

      <section className="card legal-grid">
        <p>{t(locale, 'landing.privacy')}</p>
        <p>{t(locale, 'landing.terms')}</p>
      </section>
    </main>
  );
}

function navigationItems(locale: SupportedLocale, role: AuthRole): NavItem[] {
  return [
    { to: '/today', label: t(locale, 'nav.today'), icon: <House size={18} /> },
    { to: '/plan', label: t(locale, 'nav.plan'), icon: <ClipboardList size={18} /> },
    { to: '/track', label: t(locale, 'nav.track'), icon: <Activity size={18} /> },
    { to: '/progress', label: t(locale, 'nav.progress'), icon: <ChartNoAxesCombined size={18} /> },
    { to: '/profile', label: t(locale, 'nav.profile'), icon: <CircleUserRound size={18} /> },
    { to: '/support', label: t(locale, 'nav.support'), icon: <BadgeHelp size={18} /> },
    ...(role === 'trainer' || role === 'admin' ? [{ to: '/trainer', label: t(locale, 'nav.trainer'), icon: <Users size={18} /> }] : []),
    ...(role === 'admin' ? [{ to: '/admin', label: t(locale, 'nav.admin'), icon: <ShieldCheck size={18} /> }] : [])
  ];
}

function roleLabel(locale: SupportedLocale, role: AuthRole) {
  return t(locale, `shell.role.${role}`);
}

function mobileDockItems(
  items: NavItem[],
  currentItem: NavItem
): NavItem[] {
  const preferredOrder = ['/today', '/plan', '/track', '/progress'];
  const base = preferredOrder
    .map((path) => items.find((item) => item.to === path))
    .filter((item): item is NavItem => Boolean(item));

  if (base.some((item) => item.to === currentItem.to)) {
    return [...base, items.find((item) => item.to === '/profile') ?? currentItem].slice(0, 5);
  }

  return [...base, currentItem].slice(0, 5);
}
