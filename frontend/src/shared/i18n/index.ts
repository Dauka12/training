import { kk } from './kk';
import { ru } from './ru';

const dictionary = { ru, kk } as const;

export type SupportedLocale = keyof typeof dictionary;

export function t(locale: SupportedLocale, key: string): string {
  return dictionary[locale][key] ?? dictionary.ru[key] ?? key;
}
