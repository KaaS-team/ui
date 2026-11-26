# The gate. `just check` is what CI runs; it must end with zero errors and zero warnings.

default: check

check: fmt vet test tokens prose

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
