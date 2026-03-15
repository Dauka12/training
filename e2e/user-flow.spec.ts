import { expect, test } from '@playwright/test';
import { generatePlan, login, registerUser, saveDefaultOnboarding, verifyEmail } from './helpers';

test('user can register, verify, login, onboard, generate plan, log water, switch locale and theme', async ({ page }) => {
  const email = `user-${Date.now()}@example.com`;
  const password = 'Passw0rd!123';

  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Все основные функции бесплатны' })).toBeVisible();

  const token = await registerUser(page, email, password);
  await verifyEmail(page, token);
  await login(page, email, password);
  await saveDefaultOnboarding(page);
  await generatePlan(page);

  await page.goto('/today');
  await Promise.all([
    page.waitForResponse((response) => response.url().endsWith('/api/v1/tracking/water') && response.ok()),
    page.getByRole('button', { name: '+500 мл' }).click()
  ]);
  await expect(page.locator('.metric strong')).toHaveText('500');

  await page.getByRole('button', { name: 'RU' }).click();
  await expect(page.getByRole('link', { name: 'Бүгін' })).toBeVisible();

  await page.locator('.topbar__actions .button').nth(1).click();
  await expect(page.locator('.app-shell')).toHaveAttribute('data-theme', 'dark');
});
