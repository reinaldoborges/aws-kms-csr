Type: grilling
Status: resolved

## Question

What's the testing strategy for the KMS Signer and CSR-building flow without making live AWS calls: do we define a narrow interface over the KMS SDK client (covering `GetPublicKey` and `Sign`) so tests can supply a fake, use the AWS SDK's own testing/mocking facilities, or something else? What specifically needs unit coverage (key-spec-to-algorithm mapping, unsupported-spec rejection, CSR PEM output) versus what's left to manual/integration testing against real KMS?

## Answer

Define a narrow local KMS API interface exposing only `GetPublicKey` and `Sign`. Production adapts the AWS SDK client; unit tests use a deterministic fake that records requests, supports injected failures, and asserts call count/order. The default automated suite is offline and deterministic; real-KMS integration and interactive rendering checks are opt-in/manual.

Automated coverage includes:

- Table-driven Key Spec to Signing Algorithm mapping for `RSA_2048`, `RSA_3072`, `RSA_4096`, `ECC_NIST_P256`, `ECC_NIST_P384`, and `ECC_NIST_P521`.
- Explicit rejection of unsupported HMAC, symmetric, ML-DSA, and other non-supported families.
- Exact KMS request contracts: Key Identifier, selected algorithm, `MessageType=DIGEST`, and the digest passed to the signer.
- Public-key parsing, cached/lazy public-key behavior, error propagation, and prevention of signing without usable key metadata.
- CSR PEM parsing and cryptographic verification of the CSR signature using the returned public key, including Subject Common Name and public-key type. Tests avoid brittle fixed DER snapshots.
- Pure TUI state transitions for required-field validation, default output path, overwrite confirmation, retry/focus behavior, and preservation of the generated CSR after output-path failure.
- Stable application-level error categories/stages and relevant context, without matching volatile AWS request IDs or full SDK-formatted strings.

Tests may generate ephemeral RSA and EC key pairs as fixtures. The fake KMS API can return the fixture public key and use the test-only private key to create valid signatures; production code never receives private key material. An opt-in real-KMS acceptance check uses a pre-provisioned asymmetric KMS Signing Key, externally supplied credentials/region, and required permissions, with no cloud-resource provisioning or cleanup.
