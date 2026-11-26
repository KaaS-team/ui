# The gate. `just check` is what CI runs; it must end with zero errors and zero warnings.

default: check

check: fmt vet test tokens prose axe

# gofmt lists unformatted files; any output fails.
fmt:
    @out="$(gofmt -l .)"; if [ -n "$out" ]; then echo "gofmt: $out"; exit 1; fi

vet:
    go vet ./...

# Partials render valid HTML, files match ui.Names, statuses have tokens, /_ui/ serves the CSS.
test:
    go test ./...

# tokens.css and tokens.json name the same tokens; no raw colour outside tokens.css.
tokens:
    node --test test/*.test.mjs

# House prose rules on every tracked or new Markdown file outside the vendored kit.
prose:
    git ls-files --cached --others --exclude-standard '*.md' | grep -v '^scripts/kaas/' | xargs python3 scripts/kaas/check-prose.py

# axe-core on /_ui/catalogue (every partial in every state), light and dark, 1280 and 390 wide. Zero violations.
# Installs the pinned dev dependencies on first run; on CI it also fetches the Chromium build Playwright pins.
axe:
    @[ -d node_modules/axe-core ] || npm ci --no-audit --no-fund
    @if [ -n "${CI:-}" ]; then npx playwright install chromium; fi
    node catalogue/axe.mjs

# Full-page screenshots of the catalogue in the same four views, into reports/cloud/U2 or the given folder.
screens dir="reports/cloud/U2":
    @[ -d node_modules/playwright ] || npm ci --no-audit --no-fund
    node catalogue/screens.mjs {{dir}}
