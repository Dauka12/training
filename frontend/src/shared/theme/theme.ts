export type ThemeMode = 'light' | 'dark';

export function getInitialTheme(): ThemeMode {
  const stored = localStorage.getItem('theme');
  return stored === 'dark' ? 'dark' : 'light';
}
