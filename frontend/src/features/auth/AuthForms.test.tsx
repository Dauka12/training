import { render, screen } from '@testing-library/react';
import { expect, test } from 'vitest';
import { LoginForm, RegisterForm } from './forms';

test('renders localized login form', () => {
  render(<LoginForm locale="ru" onSubmit={async () => undefined} />);
  expect(screen.getByRole('heading', { name: 'Вход' })).toBeInTheDocument();
  expect(screen.getByLabelText('Email')).toBeInTheDocument();
  expect(screen.getByLabelText('Пароль')).toBeInTheDocument();
});

test('renders localized register form', () => {
  render(<RegisterForm locale="ru" onSubmit={async () => undefined} />);
  expect(screen.getByRole('heading', { name: 'Регистрация' })).toBeInTheDocument();
  expect(screen.getByLabelText('Email')).toBeInTheDocument();
  expect(screen.getByLabelText('Пароль')).toBeInTheDocument();
  expect(screen.getByRole('button', { name: 'Создать аккаунт' })).toBeInTheDocument();
});
