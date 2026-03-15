import { expect, type Page } from '@playwright/test';

export const devSeedPassword = 'DevPassw0rd!123';

export async function registerUser(page: Page, email: string, password: string) {
  await page.goto('/register');
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Пароль').fill(password);
  await page.getByRole('button', { name: 'Создать аккаунт' }).click();
  await expect(page.getByText('Аккаунт создан. Подтвердите email.')).toBeVisible();

  const tokenText = await page.getByText(/Dev-токен:/).textContent();
  const tokenMatch = tokenText?.match(/[a-f0-9]{16,}/i);
  expect(tokenMatch?.[0]).toBeTruthy();
  return tokenMatch![0];
}

export async function verifyEmail(page: Page, token: string) {
  await page.goto('/verify-email');
  await page.getByLabel('Токен').fill(token);
  await page.getByRole('button', { name: 'Подтвердить' }).click();
  await expect(page.getByText('Email подтвержден.')).toBeVisible();
}

export async function login(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Пароль').fill(password);
  await page.getByRole('button', { name: 'Войти' }).click();
  await expect(page.getByRole('button', { name: 'Выйти' })).toBeVisible();
}

export async function logout(page: Page) {
  await page.getByRole('button', { name: 'Выйти' }).click();
  await expect(page.getByRole('heading', { name: 'Вход' })).toBeVisible();
}

export async function saveDefaultOnboarding(page: Page) {
  await page.goto('/profile');
  await page.waitForLoadState('networkidle');
  await expect(page.getByRole('heading', { name: 'Профиль и настройки' })).toBeVisible();

  await page.getByLabel('Возраст').fill('28');
  await page.getByLabel('Биологический пол').fill('male');
  await page.getByLabel('Рост, см').fill('180');
  await page.getByLabel('Текущий вес, кг').fill('86');
  await page.getByLabel('Целевой вес, кг').fill('78');
  await page.getByLabel('Главная цель').fill('lose_fat');
  await page.getByLabel('Длительность программы, недели').fill('12');
  await page.getByLabel('Уровень опыта').fill('beginner');
  await page.getByLabel('Дневная активность').fill('light');
  await page.getByLabel('Где тренируетесь').fill('mixed');
  await page.getByLabel('Часовой пояс').fill('Asia/Qyzylorda');
  await page.getByLabel('Предпочитаемый стиль тренировок').fill('balanced_strength');
  await page.getByLabel('Предпочитаемый стиль питания').fill('simple_prep');
  await page.getByLabel('Предпочтение по воде').fill('regular_small_sips');

  await Promise.all([
    page.waitForResponse((response) => response.url().endsWith('/api/v1/onboarding') && response.ok()),
    page.getByRole('button', { name: 'Сохранить онбординг' }).click()
  ]);
}

export async function generatePlan(page: Page) {
  await page.goto('/plan');
  await Promise.all([
    page.waitForResponse((response) => response.url().endsWith('/api/v1/plans/generate') && response.ok()),
    page.getByRole('button', { name: 'Сгенерировать план' }).click()
  ]);
  await expect(page.getByText('План на 12 недель')).toBeVisible();
}
