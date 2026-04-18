# Internationalization

This directory is the **only** place in the repository where non-English
content may live. Every other source file, doc, comment, and commit
message stays in English — see `CONTRIBUTING.md`.

## Layout

```
web/i18n/
├── en.json                 # English source of truth
├── zh-Hans.json            # Simplified Chinese
├── zh-Hant.json            # Traditional Chinese
└── ...                     # other locales added as needed
```

Keys mirror the English source. Missing keys fall back to `en.json`.

## Adding a locale

1. Copy `en.json` to `<locale>.json`.
2. Translate each value; leave the keys alone.
3. Register the locale in `web/app/i18n.ts` (created in LLMS-005+).
4. Open a PR titled `docs(i18n): add <locale> translation`.

## What belongs here

- UI strings shown to end-users.
- Error messages surfaced in the browser.
- Locale-specific number / date formatting configuration.

## What does NOT belong here

- Code comments
- README content
- API error messages consumed by machines
- Provider names (they are proper nouns, not translated)
- Metric names, model names, region identifiers
