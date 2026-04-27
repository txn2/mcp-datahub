---
version: alpha
spec: https://github.com/google-labs-code/design.md
name: mcp-datahub-docs
description: Local design adoption record for the mcp-datahub.txn2.com documentation site. References txn2/www DESIGN.md as the canonical visual identity for tokens, typography, components, voice, and accessibility rules. Records only the decisions and MkDocs Material learnings that the canonical does not cover.
upstream:
  design: https://github.com/txn2/www/blob/master/DESIGN.md
  tokens: https://github.com/txn2/www/blob/master/tokens.json
  sister: https://github.com/txn2/mcp-s3/blob/main/DESIGN.md
adoption: token-alignment
stack:
  generator: MkDocs
  theme: Material for MkDocs
  templates: docs/overrides/
  styles: docs/stylesheets/extra.css
---

## What is canonical

The canonical visual identity for txn2 lives in [`txn2/www/DESIGN.md`](https://github.com/txn2/www/blob/master/DESIGN.md) with tokens in [`txn2/www/tokens.json`](https://github.com/txn2/www/blob/master/tokens.json). This file defers to those for everything below. If a value here disagrees with upstream, upstream wins.

mcp-datahub mirrors the implementation pattern established by sister projects [`txn2/mcp-s3`](https://github.com/txn2/mcp-s3/blob/main/DESIGN.md) and [`txn2/mcp-trino`](https://github.com/txn2/mcp-trino/blob/main/DESIGN.md). When in doubt, copy mcp-s3 verbatim and edit only the project-specific content (symbol, hero copy, tool list, llms.txt, OG card text).

| Concern              | Source of truth |
|----------------------|-----------------|
| Color palette        | upstream `tokens.json` `color.*` |
| Typography stack     | upstream `tokens.json` `font.*` |
| Type scale           | upstream `DESIGN.md` Typography table |
| Spacing / measure    | upstream `tokens.json` `size.*` |
| Component contracts  | upstream `DESIGN.md` Components |
| Voice / copy rules   | upstream `DESIGN.md` Voice and Copy |
| Accessibility rules  | upstream `DESIGN.md` Do's and Don'ts |
| Mermaid theme        | upstream `DESIGN.md` `mcp__card--feature` block |
| MkDocs Material file map | sister `mcp-s3/DESIGN.md` File map |

Tokens are mirrored as CSS custom properties in `docs/stylesheets/extra.css` `:root`. They are duplicated for runtime use, not as a divergence point. When upstream changes a token, update the value in `extra.css` and ship.

## Adoption level: token alignment

Per the upstream downstream contract, three levels are valid:

1. Reference. Link to upstream, no visual changes.
2. Token alignment. Keep MkDocs Material, re-skin via `extra.css` against upstream tokens.
3. Full re-skin. Replace MkDocs Material with custom layouts.

mcp-datahub runs at **level 2**, matching its sister projects mcp-s3, mcp-trino, kubefwd, and txeh. The site keeps Material's instant nav, search, sidebar, code copy, and content extensions. The visual layer is replaced. The homepage is a custom Material template that takes over `block header`, `block container`, and `block footer` for full-bleed treatment.

## File map

| Path | Role |
|------|------|
| `mkdocs.yml`                       | Single dark `slate` palette. `font: false` so CSS loads the upstream Google Fonts URL with trimmed axes. |
| `docs/index.md`                    | Stub front matter with `template: home.html`. All homepage HTML lives in the template. |
| `docs/overrides/main.html`         | Adds the upstream Google Fonts `<link>`, full SEO surface (OG, Twitter, JSON-LD `SoftwareApplication`), and the canonical/author meta tags per the upstream "SEO and social cards" spec. Inherited by every page. |
| `docs/images/mcp-datahub-og.svg`   | Source for the 1200x630 social card. Edit this, then re-rasterise. |
| `docs/images/mcp-datahub-og.png`   | Rendered OG card. Linked from `og:image` and `twitter:image`. Re-render with `rsvg-convert -w 1200 -h 630 -o docs/images/mcp-datahub-og.png docs/images/mcp-datahub-og.svg`. |
| `docs/images/mcp-datahub-symbol.svg` | Square geometric mark used by the hero. Two paths only: paper-toned outer and inner cap arcs, signal-orange middle horseshoe accent. |
| `docs/llms.txt`                    | LLM-friendly docs map per the upstream `llms.txt` spec. Update when new top-level docs pages ship. |
| `docs/overrides/home.html`         | Custom homepage template. Overrides `block header` (rail), `block tabs` (empty), `block container` (page--home shell with hero, sections, flagship cards, stack, coda), `block footer` (home-footer). |
| `docs/overrides/404.html`          | Restyled not-found page. Inherits `main.html`, uses `.md-typeset` body so the rail and footer match. |
| `docs/stylesheets/extra.css`       | All design rules. Two halves: homepage components scoped under `.page--home`, and Material chrome restyle for inner pages via `[data-md-color-scheme="slate"]` variable overrides. |

## Project-specific components

Components ported from upstream (via mcp-s3) verbatim, with mcp-datahub content:

- `.rail` (replaces Material `.md-header` on the homepage). Brand links to `./`. Live UTC clock in meta. txn2.com link in meta as `part of <em class="serif">txn2</em> ↗`.
- `.hero__main` (project-site hero variant): square symbol on the left, three-row Fraunces display on the right. mcp-datahub / metadata catalog / for ai. See upstream `.hero__main` and `.hero__mark` component spec.
- `.hero__mark` linking to `https://github.com/txn2/mcp-datahub`. Symbol file at `docs/images/mcp-datahub-symbol.svg` (square, viewBox `10 10 80 80`, two paths: the canonical DataHub mark extracted from the historical `MCP-datahub-logo-banner.svg` lockup, recoloured to the txn2 palette: paper-toned outer arc plus inner cap, signal-orange middle horseshoe). Accent breathes per upstream spec.
- `.section`, `.section__index`, `.section__title`.
- `.flagship__card`. Two cards: a server card (standalone MCP install + connect demo) and a library card (Go composition with `client.New`, `tools.NewToolkit`). Top accent line animates on hover per upstream spec.
- `.terminal`, `.terminal__bar`, `.terminal__body` with `.t-prompt`, `.t-ok`, `.t-mute` classes. The only block with shadow.
- `.stack`, `.stack__row`. Each row links to a real anchor in the docs (server/, library/, reference/, ecosystem/).
- `.coda` and `.home-footer` (renamed from upstream `.footer` to avoid collision with markdown that uses `class="footer"`; see kubefwd Learning #5).

Components from upstream **not used** here, with reason:

- `.mcp__card--feature` and the MCP grid. mcp-datahub is a single MCP server, not an MCP catalog. The catalog lives at `mcp-data-platform.txn2.com`.
- The 5-column footer's `sponsors / craig` columns. The mcp-datahub home-footer has `about / docs / interfaces / code / txn2 / org` columns instead, since this site is project-scoped.

Custom additions specific to mcp-datahub:

- The hero subtitle reads "metadata catalog" rather than "object storage" (mcp-s3) or "federated sql" (mcp-trino) to position mcp-datahub as the catalog/governance bridge in the txn2 mcp data platform suite.
- The server flagship card's terminal demo uses a `go install` + `claude mcp add datahub` invocation against a DataHub GraphQL endpoint with token auth. The library card shows the canonical `multiserver.FromEnv` / `client.New` / `tools.NewToolkit` / `RegisterAll` path, including the `Config{ WriteEnabled: false }` literal that defaults the toolkit to the nine read tools.
- The stack list emphasises the discriminator pattern (twelve tools covering thirty-five operations) and column-level lineage, which are the two features that distinguish mcp-datahub from a plain GraphQL wrapper.
- An `ecosystem` callout in the stack list points at the sister MCP projects (`mcp-data-platform`, `mcp-s3`, `mcp-trino`) so readers see the broader composable suite. Links go to each project's documentation site, not its GitHub repo.

## Symbol design

`docs/images/mcp-datahub-symbol.svg` is the canonical DataHub mark, extracted from the historical `MCP-datahub-logo-banner.svg` lockup and recoloured for the txn2 palette. Square mark in viewBox `10 10 80 80` with exactly two paths:

- **Base** (filled `var(--paper)`): the outer 90-degree arc spanning lower-left to lower-right plus the inner cap (the originally-white pieces in the source lockup), combined into one path via `Z M`.
- **Accent** (filled `var(--signal)`): the middle horseshoe (the originally-blue piece in the source lockup), the bolder of the three concentric arcs. Breathes per the upstream `mark-breath` keyframes.

The original path geometry from the banner is preserved verbatim, including all cubic Bezier segments, so the mark stays visually identical to DataHub's own brand. Only the fills change: black + blue becomes paper + signal-orange. The accent breathing pulse animates only the middle arc, leaving the outer arc and inner cap static.

## MkDocs Material learnings

The Material learnings list applies to every MkDocs Material project re-skinned to the txn2 identity. mcp-datahub inherits them all from kubefwd, mcp-s3, and mcp-trino. Read the full set in [`txn2/kubefwd/DESIGN.md`](https://github.com/txn2/kubefwd/blob/master/DESIGN.md) "MkDocs Material learnings". Brief summary so this file remains useful in isolation:

1. Override the homepage via a separate template, not via CSS hacks.
2. Re-skin inner pages via Material variable overrides on `[data-md-color-scheme="slate"]`.
3. `font: false` to load fonts directly from CSS.
4. Scope every homepage component class under `.page--home`.
5. Rename `.footer` to `.home-footer` to avoid collision.
6. `h3` and `h4` are technical reference, not display type. Switch them to Instrument Sans bold; flip any heading containing inline code to JetBrains Mono via `:has(code)` with a `@supports not selector(:has(*))` fallback.
7. Tabbed content nests boxes by default. Strip `.tabbed-set` background and border, keep only the label underline.
8. Mermaid via Material's `--md-mermaid-*` CSS variables, not via separate init.
9. Guard inline scripts against `navigation.instant` rehydration. The live UTC clock uses a `window.__mcpDatahubClock` sentinel.
10. Drop the light/dark toggle. Single `scheme: slate`.
11. Atmospheric overlays at low z-index (grain and vignette at z-index 1, below rail at z-50).
12. Hugo-only token compilation does not exist in MkDocs. Token sync is a manual edit to `extra.css`.

## Voice and copy

Defers to upstream. Briefly:

- No em-dashes (U+2014) or en-dashes (U+2013) anywhere, including code comments and template comments. Use commas, periods, colons, parentheses, slashes, hyphens.
- No AI-tell vocabulary: `seamless`, `leverage`, `comprehensive`, `robust`, `delve`, `unleash`, `elevate`, `embark`, `tapestry`, `not just X but Y`, `as an AI`, `let me X`.
- Sentence case for body. Lowercase for rail and label text. Title case rare.
- Section indices: `§ 01 / title` with slash, never an em-dash.
- Year ranges use a hyphen: `2025-2026`.
- Verify before commit: `grep -RE "—|–" docs/ mkdocs.yml`.

## Updating

When the upstream `txn2/www/DESIGN.md` or `tokens.json` changes:

1. Read the upstream diff. Identify which tokens, components, or rules changed.
2. Update the matching CSS variables in `docs/stylesheets/extra.css` `:root`.
3. If a component contract changed (padding, border, hover behavior), update the homepage template in `docs/overrides/home.html` and the matching CSS rules.
4. Update this file's File map / Project-specific components sections if a new component is added or removed.
5. Run `python3 -m mkdocs build --strict` and verify in the browser before committing. Verify the home page hero, flagship cards, terminal, stack, coda, and home-footer. Verify an inner page (`/server/`, `/library/`, `/reference/tools-api/`) still inherits the look.

Keep this file thin. If a section grows past 30 lines, ask whether it belongs upstream instead.
