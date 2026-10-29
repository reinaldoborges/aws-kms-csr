package csr

import (
	"context"
	"errors"
	"testing"

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
	if err.Error() != "KMS returned an empty public key response" {
		t.Fatalf("error = %q, want empty-response error", err)
	}
}
