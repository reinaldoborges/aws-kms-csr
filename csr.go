package csr

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
)

func CreateCSR(ctx context.Context, client KMSAPI, keyID, subjectCommonName string) ([]byte, error) {
	if subjectCommonName == "" {
		return nil, fmt.Errorf("subject common name is required")
	}
	signer, err := NewKMSSigner(ctx, client, keyID)
	if err != nil {
		return nil, err
	}
	template := &x509.CertificateRequest{
		Subject:            pkix.Name{CommonName: subjectCommonName},
		SignatureAlgorithm: csrSignatureAlgorithm(signer.Public(), signer.Hash()),
	}
	der, err := x509.CreateCertificateRequest(rand.Reader, template, signer)
	if err != nil {
		return nil, fmt.Errorf("creating CSR: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: der}), nil
}

func csrSignatureAlgorithm(publicKey crypto.PublicKey, hash crypto.Hash) x509.SignatureAlgorithm {
	switch publicKey.(type) {
	case *rsa.PublicKey:
		switch hash {
		case crypto.SHA256:
			return x509.SHA256WithRSA
		}
	case *ecdsa.PublicKey:
		switch hash {
		case crypto.SHA256:
			return x509.ECDSAWithSHA256
		case crypto.SHA384:
			return x509.ECDSAWithSHA384
		case crypto.SHA512:
			return x509.ECDSAWithSHA512
		}
	}
	return 0
}
