# aws-kms-csr

`aws-kms-csr` is a terminal UI that generates a Certificate Signing Request
(CSR) using a private key held by AWS KMS. The private key never leaves KMS:
the application retrieves the public key with `GetPublicKey` and asks KMS to
sign the CSR digest with `Sign`.

The application supports RSA KMS Signing Keys (`RSA_2048`, `RSA_3072`, and
`RSA_4096`) and elliptic-curve keys (`ECC_NIST_P256`, `ECC_NIST_P384`, and
`ECC_NIST_P521`). The Signing Algorithm is selected automatically from the
Key Spec.

## Requirements

- Go 1.25 or newer
- An asymmetric KMS Signing Key in the AWS Region you will use
- AWS credentials with permission to call:
  - `kms:GetPublicKey`
  - `kms:Sign`

The application uses the AWS SDK's default credential chain. If credentials or
the Region are not available, it prompts for the missing values in the TUI.
The default Region suggestion is `sa-east-1`.

## Test

Run the offline unit tests:

```bash
go test ./...
```

Run static checks:

```bash
go vet ./...
```

The default test suite uses deterministic fakes and does not make live AWS
requests.

## Build

Build the executable in the current directory:

```bash
go build -o aws-kms-csr ./cmd/aws-kms-csr
```

Or install it into Go's binary directory:

```bash
go install ./cmd/aws-kms-csr
```

## Run

Run directly from the repository:

```bash
go run ./cmd/aws-kms-csr
```

Or run the binary produced by `go build`:

```bash
./aws-kms-csr
```

The application opens a single form. Fill in:

1. **Key Identifier** — an alias, key ID, or key ARN, such as
   `alias/my-signing-key`.
2. **Subject Common Name** — the Common Name to place in the CSR, such as
   `example.com`.
3. **Output Path** — defaults to `./request.csr`.

Use `Tab` and `Shift+Tab` to move between fields. Press `Enter` or `Ctrl+S`
to submit or write the generated CSR. If the output file already exists, the
application asks for confirmation before overwriting it. Press `r` to retry a
failed operation and `Esc` to quit.

After generation, the PEM-encoded CSR is displayed in the TUI and written to
the selected output path after confirmation when needed.

## AWS credential behavior

The application first checks the AWS SDK default credential and Region
configuration, including environment variables, shared configuration, SSO,
and instance or container roles. It prompts only for values that remain
missing:

- Access Key ID
- Secret Access Key
- Session Token
- Region

When the SDK resolves credentials automatically, the detected source is shown
in the status area.

## Error logging

Stage failures are shown in a sanitized status message. Detailed diagnostic
information is appended to:

```text
./aws-kms-csr-debug.log
```

The log is enabled automatically and includes timestamps and failure stages.
Credential values and signing material are redacted. If the debug log cannot
be written, the primary CSR flow continues normally.

## Security model

The application does not download, export, or store the KMS private key. CSR
signing happens through the KMS `Sign` API, and only the resulting CSR and
diagnostic log are written locally.
