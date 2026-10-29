# aws-kms-csr

A Go TUI that generates and signs a Certificate Signing Request using a private key held in AWS KMS; the private key material never leaves KMS.

## Language

**KMS Signing Key**:
The asymmetric KMS key that holds the private key used to sign the CSR. Identified by the user as a key alias, key ID, or key ARN — the app never lists or browses keys in the account.
_Avoid_: KMS key (ambiguous with KMS's own terminology for any key, including symmetric ones), private key

**Key Identifier**:
The single string the user supplies to name the KMS Signing Key: a key alias (e.g. `alias/my-key`), a key ID, or a key ARN. KMS accepts all three interchangeably wherever a key reference is required.
_Avoid_: key name, key ref

**Key Spec**:
The cryptographic family and size of the KMS Signing Key, reported by KMS (e.g. `RSA_2048`, `ECC_NIST_P256`). Determines which Signing Algorithm is valid and drives auto-detection; specs outside the supported RSA/EC families are rejected.
_Avoid_: key type (used loosely elsewhere for this), algorithm (that's a property the spec implies, not the spec itself)

**Signing Algorithm**:
The specific KMS signing algorithm paired to a Key Spec (e.g. `RSASSA_PKCS1_V1_5_SHA_256` for RSA specs, `ECDSA_SHA_256`/`384`/`512` for the matching EC curve). Chosen automatically from the Key Spec, never asked of the user.

**KMS Signer**:
The app's `crypto.Signer` implementation backing CSR creation: `Public()` fetches and parses the KMS Signing Key's public key once; `Sign()` sends the digest to KMS's `Sign` API with `MessageType=DIGEST` and returns the signature. The standard library's CSR builder calls this signer without ever holding the private key.
_Avoid_: signer (too generic in a file full of AWS SDK signers), KMS client (that's the lower-level SDK object the KMS Signer wraps)

**Credential Fallback**:
The TUI's behavior when the AWS SDK's default credential chain or region resolution comes up short: prompting interactively for whatever is missing (access key ID, secret access key, session token, region) rather than failing outright.
_Avoid_: credential prompt (names only the UI step, not the whole resolve-then-fallback behavior)

**Subject Common Name**:
The sole CSR subject field the TUI collects from the user; the only required input besides the Key Identifier. Other X.509 subject fields (organization, locality, etc.) are out of scope for this effort, not merely left blank.
_Avoid_: CN, subject (subject is the broader X.509 concept; this is the one field the app surfaces)
