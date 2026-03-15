import { createStore } from 'zustand/vanilla';

export type Locale = 'ru' | 'kk';
export type Theme = 'light' | 'dark';

type PreferencesState = {
  locale: Locale;
  theme: Theme;
  setLocale: (value: Locale) => void;
  setTheme: (value: Theme) => void;
  hydrateFromServer: (value: { locale?: Locale; theme?: Theme }) => void;
};

export function createPreferencesStore() {
  const storage = typeof localStorage === 'undefined' ? null : localStorage;
  const stored = storage?.getItem('preferences');
  const parsed = stored ? (JSON.parse(stored) as { locale?: Locale; theme?: Theme }) : {};

  function persist(locale: Locale, theme: Theme) {
    storage?.setItem('preferences', JSON.stringify({ locale, theme }));
  }

  const store = createStore<PreferencesState>((set, get) => ({
    locale: parsed.locale ?? 'ru',
    theme: parsed.theme ?? 'light',
    setLocale: (value) => {
      set({ locale: value });
      persist(value, get().theme);
    },
    setTheme: (value) => {
      set({ theme: value });
      persist(get().locale, value);
    },
    hydrateFromServer: (value) => {
      const locale = value.locale ?? get().locale;
      const theme = value.theme ?? get().theme;
      set({ locale, theme });
      persist(locale, theme);
    }
  }));

  return store;
}

export const preferencesStore = createPreferencesStore();
