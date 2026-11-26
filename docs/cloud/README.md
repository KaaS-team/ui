# Cloud tickets U1 and U2

Two sessions build this repo in parallel. Touch only what your brief owns.

- **U1, tokens and CSS layers**: `css/` and `tokens.json`. Tune token values and fill the layer files. Token names are fixed; a rename changes both files and every consumer.
- **U2, Go side**: `*.go`, `templates/`, the catalogue page and the Lip Gloss theme. Fill each partial against its struct in `data.go`; the names in `ui.Names` are fixed.

Parity tests that keep the halves in step:

- `test/tokens.test.mjs`: tokens.css and tokens.json name the same tokens. U2 extends this to Go (the Lip Gloss theme reads tokens.json or tokens.css, never its own copy).
- `test/colours.test.mjs`: no raw colour outside `css/tokens.css`.
- `partials_test.go`: every partial renders valid HTML from its sample; statuses have tokens.

Run `just check` before every push. It must pass with zero errors.
