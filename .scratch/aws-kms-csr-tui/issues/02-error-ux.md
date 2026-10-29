Type: grilling
Status: resolved

## Question

How should AWS/KMS failures surface in the TUI: a KMS `AccessDeniedException` on `Sign`, an invalid/nonexistent Key Identifier, a rejected Key Spec, a credential-fallback failure, or a network/timeout error? Decide the UX (inline message vs. modal/full-screen error view), whether the app offers retry or re-prompts the failing field versus exiting with a message, and what's shown for failures that aren't anticipated by name.

## Answer

Use the single-form layout with entered values preserved. After submission, show a persistent full-width status/error section below the fields; keep normal Tab navigation, make `Retry` the primary action, and provide an explicit `Quit` action. Focus the first actionable field for invalid input or incomplete credential fallback; otherwise focus `Retry`.

Failure handling:

- Invalid or nonexistent Key Identifier: return to the form, focus the Key Identifier, and allow correction.
- Credential-fallback failure: return to the first missing credential/region field and allow correction.
- Unsupported Key Spec: stay on the form and report the received spec, explaining that only supported RSA and EC families are accepted; keep `Retry` available.
- KMS `AccessDeniedException` and network/timeout failures: show a non-field error with a retry action. For authorization failures, explain that the KMS policy must allow `kms:GetPublicKey` and `kms:Sign`.
- Output-path or write failure: preserve the generated CSR, focus Output Path, and allow correction. Declining overwrite is a normal editable branch, not an error.
- Unexpected failures: show the failing stage plus a concise sanitized service code/message where available. Never render credential values, signing material, or unbounded raw SDK dumps.

Retry resumes from the failed stage while retaining safe resolved state; it reacquires or invalidates KMS public-key state when key lookup failed or the Key Identifier changed. The application exits after a successful write or an explicit quit; it does not loop into a new CSR after a terminal outcome.
