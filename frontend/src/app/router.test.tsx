import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { expect, test } from 'vitest';
import { AppRouter } from './router';
import { t } from '../shared/i18n';

test('redirects guest from today route to login screen', () => {
  render(<AppRouter initialEntries={['/today']} initialAuth={{ isAuthenticated: false, role: 'user' }} />);
  expect(screen.getByRole('heading', { name: t('ru', 'auth.login.title') })).toBeInTheDocument();
});

test('shows admin screen for admin role', () => {
  render(<AppRouter initialEntries={['/admin']} initialAuth={{ isAuthenticated: true, role: 'admin' }} />);
  expect(screen.getByRole('heading', { name: t('ru', 'admin.title') })).toBeInTheDocument();
});

test('shows register page for guest route', () => {
  render(<AppRouter initialEntries={['/register']} initialAuth={{ isAuthenticated: false, role: 'user' }} />);
  expect(screen.getByRole('heading', { name: t('ru', 'auth.register.title') })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: t('ru', 'auth.register.submit') })).toBeInTheDocument();
});

test('shows a cohesive auth shell on login page', () => {
  render(<AppRouter initialEntries={['/login']} initialAuth={{ isAuthenticated: false, role: 'user' }} />);

  expect(screen.getByRole('heading', { name: t('ru', 'auth.login.title') })).toBeInTheDocument();
  expect(screen.getByText(t('ru', 'auth.login.lead'))).toBeInTheDocument();
  expect(screen.getAllByText(t('ru', 'landing.free')).length).toBeGreaterThan(0);
  expect(screen.getByText(t('ru', 'landing.benefit.plan'))).toBeInTheDocument();
  expect(screen.getByRole('link', { name: t('ru', 'auth.links.register') })).toBeInTheDocument();
});

test('shows primary authenticated navigation for user', () => {
  render(<AppRouter initialEntries={['/today']} initialAuth={{ isAuthenticated: true, role: 'user', onboardingDone: true }} />);
  expect(screen.getAllByRole('link', { name: t('ru', 'nav.today') }).length).toBeGreaterThan(0);
  expect(screen.getAllByRole('link', { name: t('ru', 'nav.plan') }).length).toBeGreaterThan(0);
  expect(screen.getAllByRole('link', { name: t('ru', 'nav.track') }).length).toBeGreaterThan(0);
  expect(screen.getAllByRole('link', { name: t('ru', 'nav.progress') }).length).toBeGreaterThan(0);
});

test('blocks trainer route for regular user', () => {
  render(<AppRouter initialEntries={['/trainer']} initialAuth={{ isAuthenticated: true, role: 'user', onboardingDone: true }} />);
  expect(screen.getByRole('heading', { name: t('ru', 'common.forbidden') })).toBeInTheDocument();
});

test('redirects regular user without onboarding to profile', () => {
  render(
    <AppRouter
      initialEntries={['/today']}
      initialAuth={{ isAuthenticated: true, role: 'user', onboardingDone: false, email: 'member@example.com' }}
    />
  );

  expect(screen.getByRole('heading', { name: t('ru', 'profile.title') })).toBeInTheDocument();
});

test('does not block admin route when onboarding is incomplete', () => {
  render(
    <AppRouter
      initialEntries={['/admin']}
      initialAuth={{ isAuthenticated: true, role: 'admin', onboardingDone: false, email: 'admin@example.com' }}
    />
  );

  expect(screen.getByRole('heading', { name: t('ru', 'admin.title') })).toBeInTheDocument();
});

test('redirects authenticated user to forced password change screen when required', () => {
  render(
    <AppRouter
      initialEntries={['/today']}
      initialAuth={{ isAuthenticated: true, role: 'admin', onboardingDone: true, email: 'admin@example.com', mustChangePassword: true }}
    />
  );

  expect(screen.getByRole('heading', { name: t('ru', 'auth.changePassword.title') })).toBeInTheDocument();
});

test('renders workspace shell with account summary for authenticated user', () => {
  render(
    <AppRouter
      initialEntries={['/today']}
      initialAuth={{ isAuthenticated: true, role: 'user', onboardingDone: true, email: 'member@example.com' }}
    />
  );

  expect(screen.getByRole('navigation', { name: t('ru', 'shell.navigation') })).toBeInTheDocument();
  expect(screen.getByRole('navigation', { name: t('ru', 'shell.quickNavigation') })).toBeInTheDocument();
  expect(screen.getByTestId('workspace-shell')).toBeInTheDocument();
  expect(screen.getByTestId('workspace-sidebar')).toBeInTheDocument();
  expect(screen.getByTestId('workspace-header')).toBeInTheDocument();
  expect(screen.getAllByText('member@example.com').length).toBeGreaterThan(0);
  expect(screen.getAllByText(t('ru', 'nav.today')).length).toBeGreaterThan(0);
  expect(screen.getAllByRole('link', { name: t('ru', 'nav.today') }).length).toBeGreaterThan(0);
});

test('navigates from today to plan through the shell rail', async () => {
  const user = userEvent.setup();

  render(
    <AppRouter
      initialEntries={['/today']}
      initialAuth={{ isAuthenticated: true, role: 'user', onboardingDone: true, email: 'member@example.com' }}
    />
  );

  await user.click(screen.getAllByRole('link', { name: t('ru', 'nav.plan') })[0]);
  expect(await screen.findByRole('heading', { name: t('ru', 'plan.title') })).toBeInTheDocument();
});

test('landing page exposes the free core-features message and public trust sections', () => {
  render(<AppRouter initialEntries={['/']} initialAuth={{ isAuthenticated: false, role: 'user' }} />);

  expect(screen.getAllByText(t('ru', 'landing.free')).length).toBeGreaterThan(0);
  expect(screen.getByRole('heading', { name: t('ru', 'landing.title') })).toBeInTheDocument();
  expect(screen.getByRole('heading', { name: t('ru', 'landing.how.title') })).toBeInTheDocument();
  expect(screen.getByRole('heading', { name: t('ru', 'landing.faq.title') })).toBeInTheDocument();
  expect(screen.getByText(t('ru', 'landing.privacy'))).toBeInTheDocument();
  expect(screen.getByText(t('ru', 'landing.terms'))).toBeInTheDocument();
  expect(screen.getAllByRole('link', { name: t('ru', 'landing.cta') }).length).toBeGreaterThan(0);
});
