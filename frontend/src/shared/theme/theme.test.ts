import { getInitialTheme } from './theme';
import { expect, test } from 'vitest';

test('prefers stored theme over system', () => {
  localStorage.setItem('theme', 'dark');
  expect(getInitialTheme()).toBe('dark');
});
