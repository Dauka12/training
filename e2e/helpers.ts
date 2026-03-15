import { expect, type Page } from '@playwright/test';
import { t } from '../frontend/src/shared/i18n';

export const devSeedPassword = 'DevPassw0rd!123';
const locale = 'ru';

export async function registerUser(page: Page, email: string, password: string) {
  await page.goto('/register');
  await page.getByLabel(t(locale, 'auth.email')).fill(email);
  await page.getByLabel(t(locale, 'auth.password')).fill(password);
  await page.getByRole('button', { name: t(locale, 'auth.register.submit') }).click();
  await expect(page.getByText(t(locale, 'auth.success.registered'))).toBeVisible();

  const tokenText = await page.getByText(new RegExp(`${t(locale, 'auth.devToken')}:`, 'i')).textContent();
  const tokenMatch = tokenText?.match(/[a-f0-9]{16,}/i);
  expect(tokenMatch?.[0]).toBeTruthy();
  return tokenMatch![0];
}

export async function verifyEmail(page: Page, token: string) {
  await page.goto('/verify-email');
  await page.getByLabel(t(locale, 'auth.token')).fill(token);
  await page.getByRole('button', { name: t(locale, 'auth.verify.submit') }).click();
  await expect(page.getByText(t(locale, 'auth.success.verified'))).toBeVisible();
}

export async function login(page: Page, email: string, password: string) {
  await page.goto('/login');
  await page.getByLabel(t(locale, 'auth.email')).fill(email);
  await page.getByLabel(t(locale, 'auth.password')).fill(password);
  await page.getByRole('button', { name: t(locale, 'auth.login.submit') }).click();
  await expect(page.getByRole('button', { name: t(locale, 'common.logout') })).toBeVisible();
}

export async function logout(page: Page) {
  await page.getByRole('button', { name: t(locale, 'common.logout') }).click();
  await expect(page.getByRole('heading', { name: t(locale, 'auth.login.title') })).toBeVisible();
}

export async function saveDefaultOnboarding(page: Page) {
  await page.goto('/profile');
  await page.waitForLoadState('networkidle');
  await expect(page.getByRole('heading', { name: t(locale, 'profile.title') })).toBeVisible();

  await page.getByLabel(t(locale, 'profile.age')).fill('28');
  await page.getByLabel(t(locale, 'profile.sex')).selectOption('male');
  await page.getByLabel(t(locale, 'profile.height')).fill('180');
  await page.getByLabel(t(locale, 'profile.weight')).fill('86');
  await page.getByLabel(t(locale, 'profile.targetWeight')).fill('78');
  await page.getByRole('button', { name: t(locale, 'common.next') }).click();

  await expect(page.getByRole('heading', { name: t(locale, 'profile.step.goals') })).toBeVisible();
  await page.getByLabel(t(locale, 'profile.goal')).selectOption('lose_fat');
  await page.getByLabel(t(locale, 'profile.duration')).selectOption('12');
  await page.getByLabel(t(locale, 'profile.experience')).selectOption('beginner');
  await page.getByLabel(t(locale, 'profile.activity')).selectOption('light');
  await page.getByLabel(t(locale, 'profile.location')).selectOption('mixed');
  await page.getByLabel(t(locale, 'profile.timezone')).selectOption('Asia/Qyzylorda');
  await page.getByRole('button', { name: t(locale, 'common.next') }).click();

  await expect(page.getByRole('heading', { name: t(locale, 'profile.step.setup') })).toBeVisible();
  await page.getByRole('checkbox', { name: t(locale, 'weekday.thursday') }).check();
  await page.getByLabel(`${t(locale, 'weekday.thursday')} ${t(locale, 'plan.duration')}`).selectOption('60');
  await page.getByRole('button', { name: t(locale, 'common.next') }).click();

  await expect(page.getByRole('heading', { name: t(locale, 'profile.step.preferences') })).toBeVisible();
  await page.getByLabel(t(locale, 'profile.trainingStyle')).selectOption('balanced_strength');
  await page.getByLabel(t(locale, 'profile.mealStyle')).selectOption('simple_prep');
  await page.getByLabel(t(locale, 'profile.hydrationPreference')).selectOption('regular_small_sips');

  await Promise.all([
    page.waitForResponse((response) => response.url().endsWith('/api/v1/onboarding') && response.ok()),
    page.getByRole('button', { name: t(locale, 'profile.saveOnboarding') }).click()
  ]);
}

export async function generatePlan(page: Page) {
  await page.getByRole('link', { name: t(locale, 'nav.plan') }).first().click();
  await expect(page.getByRole('heading', { name: t(locale, 'plan.title') })).toBeVisible();
  await Promise.all([
    page.waitForResponse((response) => response.url().endsWith('/api/v1/plans/generate') && response.ok()),
    page.getByRole('button', { name: t(locale, 'plan.generate') }).click()
  ]);
  await expect(page.getByText(/12/)).toBeVisible();
}
