Type: research
Status: resolved

## Question

Confirm the current module scaffolding facts needed before a build session starts: the latest stable versions of `github.com/aws/aws-sdk-go-v2/service/kms`, `github.com/aws/aws-sdk-go-v2/config` (for default credential/region resolution), `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/bubbles`, and `github.com/charmbracelet/lipgloss`; the minimum Go version each requires; and whether `aws-sdk-go-v2/config` exposes a way to probe "would the default chain resolve credentials/region" without making a live call (needed for the Credential Fallback decision already made).

## Answer

Confirmed versions (2026-10-28): `aws-sdk-go-v2/service/kms` v1.62.1, `aws-sdk-go-v2/config` v1.33.12 (both go 1.24); `charmbracelet/bubbletea` v1.3.10 (go 1.24.0), `charmbracelet/bubbles` v1.0.0 (go 1.24.2), `charmbracelet/lipgloss` v1.1.0 (go 1.18). Note: Charm's v2 line lives at a *different* module path (`charm.land/.../v2`, requiring go 1.26.0/1.24.2), not an upgrade of `github.com/charmbracelet/...`; stick with the v1 `github.com/charmbracelet` line as scaffolded.

No dedicated dry-run validation API exists for the credential chain. Idiomatic probe: check `cfg.Region == ""` locally (free), and call `cfg.Credentials.Retrieve(ctx)` once with a short timeout right after `LoadDefaultConfig` — cached by `aws.CredentialsCache` so it costs nothing extra later. Caveat: if the chain falls through to EC2 IMDS, `Retrieve` attempts a local-network call to `169.254.169.254` unless IMDS is explicitly disabled.

Full citations and code-level detail: branch `research/dependency-versions`, `.scratch/aws-kms-csr-tui/research/dependency-versions.md` (commit `4d4391e`).
