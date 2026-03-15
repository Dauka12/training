import { expect, test } from '@playwright/test';
import { t } from '../frontend/src/shared/i18n';
import { devSeedPassword, login, logout, registerUser, saveDefaultOnboarding, verifyEmail } from './helpers';

test('admin can assign trainer and trainer can review user and create support notifications', async ({ page }) => {
  const email = `member-${Date.now()}@example.com`;
  const password = 'Passw0rd!123';

  const token = await registerUser(page, email, password);
  await verifyEmail(page, token);
  await login(page, email, password);
  await saveDefaultOnboarding(page);
  await logout(page);
  await login(page, 'admin@local.test', devSeedPassword);

  await page.goto('/admin');
  await page.waitForLoadState('networkidle');
  await page.getByLabel(t('ru', 'admin.userEmail')).selectOption(email);
  await page.getByLabel(t('ru', 'admin.trainerEmail')).selectOption('trainer@local.test');
  await page.getByRole('button', { name: t('ru', 'admin.assignTrainer') }).click();

  const equipmentName = `Тестовое оборудование ${Date.now()}`;
  await page.getByLabel(t('ru', 'admin.addEquipment')).fill(equipmentName);
  await page.getByRole('button', { name: t('ru', 'admin.addEquipment') }).click();
  await expect(page.getByText(equipmentName)).toBeVisible();

  const exerciseName = `Тестовое упражнение ${Date.now()}`;
  await page.getByLabel(t('ru', 'admin.addExercise')).fill(exerciseName);
  await page.getByRole('button', { name: t('ru', 'admin.addExercise') }).click();
  await expect(page.getByText(exerciseName)).toBeVisible();

  await logout(page);
  await login(page, email, password);

  await page.goto('/support');
  await page.getByLabel(t('ru', 'support.threadTitle')).fill('Нужна помощь');
  await page.getByLabel(t('ru', 'support.threadBody')).fill('Болит колено после тренировки');
  await Promise.all([
    page.waitForResponse((response) => response.url().endsWith('/api/v1/support/threads') && response.ok()),
    page.getByRole('button', { name: t('ru', 'support.threadCreate') }).click()
  ]);
  await expect(page.getByText('Нужна помощь')).toBeVisible();

  await page.getByRole('button', { name: t('ru', 'support.feed.communityTab') }).click();
  await page.getByLabel(t('ru', 'support.threadTitle')).fill('Завтраки');
  await page.getByLabel(t('ru', 'support.category')).selectOption('nutrition');
  await page.getByLabel(t('ru', 'support.threadBody')).fill('Какие быстрые завтраки подойдут?');
  await Promise.all([
    page.waitForResponse((response) => response.url().endsWith('/api/v1/discussions/threads') && response.ok()),
    page.getByRole('button', { name: t('ru', 'support.discussionCreate') }).click()
  ]);
  await expect(page.getByText('Завтраки')).toBeVisible();

  await logout(page);
  await login(page, 'trainer@local.test', devSeedPassword);

  await page.goto('/trainer');
  await expect(page.getByText(email)).toBeVisible();
  await page.getByRole('button', { name: t('ru', 'trainer.details') }).click();
  await expect(page.getByRole('heading', { name: t('ru', 'trainer.selectedUser') })).toBeVisible();

  await page.goto('/support');
  const supportThread = page.locator('article').filter({ hasText: 'Нужна помощь' }).first();
  await expect(supportThread).toBeVisible();
  await supportThread.getByLabel(t('ru', 'support.replyPlaceholder')).fill('Снизьте нагрузку на этой неделе');
  await Promise.all([
    page.waitForResponse((response) => response.url().includes('/api/v1/support/threads/') && response.url().endsWith('/messages') && response.ok()),
    supportThread.getByRole('button', { name: t('ru', 'support.replyAction') }).click()
  ]);

  await page.getByRole('button', { name: t('ru', 'support.feed.communityTab') }).click();
  const discussionThread = page.locator('article').filter({ hasText: 'Завтраки' }).first();
  await expect(discussionThread).toBeVisible();
  await discussionThread.getByLabel(t('ru', 'support.replyPlaceholder')).fill('Попробуйте овсянку и яйца');
  await Promise.all([
    page.waitForResponse((response) => response.url().includes('/api/v1/discussions/threads/') && response.url().endsWith('/replies') && response.ok()),
    discussionThread.getByRole('button', { name: t('ru', 'support.replyAction') }).click()
  ]);

  await logout(page);
  await login(page, email, password);

  await page.goto('/today');
  await expect(page.getByText('Есть ответ в поддержке')).toBeVisible();
  await expect(page.getByText('Новый ответ в обсуждении')).toBeVisible();
});
