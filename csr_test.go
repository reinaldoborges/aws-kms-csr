package csr

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

type fakeKMS struct {
	publicKey []byte
	keySpec   types.KeySpec
	sign      func([]byte) []byte
	lastSign  *kms.SignInput
}

func (f *fakeKMS) GetPublicKey(context.Context, *kms.GetPublicKeyInput, ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error) {
	return &kms.GetPublicKeyOutput{PublicKey: f.publicKey, KeySpec: f.keySpec}, nil
}

func (f *fakeKMS) Sign(_ context.Context, input *kms.SignInput, _ ...func(*kms.Options)) (*kms.SignOutput, error) {
	f.lastSign = input
	return &kms.SignOutput{Signature: f.sign(input.Message), SigningAlgorithm: input.SigningAlgorithm}, nil
}

func TestCreateCSRUsesKMSAndProducesVerifiablePEM(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeKMS{
		publicKey: publicDER,
		keySpec:   types.KeySpecRsa2048,
		sign: func(message []byte) []byte {
			signature, signErr := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, message)
			if signErr != nil {
				t.Fatal(signErr)
			}
			return signature
		},
	}

	request, err := CreateCSR(context.Background(), fake, "alias/example", "example.test")
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(request)
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		t.Fatalf("expected CSR PEM, got %q", request)
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if err := csr.CheckSignature(); err != nil {
		t.Fatalf("CSR signature is invalid: %v", err)
	}
	if csr.Subject.CommonName != "example.test" {
		t.Fatalf("common name = %q, want example.test", csr.Subject.CommonName)
	}
	if fake.lastSign == nil || fake.lastSign.KeyId == nil || *fake.lastSign.KeyId != "alias/example" {
		t.Fatalf("unexpected sign request: %#v", fake.lastSign)
	}
	if fake.lastSign.MessageType != types.MessageTypeDigest {
		t.Fatalf("message type = %v, want DIGEST", fake.lastSign.MessageType)
	}
	if fake.lastSign.SigningAlgorithm != types.SigningAlgorithmSpecRsassaPkcs1V15Sha256 {
		t.Fatalf("algorithm = %v", fake.lastSign.SigningAlgorithm)
	}
}
