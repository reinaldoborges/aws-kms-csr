# aws-kms-csr

Generates and signs a Certificate Signing Request (CSR) using a private key held in AWS KMS; the key never leaves KMS, so signing happens via KMS's sign API rather than local key material.

## Agent skills

### Issue tracker

Issues and specs live as Markdown files under `.scratch/`. See `docs/agents/issue-tracker.md`.

### Domain docs

This is a single-context repo with root `GLOSSARY.md` and `docs/adr/`. See `docs/agents/domain.md`.
