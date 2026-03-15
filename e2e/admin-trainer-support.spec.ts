import { expect, test } from '@playwright/test';
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
  await page.getByLabel('Email пользователя').fill(email);
  await page.getByLabel('Email тренера').fill('trainer@local.test');
  await page.getByRole('button', { name: 'Назначить тренера' }).click();

  const equipmentName = `Тестовое оборудование ${Date.now()}`;
  await page.getByLabel('Добавить оборудование').fill(equipmentName);
  await page.getByRole('button', { name: 'Добавить оборудование' }).click();
  await expect(page.getByText(equipmentName)).toBeVisible();

  const exerciseName = `Тестовое упражнение ${Date.now()}`;
  await page.getByLabel('Добавить упражнение').fill(exerciseName);
  await page.getByRole('button', { name: 'Добавить упражнение' }).click();
  await expect(page.getByText(exerciseName)).toBeVisible();

  await logout(page);
  await login(page, email, password);

  await page.goto('/support');
  await page.getByLabel('Тема').first().fill('Нужна помощь');
  await page.getByLabel('Сообщение').first().fill('Болит колено после тренировки');
  await page.getByRole('button', { name: 'Создать тикет' }).click();
  await expect(page.getByText('Нужна помощь')).toBeVisible();

  await page.getByLabel('Тема').nth(1).fill('Завтраки');
  await page.getByLabel('Категория').fill('nutrition');
  await page.getByLabel('Сообщение').nth(1).fill('Какие быстрые завтраки подойдут?');
  await page.getByRole('button', { name: 'Создать обсуждение' }).click();
  await expect(page.getByText('Завтраки')).toBeVisible();

  await logout(page);
  await login(page, 'trainer@local.test', devSeedPassword);

  await page.goto('/trainer');
  await expect(page.getByText(email)).toBeVisible();
  await page.getByRole('button', { name: 'Подробнее' }).click();
  await expect(page.getByRole('heading', { name: 'Пользователь' })).toBeVisible();

  await page.goto('/support');
  const supportThread = page.locator('article').filter({ hasText: 'Нужна помощь' }).first();
  await expect(supportThread).toBeVisible();
  await supportThread.getByPlaceholder('Ответ').fill('Снизьте нагрузку на этой неделе');
  await supportThread.getByRole('button', { name: 'Ответить' }).click();

  const discussionThread = page.locator('article').filter({ hasText: 'Завтраки' }).first();
  await expect(discussionThread).toBeVisible();
  await discussionThread.getByPlaceholder('Ответ').fill('Попробуйте овсянку и яйца');
  await discussionThread.getByRole('button', { name: 'Ответить' }).click();

  await logout(page);
  await login(page, email, password);

  await page.goto('/today');
  await expect(page.getByText('Есть ответ в поддержке')).toBeVisible();
  await expect(page.getByText('Новый ответ в обсуждении')).toBeVisible();
});
