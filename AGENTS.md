# ui

The KaaS design system. One repo, two packages:

- Go module `github.com/KaaS-team/ui` (go 1.26): embeds `css/`, serves it at `/_ui/` with `ui.Handler()`, and ships the html/template partials in `ui.Partials` (names in `ui.Names`, data structs in `data.go`). Used by factory, kaasbot and dx.
- npm package `@kaas-team/ui`, CSS only: `css/ui.css` imports `tokens.css` then the layer files. Used by the QA dashboard.

## Commands

| Command | What it runs |
| --- | --- |
| `just check` | The gate: fmt, vet, test, tokens, prose. CI runs it. |
| `just test` | `go test ./...` |
| `just tokens` | `node --test test/*.test.mjs` |
| `just prose` | `scripts/kaas/check-prose.py` on every Markdown file |

## Boundaries

- A colour is written only in `css/tokens.css`. Everything else uses `var(--ui-*)`.
- A token name changes in `css/tokens.css` and `tokens.json` together.
- A new partial is a `templates/<name>.tmpl`, an entry in `ui.Names`, a data struct in `data.go` and a sample in `partials_test.go`.
- `scripts/kaas/` is vendored from the kit. Do not edit it here.
- No runtime npm dependencies. No Go dependency without `/eng-add-dependency`.

## Tests

| Test | Fails when |
| --- | --- |
| `test/tokens.test.mjs` | tokens.css and tokens.json name different tokens, a dark group is not redefined, a `var()` names an unknown token, or ui.css skips a layer |
| `test/colours.test.mjs` | a raw hex, rgb or hsl colour sits outside tokens.css and `colour-allow.txt`, or an allow entry is stale |
| `partials_test.go` | a partial fails to render its sample, renders unbalanced HTML, or its file and `ui.Names` disagree; a status lacks its tokens; `/_ui/` does not serve CSS |
