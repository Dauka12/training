import { useEffect } from 'react';
import { useStore } from 'zustand';
import { Languages, MoonStar, PlaySquare, Search, SunMedium } from 'lucide-react';
import { BrowserAppRouter } from './router';
import { apiRequest } from '../shared/api/client';
import { useAuthStore } from '../shared/auth/store';
import { preferencesStore } from '../shared/preferences/store';
import { t } from '../shared/i18n';

export function App() {
  const locale = useStore(preferencesStore, (state) => state.locale);
  const theme = useStore(preferencesStore, (state) => state.theme);
  const setLocale = useStore(preferencesStore, (state) => state.setLocale);
  const setTheme = useStore(preferencesStore, (state) => state.setTheme);
  const hydrateFromServer = useStore(preferencesStore, (state) => state.hydrateFromServer);
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const setAuthenticated = useAuthStore((state) => state.setAuthenticated);
  const clear = useAuthStore((state) => state.clear);

  useEffect(() => {
    let cancelled = false;

    apiRequest<{ email: string; locale: 'ru' | 'kk'; theme: 'light' | 'dark'; roles: string[]; onboarding_done: boolean }>('/me')
      .then((response) => {
        if (!cancelled) {
          setAuthenticated({
            role: (response.roles[0] as 'user' | 'trainer' | 'admin') ?? 'user',
            email: response.email,
            onboardingDone: response.onboarding_done
          });
          hydrateFromServer({ locale: response.locale, theme: response.theme });
        }
      })
      .catch(() => {
        if (!cancelled) {
          clear();
        }
      });

    return () => {
      cancelled = true;
    };
  }, [clear, hydrateFromServer, setAuthenticated]);

  function persistServerPreferences(nextLocale: 'ru' | 'kk', nextTheme: 'light' | 'dark') {
    if (!isAuthenticated) {
      return;
    }
    void apiRequest('/me/preferences', {
      method: 'PUT',
      body: JSON.stringify({
        locale: nextLocale,
        theme: nextTheme,
        water_override_ml: 0
      })
    }).catch(() => undefined);
  }

  function handleLocaleToggle() {
    const nextLocale = locale === 'ru' ? 'kk' : 'ru';
    setLocale(nextLocale);
    persistServerPreferences(nextLocale, theme);
  }

  function handleThemeToggle() {
    const nextTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(nextTheme);
    persistServerPreferences(locale, nextTheme);
  }

  return (
    <div data-theme={theme} className="app-shell">
      <header className="topbar">
        <div className="topbar__brand">
          <div className="topbar__brand-mark" aria-hidden="true">
            <PlaySquare size={18} />
          </div>
          <div>
            <strong>{t(locale, 'brand.name')}</strong>
            <span className="muted">{t(locale, 'shell.workspace')}</span>
          </div>
        </div>
        <div className="topbar__center">
          <div className="topbar__search" aria-hidden="true">
            <Search size={16} />
            <span className="muted">{t(locale, 'shell.searchPlaceholder')}</span>
          </div>
        </div>
        <div className="topbar__actions">
          <button type="button" className="button" onClick={handleLocaleToggle}>
            <Languages size={18} aria-hidden="true" />
            {locale.toUpperCase()}
          </button>
          <button type="button" className="button" onClick={handleThemeToggle}>
            {theme === 'light' ? <SunMedium size={18} aria-hidden="true" /> : <MoonStar size={18} aria-hidden="true" />}
            {t(locale, `theme.${theme}`)}
          </button>
        </div>
      </header>
      <BrowserAppRouter locale={locale} />
    </div>
  );
}
