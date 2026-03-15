import { expect, test } from 'vitest';
import { t } from './index';

test('returns clean russian translation', () => {
  expect(t('ru', 'common.save')).toBe('Сохранить');
  expect(t('ru', 'auth.login.title')).toBe('Вход');
});

test('returns kazakh translation', () => {
  expect(t('kk', 'auth.login.title')).toBe('Кіру');
  expect(t('kk', 'common.save')).toBe('Сақтау');
});
