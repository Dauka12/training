import { render, screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import { LoginForm, RegisterForm } from './forms';
import { t } from '../../shared/i18n';

test('renders localized login form', () => {
  render(<LoginForm locale="ru" onSubmit={async () => undefined} />);
  expect(screen.getByRole('heading', { name: t('ru', 'auth.login.title') })).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'auth.email'))).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'auth.password'))).toBeInTheDocument();
  expect(screen.getByRole('link', { name: t('ru', 'auth.google.continue') })).toBeInTheDocument();
});

test('renders localized register form', () => {
  render(<RegisterForm locale="ru" onSubmit={async () => undefined} />);
  expect(screen.getByRole('heading', { name: t('ru', 'auth.register.title') })).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'auth.email'))).toBeInTheDocument();
  expect(screen.getByLabelText(t('ru', 'auth.password'))).toBeInTheDocument();
  expect(screen.getByRole('button', { name: t('ru', 'auth.register.submit') })).toBeInTheDocument();
  expect(screen.getByRole('link', { name: t('ru', 'auth.google.continue') })).toBeInTheDocument();
});
