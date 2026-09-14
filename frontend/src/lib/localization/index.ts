export interface StringTable extends Record<string, null | string | StringTable> {}

const localeFiles = import.meta.glob<StringTable>('./locales/*.json', { eager: true, import: 'default' });

export const i18n: Record<string, StringTable> = Object.fromEntries(
  Object.entries(localeFiles).map(([path, content]) => {
    const lang = path.match(/([^/]+)\.json$/)?.[1] ?? path;
    return [lang, content];
  }),
);

function allStrings(stringTable: StringTable): (string | null)[] {
  return Object.values(stringTable).map((entry) => {
    if (!entry) {
      return [null];
    }
    if (typeof entry === 'string') {
      return [entry];
    }
    return allStrings(entry);
  }).flat();
}

export const languages = Object.entries(i18n).map(([lang, stringTable]) => {
  const strings = allStrings(stringTable);
  return {
    languageCode: lang,
    stringTable,
    name: new Intl.DisplayNames([lang], { type: 'language' }).of(lang) ?? lang,
    // TODO flag
    completeness: strings.filter((s) => !!s).length / strings.length,
  };
}).filter((lang) => lang.completeness > 0);
