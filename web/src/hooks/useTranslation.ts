import { useState } from 'react';
import { en } from '@/locales/en';
import { ja } from '@/locales/ja';

type Locale = 'en' | 'ja';
type LocaleData = typeof en;
type LocaleKey = keyof LocaleData;

const locales: Record<Locale, LocaleData> = {
  en,
  ja,
};

function getInitialLanguage(): Locale {
  try {
    const browserLang = navigator.language;
    if (browserLang.startsWith('ja')) {
      return 'ja';
    }
  } catch (e) {
    // Ignore error (e.g. during SSR if any)
  }
  return 'en';
}

export function useTranslation() {
  const [lang] = useState<Locale>(getInitialLanguage);

  const t = (key: LocaleKey, params?: Record<string, string | number>): string => {
    const template = locales[lang][key] || locales['en'][key] || key;
    if (!params) return template;

    return Object.entries(params).reduce((acc, [k, v]) => {
      return acc.replace(`{${k}}`, String(v));
    }, template);
  };

  return { t, lang };
}
