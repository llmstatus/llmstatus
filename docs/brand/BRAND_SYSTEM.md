# llmstatus.io — Brand System

**Version**: 0.1
**Last updated**: 2026-04-17
**Philosophy**: This is not another AI tool site. This is the public observatory for AI infrastructure.

> Non-English translations of taglines and UI copy live under `web/i18n/`.
> This document is the canonical English source.

---

## 1. Brand Positioning

### 1.1 One-line definition

> **llmstatus.io — Independent real-time monitoring for the AI infrastructure.**

### 1.2 Long form (About page / media use)

> llmstatus.io continuously observes 40+ AI API providers worldwide.
> We do not scrape their status pages — we make real API calls from
> 7 geographic locations every 30 seconds and publish latency, error
> rate, and quality data as open, verifiable records.
>
> This is a public project. All data is free to access.
> Built and maintained by the llmstatus.io maintainers.

### 1.3 Mental model

We want visitors to associate llmstatus.io with:

- Bloomberg Terminal (authoritative, dense, professional)
- A weather observatory (independent, calm, scientific)
- NASA Mission Control (formal, serious, trustworthy)
- Downdetector (the place everyone knows to go when something breaks)

We do **not** want to be associated with:

- "Yet another AI tool site"
- A SaaS dashboard (we are not a dashboard)
- A clone of status.openai.com (we are an independent observer)
- Colorful, playful data-viz sites

### 1.4 Voice

**How we speak:**

| Is | Is not |
|---|---|
| Calm | Alarmist |
| Data-backed | Opinion-driven |
| Precise | Vague |
| Neutral | Judgmental |
| Epistemically humble | Overclaiming |

**Three copy rules:**

1. **Always state the data before the judgment.**
   - ✅ "OpenAI API error rate climbed to 8.4% between 14:23–15:07 UTC."
   - ❌ "OpenAI is having a terrible outage right now."

2. **Treat every provider identically.**
   - The same phrasing template applies to all providers.
   - No favoritism toward the maintainers' other products. No disparagement of OpenRouter.

3. **Acknowledge methodological limits.**
   - Our data is "observed from 7 locations," not "absolute truth."
   - When something is uncertain, say so.

---

## 2. Design Thesis

### 2.1 Aesthetic lineage

**Control Room Aesthetic**

A blend of:
- The data density of Bloomberg Terminal
- The seriousness of 1970s NASA Mission Control
- The restraint of the Japan Meteorological Agency
- The typographic precision of Swiss Design

### 2.2 Design principles

**1. Dark-first**

Dark is not a style choice for "coolness." It is because:
- Developers work in dim environments (literally)
- Data visualizations are clearer on dark backgrounds
- It is the native language of serious tools

**2. Undecorated**

Do not use:
- Rounded corners (max 2px if needed)
- Gradients
- Soft shadows
- Frosted glass
- Glow effects
- Illustrations
- Emoji (except in extremely specific cases)

**3. Grid-obsessed**

All elements align to a 4px grid.
Tables, dividers, and rulers are visible.
Engineering blueprint, not fine art.

**4. Typography as design**

No hero images, no illustrations, no 3D. Type is the primary visual language.

**5. Data as content**

Numbers, percentages, timestamps, and latency values are first-class citizens.
Copy is second-class. Decoration is zeroth-class (i.e., nonexistent).

---

## 3. Logo System

### 3.1 Concept

The primary mark is a **wordmark** wrapped in square brackets:

```
[ llmstatus ]
```

**Meaning of the brackets:**
- Developer language (array index, regex, JSON arrays)
- A sense of observation / sampling / measurement
- A frame, a viewport, a boundary

**All-lowercase:**
- Downplays "brandiness," emphasizes "tool-ness"
- Reads like a CLI command or an import path
- Does not shout — it observes calmly

### 3.2 Logo variants

**Primary Logomark:**
```
[ llmstatus ]
```
Used for: site header, social banners, brand materials, OG images.

**Compact Mark:**
```
[ ls ]
```
Used for: favicon, app icon, small-format applications, social avatars.

**Status Pulse Mark:**
```
[ ● ]
```
Used for: badges, loading indicators, animated tab favicon.

**Primary Wordmark (plain text):**
```
llmstatus
```
Used for: plaintext credit, footer, API response headers (e.g. `X-Powered-By`).

### 3.3 Logo color rules

- **Dark background (default):** wordmark in `--ink-100` (near-white), brackets in `--signal-amber`.
- **Light background (rare):** everything in `--ink-900` (near-black).
- **Single-color print:** all black or all white, no accent.

### 3.4 Clearspace

Reserve at least half the logo's height as blank space on every side. Nothing may enter this zone.

### 3.5 Prohibited uses

- ❌ Do not rotate the logo
- ❌ Do not change colors (except the three cases in §3.3)
- ❌ Do not add effects (glow, shadow, gradient)
- ❌ Do not stretch or distort
- ❌ Do not redraw the logo in a non-monospace typeface
- ❌ Do not place it on low-contrast backgrounds

---

## 4. Color System

All colors are defined as CSS variables. Never hard-code color values.

### 4.1 Canvas (backgrounds)

```css
--canvas-base:    #0A0E14;  /* primary background — deep ink-blue-black */
--canvas-raised:  #0F141B;  /* cards and panels — one step above base */
--canvas-sunken:  #060A10;  /* sunken elements, code blocks */
--canvas-overlay: #151B24;  /* hover and selected state backgrounds */
```

**Why not pure black:**
Pure `#000000` is harsh on OLED screens and looks cheap on regular displays.
Deep ink-blue-black `#0A0E14` has a cool cast — like an observatory, like Bloomberg Terminal.

### 4.2 Ink (text and borders)

```css
--ink-100: #E8ECF1;  /* primary text, headings */
--ink-200: #C1C8D4;  /* secondary text */
--ink-300: #8B94A3;  /* supporting text, labels */
--ink-400: #5A6472;  /* tertiary, metadata */
--ink-500: #3C4452;  /* dividers, borders */
--ink-600: #242B36;  /* faint dividers */
```

### 4.3 Signals (status and emphasis)

```css
/* Normal — metallic cyan */
--signal-ok:      #5EEAD4;  /* operational */
--signal-ok-bg:   #0F3530;  /* matching tinted background */

/* Warning — amber */
--signal-warn:    #F59E0B;  /* elevated latency, degraded */
--signal-warn-bg: #3D2909;

/* Critical — coral red */
--signal-down:    #F87171;  /* outage, severe issue */
--signal-down-bg: #3A1515;

/* Brand accent — electric amber */
--signal-amber:   #D97706;  /* brand color, logo brackets, highlights */
```

**Why these four colors:**

- **Metallic cyan `#5EEAD4`** — not the usual "playground green." Calm, scientific, like an oscilloscope trace.
- **Amber `#F59E0B`** — the color of vintage CRT displays. Nostalgic and information-dense.
- **Coral red `#F87171`** — not alarm red `#FF0000` (too aggressive). A "restrained bad."
- **Electric amber `#D97706`** — brand accent. Evokes old terminals, old oscilloscopes, console indicator lamps.

### 4.4 Data-viz palette

For charts, leaderboards, maps, etc.:

```css
--viz-1: #5EEAD4;   /* metallic cyan */
--viz-2: #60A5FA;   /* steel blue */
--viz-3: #F59E0B;   /* amber */
--viz-4: #A78BFA;   /* lavender (the only purple, reserved for data viz) */
--viz-5: #FB923C;   /* orange */
--viz-6: #34D399;   /* emerald */
--viz-7: #F472B6;   /* pink (use sparingly, only for many-category charts) */
```

### 4.5 Color usage ratios

A well-designed llmstatus.io page should look roughly:

- **85%** Canvas (dark background)
- **12%** Ink (text, borders)
- **2%** Signal OK (normal state)
- **1%** Signal Warn / Down or brand accent

**Warning:** If your page looks colorful, you did it wrong.

---

## 5. Typography

### 5.1 Type choices

**Display & Logo:**
- **Primary**: `Berkeley Mono` (commercial license, ~$75 one-time)
- **Fallback 1**: `MD IO` (free for commercial use)
- **Fallback 2**: `JetBrains Mono` (free, Google Fonts)
- **Strictly forbidden**: Inter, Space Grotesk, Roboto (these are "AI-slop fonts")

**Body:**
- **Primary**: `IBM Plex Sans` (free, Google Fonts)
- **Fallback**: `Söhne` or `Untitled Sans` (commercial)
- **Strictly forbidden**: Inter (same reason as above)

**Numeric / Code:**
- **Primary**: `Berkeley Mono` (same family as display for consistency)
- **Fallback**: `JetBrains Mono`

### 5.2 Why Berkeley Mono / MD IO

Both typefaces carry a strong "console / drafting / laboratory logbook" character. Their letterforms evoke:
- 1980s terminals
- Annotations on engineering drawings
- Laboratory-equipment nameplates

That is precisely the tone we want.

### 5.3 Size scale

```css
--text-xs:    11px;   /* metadata, timestamps */
--text-sm:    13px;   /* secondary text, table content */
--text-base:  14px;   /* body (intentionally 2px smaller than typical web) */
--text-md:    16px;   /* emphasized body */
--text-lg:    20px;   /* small headings */
--text-xl:    28px;   /* section headings */
--text-2xl:   40px;   /* page titles */
--text-3xl:   64px;   /* home hero */
```

**Note:** The 14px base is deliberate — it matches developer-tool defaults (VS Code 14px, GitHub 14px, Vercel 14px). It silently signals "this is for developers."

### 5.4 Weights

Only three weights are used:

```css
--weight-regular: 400;
--weight-medium:  500;
--weight-bold:    700;
```

Forbidden: 300 (light), 900 (black). They break the restraint.

### 5.5 Line-height

```css
--leading-tight:  1.2;   /* headings */
--leading-normal: 1.5;   /* body */
--leading-loose:  1.7;   /* long-form articles */
```

### 5.6 Letter-spacing

```css
--tracking-tight:  -0.02em;  /* large display */
--tracking-normal: 0;
--tracking-wide:   0.05em;   /* ALL-CAPS labels like "LIVE", "CHINA VIEW" */
```

---

## 6. UI Component Specifications

### 6.1 StatusPill

```
●  OPERATIONAL    ← dot + gap + all-caps label
●  DEGRADED
●  DOWN
```

Rules:
- Dot diameter 8px
- 8px gap between dot and text
- Text 11px, all-caps, 0.05em tracking
- No rounded background pill (that is a commercial-SaaS pattern)
- Text color inherits from the status color

### 6.2 Data table

```
╭────────────────────────────────────────────────────╮
│ Provider          p95      Uptime 24h   Status    │
├────────────────────────────────────────────────────┤
│ OpenAI            680ms    99.98%       ●         │
│ Anthropic       1,240ms    99.67%       ●         │
│ Google Gemini     920ms   100.00%       ●         │
╰────────────────────────────────────────────────────╯
```

Rules:
- Dividers use `--ink-600` (faint)
- Numeric columns right-aligned
- Numbers in Berkeley Mono (monospace)
- Text columns left-aligned
- Header text one size smaller than body (e.g. body 14px, header 11px)
- Headers all-caps, with widened letter-spacing
- Hover highlights the full row with `--canvas-overlay`

### 6.3 Buttons

**Primary** — used rarely, only for critical CTAs.
```
┌──────────────┐
│  SUBSCRIBE   │
└──────────────┘
```
- Background `--signal-amber`
- Black text
- No rounded corners
- ALL CAPS

**Secondary** — the default button.
```
┌──────────────┐
│  View incident →
└──────────────┘
```
- Transparent background
- Border `--ink-500`
- Text `--ink-100`
- Hover: border becomes `--ink-300`

### 6.4 Timestamp component

Always show "time since" rather than absolute time:

```
Last checked 23s ago
Incident started 4h 12m ago
Resolved 2d ago
```

Rules:
- Color `--ink-400` (supporting)
- 11px
- Auto-refresh every 10s on the client

### 6.5 Grid background (optional decoration)

A very faint grid background is permitted on the home hero:

```css
background-image:
  linear-gradient(to right, rgba(255,255,255,0.02) 1px, transparent 1px),
  linear-gradient(to bottom, rgba(255,255,255,0.02) 1px, transparent 1px);
background-size: 8px 8px;
```

**Extremely restrained.** No more than one section.

### 6.6 Prohibited component patterns

- ❌ Toast notifications (break the restraint)
- ❌ Floating chatbot (the opposite tone)
- ❌ Cookie banners that block content (use the smallest variant when legally required)
- ❌ "Sign up for newsletter" pop-ups (subscription lives only in the footer)
- ❌ Rainbow-colored progress bars
- ❌ Large gradient backgrounds

---

## 7. Copy Library

The following copy is ready to use verbatim.

### 7.1 Home

**Hero tagline:**
> Independent real-time monitoring
> for the AI infrastructure.

**Hero subhead:**
> Measured from 7 global locations.
> Not scraped from official status pages.

**Live indicator:**
> Last checked 23 seconds ago · Next check in 37s

### 7.2 About page

**Opening:**
> llmstatus.io observes 40+ AI API providers in real time.
>
> We measure what actually happens when you call these APIs,
> from seven geographic locations. We do not scrape official
> status pages — we make real requests and report what we see.
>
> This is a public project. Data is free to access, free to
> embed, free to cite.

**Methodology hero (methodology page):**
> How we measure:
> — Real API calls, not scraped status pages
> — From 7 geographic locations, including mainland China
> — Every 30–60 seconds, 24/7
> — We publish our probe logic as open source
> — We monitor ourselves too — any service operated by the maintainers, probed by the same rules

### 7.3 Subscribe / CTA

Never use marketing speak. Use facts.

- ✅ "Get notified when providers are down."
- ❌ "Never miss a critical incident! Join 10k+ developers!"

### 7.4 Incident copy templates

**Incident detected (agent-generated):**
> [{provider}] Elevated error rate detected.
> Our probes from {regions} observed a {X}% error rate between
> {start_time} and {latest_time}, compared to a baseline of {Y}%.
> Affected models: {list}.
> We are continuing to monitor.

**Incident resolved:**
> [{provider}] Incident resolved.
> Error rate returned to baseline at {time}.
> Total duration: {X} minutes.
> Peak error rate: {Y}%.

### 7.5 Banned vocabulary

These words must never appear anywhere on llmstatus.io:

| Banned | Reason |
|---|---|
| "leading" / "best-in-class" | Marketing filler |
| "revolutionary" / "game-changing" | Marketing filler |
| "AI-powered" / "powered by AI" | Everything is AI now; it says nothing |
| "seamless" / "seamlessly" | No customer ever describes their experience this way |
| "unlock" | Unless it is literally a lock |
| "delight" / "delightful" | Developers do not use this word |
| "elevate" / "empower" | Management-consulting vocabulary |
| "journey" | Unless it is literally a journey |

### 7.6 Recommended vocabulary

Words that carry our tone:

| Context | Preferred words |
|---|---|
| Observation verbs | observe, measure, probe, sample, detect |
| Describing anomalies | elevated, degraded, disruption, irregularity |
| Describing normal | operational, nominal, within baseline |
| Describing data | recorded, reported, logged, captured |
| Intensifiers | substantially, marginally, briefly |

---

## 8. Templates

### 8.1 OG Image template (social sharing)

Fixed layout, generated dynamically:

```
┌──────────────────────────────────────────────────┐
│                                                    │
│  [ llmstatus ]                                    │
│                                                    │
│  OpenAI                                           │
│  ●  OPERATIONAL                                   │
│                                                    │
│  p95 latency      680 ms                          │
│  Uptime 24h       99.98 %                         │
│  Uptime 7d        99.94 %                         │
│                                                    │
│  llmstatus.io · April 17, 2026                    │
│                                                    │
└──────────────────────────────────────────────────┘
```

1200×630 px, dark background, Berkeley Mono.

### 8.2 Twitter/X post template (human-posted)

```
[llmstatus]
OpenAI elevated errors detected.

8.4% error rate from 14:23–15:07 UTC
(baseline: <0.1%)

Affected: gpt-4o, o1-mini
Regions impacted: all

Details: llmstatus.io/incidents/...
```

### 8.3 Footer credit

```
A public project by the llmstatus.io maintainers.
All data is free to access, embed, and cite.
```

---

## 9. Transparency & Trust-building

Because the product we ship is *credibility*, part of the brand is **auditability**.

- ✅ Publish the core probe code (adapters) as open source
- ✅ Publish the methodology document
- ✅ Publish monthly aggregate CSV downloads
- ✅ Credit "Built by the llmstatus.io maintainers" explicitly in the footer
- ✅ Disclose on the About page that AI agents assist operations
- ✅ Show our own monitoring data for any self-monitored service — including failures

---

## 10. Brand Don'ts

**Never, under any circumstance:**

1. **Never advertise the maintainers' other products on llmstatus.io.**
   - Attribution, if any, lives only in the footer.
   - When an article naturally mentions a product operated by the maintainers, it must also name OpenRouter / LiteLLM / Portkey.
   - The site must never favor a self-monitored service.

2. **Never do a "marketing-style launch."**
   - No "We're excited to announce!" posts.
   - New features ship by updating the changelog, nothing more.

3. **Never cherry-pick data.**
   - When our own monitoring data looks bad, we show it.
   - When a competitor looks good, we show it.

4. **Never accept advertising.**
   - No ad slots on the site.
   - No paid placements, ever.
   - This is core to the "public good" positioning.

5. **Never lean on memes or internet slang.**
   - No emoji pile-ups.
   - No "banger," "slaps," "GOATed."
   - The tone is a serious scientist, not a Twitter personality.

---

## 11. Visual reference library (moodboard)

Show a designer these references to convey the tone:

- **Bloomberg Terminal** screenshots
- **Berkeley Graphics** type specimens
- **NASA Mission Control** photographs from the 1970s–80s
- **Japan Meteorological Agency** website (jma.go.jp)
- **Swiss Style** classic posters
- **MIT Media Lab** academic visual identity
- **Werner's Nomenclature of Colours** book typography
- **Braun Design** product manuals (Dieter Rams era)

**Do not reference:**
- Any Vercel template
- Any Dribbble popular shot
- Any typical SaaS landing page
- Any AI-product home page (OpenAI and Anthropic excepted)

---

End of BRAND_SYSTEM.md — Version 0.1
