import { t } from './index';
import { expect, test } from 'vitest';

test('returns kazakh translation', () => {
  expect(t('kk', 'auth.login.title')).toBe('Кіру');
});
