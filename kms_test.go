package csr

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"strings"
	"testing"

	"crypto/x509"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/smithy-go"
)

func TestSigningAlgorithmSupportsRSAAndECKeySpecs(t *testing.T) {
	tests := []struct {
		spec      types.KeySpec
		algorithm types.SigningAlgorithmSpec
	}{
		{types.KeySpecRsa2048, types.SigningAlgorithmSpecRsassaPkcs1V15Sha256},
		{types.KeySpecRsa3072, types.SigningAlgorithmSpecRsassaPkcs1V15Sha256},
		{types.KeySpecRsa4096, types.SigningAlgorithmSpecRsassaPkcs1V15Sha256},
		{types.KeySpecEccNistP256, types.SigningAlgorithmSpecEcdsaSha256},
		{types.KeySpecEccNistP384, types.SigningAlgorithmSpecEcdsaSha384},
		{types.KeySpecEccNistP521, types.SigningAlgorithmSpecEcdsaSha512},
	}
	for _, test := range tests {
		algorithm, _, err := signingAlgorithm(test.spec)
		if err != nil {
			t.Fatalf("spec %s: %v", test.spec, err)
		}
		if algorithm != test.algorithm {
			t.Errorf("spec %s: algorithm = %s, want %s", test.spec, algorithm, test.algorithm)
		}
	}
}

func TestSigningAlgorithmRejectsUnsupportedKeySpecs(t *testing.T) {
	if _, _, err := signingAlgorithm(types.KeySpecSymmetricDefault); err == nil {
		t.Fatal("expected symmetric Key Spec to be rejected")
	}
}

func TestKMSAccountErrorExplainsRequiredPermissionsForAccessDenied(t *testing.T) {
	err := &KMSAccountError{
		Stage: "signing CSR with KMS",
		Err: &smithy.GenericAPIError{
			Code:    "AccessDeniedException",
			Message: "not authorized",
		},
	}

	got := err.Error()
	want := "signing CSR with KMS: authorization failed; KMS policy must allow kms:GetPublicKey and kms:Sign"
	if got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
	if !errors.Is(err, err.Err) {
		t.Fatal("KMS account error should unwrap its underlying error")
	}
}

func TestKMSAccountErrorSanitizesServiceMessage(t *testing.T) {
	err := &KMSAccountError{
		Stage: "fetching KMS public key",
		Err: &smithy.GenericAPIError{
			Code:    "ServiceError",
			Message: "line one\nline two " + string(make([]byte, 300)),
		},
	}

	got := err.Error()
	if len(got) > 320 {
		t.Fatalf("error length = %d, want bounded service message", len(got))
	}
	if strings.Contains(got, "\n") {
		t.Fatalf("error contains a newline: %q", got)
	}
}

type emptyKMSResponse struct{}

func (emptyKMSResponse) GetPublicKey(context.Context, *kms.GetPublicKeyInput, ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error) {
	return nil, nil
}

func (emptyKMSResponse) Sign(context.Context, *kms.SignInput, ...func(*kms.Options)) (*kms.SignOutput, error) {
	return nil, errors.New("sign should not be called")
}

func TestNewKMSSignerRejectsEmptyPublicKeyResponse(t *testing.T) {
	_, err := NewKMSSigner(context.Background(), emptyKMSResponse{}, "alias/example")
	if err == nil {
		t.Fatal("expected empty KMS response to be rejected")
	}
	if err.Error() != "fetching KMS public key: KMS request failed" {
		t.Fatalf("error = %q, want empty-response error", err)
	}
}

func TestKMSAccountErrorClassifiesActionableKeyLookupFailures(t *testing.T) {
	err := &KMSAccountError{
		Stage: "fetching KMS public key",
		Err:   &smithy.GenericAPIError{Code: "NotFoundException"},
	}
	if !err.FocusKeyIdentifier() {
		t.Fatal("expected not-found key lookup failure to focus Key Identifier")
	}

	authErr := &KMSAccountError{
		Stage: "fetching KMS public key",
		Err:   &smithy.GenericAPIError{Code: "AccessDeniedException"},
	}
	if authErr.FocusKeyIdentifier() {
		t.Fatal("expected authorization failure not to focus Key Identifier")
	}
}

type emptySignResponse struct{}

func (emptySignResponse) GetPublicKey(context.Context, *kms.GetPublicKeyInput, ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	publicKey, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	return &kms.GetPublicKeyOutput{KeySpec: types.KeySpecRsa2048, PublicKey: publicKey}, nil
}

func (emptySignResponse) Sign(context.Context, *kms.SignInput, ...func(*kms.Options)) (*kms.SignOutput, error) {
	return nil, nil
}

func TestKMSSignerRejectsEmptySignResponse(t *testing.T) {
	signer, err := NewKMSSigner(context.Background(), emptySignResponse{}, "alias/example")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := signer.Sign(rand.Reader, []byte("digest"), nil); err == nil || err.Error() != "signing CSR with KMS: KMS request failed" {
		t.Fatalf("sign error = %v, want staged empty-response error", err)
	}
}
