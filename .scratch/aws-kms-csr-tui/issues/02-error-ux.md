Type: grilling
Status: open

## Question

How should AWS/KMS failures surface in the TUI: a KMS `AccessDeniedException` on `Sign`, an invalid/nonexistent Key Identifier, a rejected Key Spec, a credential-fallback failure, or a network/timeout error? Decide the UX (inline message vs. modal/full-screen error view), whether the app offers retry or re-prompts the failing field versus exiting with a message, and what's shown for failures that aren't anticipated by name.
