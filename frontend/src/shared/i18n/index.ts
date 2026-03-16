import { kk } from './kk';
import { ru } from './ru';

const overrides: Record<'ru' | 'kk', Record<string, string>> = {
  ru: {
    'auth.changePassword.title': 'Смените пароль',
    'auth.changePassword.submit': 'Обновить пароль',
    'auth.changePassword.lead': 'Это первый вход с временным паролем. Сначала задайте новый пароль, затем откроется полный доступ.',
    'auth.currentPassword': 'Текущий пароль',
    'auth.success.passwordChanged': 'Пароль обновлён.',
    'exercise.technique': 'Техника',
    'exercise.location': 'Где выполнять',
    'exercise.equipment': 'Оборудование',
    'exercise.substitutions': 'Замены',
    'exercise.contraindications': 'Ограничения',
    'admin.previewWger': 'Предпросмотр Wger',
    'admin.previewSamples': 'Предпросмотр импорта',
    'admin.exerciseEditor': 'Редактировать упражнение',
    'admin.exerciseMedia': 'Фото или media URL',
    'admin.exerciseTechnique': 'Техника упражнения',
    'admin.exerciseContraindications': 'Противопоказания',
    'admin.exerciseSubstitutions': 'ID замен через запятую',
    'admin.saveExerciseMeta': 'Сохранить описание'
  },
  kk: {
    'auth.changePassword.title': 'Құпиясөзді ауыстыру',
    'auth.changePassword.submit': 'Құпиясөзді жаңарту',
    'auth.changePassword.lead': 'Бұл уақытша құпиясөзбен алғашқы кіру. Алдымен жаңа құпиясөз орнатып, содан кейін толық қолжетімділік ашылады.',
    'auth.currentPassword': 'Ағымдағы құпиясөз',
    'auth.success.passwordChanged': 'Құпиясөз жаңартылды.',
    'exercise.technique': 'Техника',
    'exercise.location': 'Орындау орны',
    'exercise.equipment': 'Жабдық',
    'exercise.substitutions': 'Балама жаттығулар',
    'exercise.contraindications': 'Шектеулер',
    'admin.previewWger': 'Wger алдын ала қарау',
    'admin.previewSamples': 'Импорт үлгілері',
    'admin.exerciseEditor': 'Жаттығуды өңдеу',
    'admin.exerciseMedia': 'Фото немесе media URL',
    'admin.exerciseTechnique': 'Жаттығу техникасы',
    'admin.exerciseContraindications': 'Қарсы көрсетілімдер',
    'admin.exerciseSubstitutions': 'Балама ID, үтірмен',
    'admin.saveExerciseMeta': 'Сипаттаманы сақтау'
  }
};

const dictionary: Record<'ru' | 'kk', Record<string, string>> = {
  ru: { ...ru, ...overrides.ru },
  kk: { ...kk, ...overrides.kk }
};

export type SupportedLocale = keyof typeof dictionary;

export function t(locale: SupportedLocale, key: string): string {
  return dictionary[locale][key] ?? dictionary.ru[key] ?? key;
}
