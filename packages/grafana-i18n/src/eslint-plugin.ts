import noTranslationTopLevel from './eslint-plugin/no-translation-top-level/no-translation-top-level';
import noUntranslatedStrings from './eslint-plugin/no-untranslated-strings/no-untranslated-strings';

export default {
  rules: {
    'no-untranslated-strings': noUntranslatedStrings,
    'no-translation-top-level': noTranslationTopLevel,
  },
};
