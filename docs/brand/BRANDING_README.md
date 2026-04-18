# Brand Assets

Everything needed to render the llmstatus.io brand.

## Files in this directory

| File | Purpose |
|---|---|
| `BRAND_SYSTEM.md` | Brand positioning, voice, color system, typography, copy rules |
| `preview.html` | Visual mockup of homepage, china view, badges, colors, type |
| `logo-primary.svg` | Primary wordmark `[ llmstatus ]` — use for header, OG image, marketing |
| `logo-compact.svg` | Compact mark `[ ls ]` — use for favicon, small sizes, app icons |
| `logo-pulse.svg` | Pulse mark `[ ● ]` with animation — use for live indicators, loading states |

## How to use

### 1. Review the visual system
Open `preview.html` in any browser to see the design system rendered live.

### 2. Consult `BRAND_SYSTEM.md`
Anyone writing copy, designing a page, or producing a report must reference
this document. Section 7 (Copy Library) contains ready-to-use text.

### 3. Logos
- All three SVGs use CSS variables, so colors can be overridden via CSS.
- Replace `JetBrains Mono` with `Berkeley Mono` in production if a license is purchased.

## Non-negotiable principles

1. **Monitor our own services honestly.** If a self-monitored service breaks, show it breaking.
2. **Never scrape official status pages.** Only real API probes from our nodes.
3. **Never advertise the maintainers' other products on llmstatus.io.** Footer attribution only.
4. **Never accept ads.** This is a public project, not a SaaS.
5. **Publish methodology.** Every metric must be reproducible.

---

Version 0.1 · 2026-04-17
