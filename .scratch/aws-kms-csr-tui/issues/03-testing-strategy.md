Type: grilling
Status: open

## Question

What's the testing strategy for the KMS Signer and CSR-building flow without making live AWS calls: do we define a narrow interface over the KMS SDK client (covering `GetPublicKey` and `Sign`) so tests can supply a fake, use the AWS SDK's own testing/mocking facilities, or something else? What specifically needs unit coverage (key-spec-to-algorithm mapping, unsupported-spec rejection, CSR PEM output) versus what's left to manual/integration testing against real KMS?
