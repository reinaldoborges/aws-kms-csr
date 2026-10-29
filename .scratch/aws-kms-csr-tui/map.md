## Destination

A spec for a Go TUI application (`aws-kms-csr`, module `github.com/reinaldo/aws-kms-csr`) that:

- Resolves AWS credentials via the SDK's default chain, falling back to TUI prompts (Access Key ID, Secret Access Key, Session Token, Region — region suggested as `sa-east-1`) only for whatever's missing.
- Prompts for the KMS Key Identifier (alias, key ID, or ARN) and the Subject Common Name.
- Fetches the KMS Signing Key's public key and Key Spec, builds an unsigned CSR, and signs it via a `KMS Signer` (`crypto.Signer` wrapping KMS's `GetPublicKey`/`Sign`), auto-selecting the Signing Algorithm (RSA or EC families; anything else rejected with a clear error).
- Writes the PEM-encoded CSR to a user-chosen output path (default `./request.csr`, confirming before overwrite) and echoes it to screen.

Reaching the end of this map means: no open decision remains before a build session can implement the whole flow in one pass.

## Notes

- Plan-only: tickets resolve decisions, not code. No override.
- Read `GLOSSARY.md` and `docs/adr/0001-kms-backed-crypto-signer.md` before working any ticket; they fix the vocabulary and the signer architecture.
- Every session should consult `/domain-modeling` when a new term surfaces, and `/grilling` for any HITL ticket.
- TUI library is settled: Bubble Tea + `bubbles` + `lipgloss`.

## Decisions so far

- Destination, scope, and architecture (TUI library, KMS Signer design, key-type support, subject fields, credential flow, key-identifier input, output handling) settled via the charting session's grilling round. See `GLOSSARY.md` and ADR-0001.
- [Dependency versions](./issues/04-dependency-versions.md): pinned to `aws-sdk-go-v2/service/kms` v1.62.1, `aws-sdk-go-v2/config` v1.33.12, `charmbracelet/bubbletea` v1.3.10, `bubbles` v1.0.0, `lipgloss` v1.1.0 (the v1 module paths, not Charm's separately-pathed v2 line); credential-chain fallback detected via `cfg.Region == ""` + a short-timeout `cfg.Credentials.Retrieve(ctx)`. Detail on branch `research/dependency-versions`.

## Not yet specified

- Distribution/packaging (how the binary gets built and released) — not sharp enough to ticket yet.
- Documentation of the IAM permissions the KMS Signing Key's policy must grant (`kms:GetPublicKey`, `kms:Sign`) — will graduate once the build phase confirms the exact call shapes.

## Out of scope

- (none yet — out-of-scope work is recorded here as it's ruled out)
