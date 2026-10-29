Type: prototype
Status: resolved

## Question

What should the TUI's screen flow and navigation be: a single scrolling form with all fields (Key Identifier, Common Name, any credential-fallback fields, output path) visible at once, or a step-by-step wizard (one field/decision per screen, with back-navigation)? How is per-field validation surfaced (inline as you type vs. on submit), and what does the happy-path screen sequence look like end to end?

## Answer

**Variant A: single scrolling form.** All fields (Key Identifier, Common Name, Output path) visible on one screen; `Tab`/`Shift+Tab` moves focus between them; validation is inline and live, rendered under each invalid field as it's edited, not deferred to submit. The happy path is one screen: fill both required fields (Key Identifier, Common Name), optionally adjust the output path (defaulted to `./request.csr`), then submit.

Reached by building three structurally different runnable variants (single form, full step-by-step wizard, grouped wizard) with Bubble Tea and having the user try all three interactively. Full prototype code (all three variants, switchable at runtime): branch `prototype/tui-screen-flow` (commit `4761e71`), never merged.
