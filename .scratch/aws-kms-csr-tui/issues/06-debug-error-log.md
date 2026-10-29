Type: grilling
Status: resolved

## Question

Should the TUI register detailed error logs for debug, separate from the already-decided sanitized on-screen error messages (Error handling and recovery ticket)? If so: what triggers logging, where does the log live, does it append or reset, what format and content does each entry have, and what happens if the log write itself fails?

## Answer

Yes. A separate, file-only debug log captures full unsanitized error detail — additive to, never a replacement for, the sanitized on-screen status message already decided in Error handling and recovery.

**Scope — what gets logged**: only actual stage failures (KMS errors, network/timeout, output-write errors), not routine field-validation messages (e.g. "Key Identifier is required."), which are expected user-input states, not debuggable errors.

**Content**: the full raw error detail per entry — stage, AWS SDK error type/code, request ID, and stack/error-chain detail via Go's `%+v`-style formatting. Still never logs credential values (Access Key ID, Secret Access Key, Session Token) or signing material, even though the file is debug-only; the exclusions from the on-screen decision carry over unconditionally.

**Location and format**: a fixed relative path, `./aws-kms-csr-debug.log` in the current working directory — matching this app's existing simple-relative-path precedent (`./request.csr`), not an OS-specific state directory. Plain text, one entry per logged error, each prefixed with a timestamp and stage; no structured/JSON format, since there's no log-aggregation pipeline for a single-developer debug aid.

**Lifecycle**: always-on whenever a stage failure occurs — no flag, environment variable, or TUI setting needed to enable it. Appends across runs (never truncated), so intermittent/flaky failures build a history; unbounded growth over time is an accepted tradeoff for a single-purpose CLI tool.

**Visibility**: the TUI tells the user the log's path once — shown in the status area alongside the first logged error — reusing the precedent set by Credential source display for small informational lines in that same area.

**Failure handling**: if writing the debug log itself fails (permission denied, disk full, etc.), it silently degrades. The on-screen sanitized error still appears normally; a logging failure never blocks, crashes, or otherwise disrupts the primary CSR-generation flow.
