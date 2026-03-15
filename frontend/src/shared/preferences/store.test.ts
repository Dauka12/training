import { expect, test } from 'vitest';
import { createPreferencesStore } from './store';

test('persists locale and theme', () => {
  localStorage.clear();
  const store = createPreferencesStore();
  store.getState().setLocale('kk');
  store.getState().setTheme('dark');

  expect(store.getState().locale).toBe('kk');
  expect(store.getState().theme).toBe('dark');
  expect(localStorage.getItem('preferences')).toContain('"locale":"kk"');
});

test('hydrates locale and theme from server profile', () => {
  localStorage.clear();
  const store = createPreferencesStore();

  store.getState().hydrateFromServer({ locale: 'kk', theme: 'dark' });

  expect(store.getState().locale).toBe('kk');
  expect(store.getState().theme).toBe('dark');
  expect(localStorage.getItem('preferences')).toContain('"theme":"dark"');
});
