package csr

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms/types"
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
