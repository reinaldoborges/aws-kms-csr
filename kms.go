package csr

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/smithy-go"
)

type KMSAPI interface {
	GetPublicKey(context.Context, *kms.GetPublicKeyInput, ...func(*kms.Options)) (*kms.GetPublicKeyOutput, error)
	Sign(context.Context, *kms.SignInput, ...func(*kms.Options)) (*kms.SignOutput, error)
}

type KMSAccountError struct {
	Stage string
	Err   error
}

func (e *KMSAccountError) Error() string {
	var apiErr smithy.APIError
	if errors.As(e.Err, &apiErr) {
		if apiErr.ErrorCode() == "AccessDeniedException" || apiErr.ErrorCode() == "UnauthorizedException" {
			return fmt.Sprintf("%s: authorization failed; KMS policy must allow kms:GetPublicKey and kms:Sign", e.Stage)
		}
		if message := apiErr.ErrorMessage(); message != "" {
			return fmt.Sprintf("%s: %s (%s)", e.Stage, sanitizeServiceMessage(message), apiErr.ErrorCode())
		}
	}
	return fmt.Sprintf("%s: KMS request failed", e.Stage)
}
func (e *KMSAccountError) Unwrap() error      { return e.Err }
func (e *KMSAccountError) DebugStage() string { return e.Stage }
func (e *KMSAccountError) DebugErr() error    { return e.Err }

func (e *KMSAccountError) FocusKeyIdentifier() bool {
	if e.Stage != "fetching KMS public key" {
		return false
	}
	var apiErr smithy.APIError
	if !errors.As(e.Err, &apiErr) {
		return false
	}
	switch apiErr.ErrorCode() {
	case "NotFoundException", "ResourceNotFoundException", "ValidationException":
		return true
	default:
		return false
	}
}

func sanitizeServiceMessage(message string) string {
	message = strings.Join(strings.Fields(message), " ")
	const maxLength = 256
	if len(message) > maxLength {
		return message[:maxLength-3] + "..."
	}
	return message
}

type KMSSigner struct {
	ctx       context.Context
	client    KMSAPI
	keyID     string
	algorithm types.SigningAlgorithmSpec
	hash      crypto.Hash
	publicKey crypto.PublicKey
	mu        sync.Mutex
}

func NewKMSSigner(ctx context.Context, client KMSAPI, keyID string) (*KMSSigner, error) {
	if keyID == "" {
		return nil, errors.New("key identifier is required")
	}
	output, err := client.GetPublicKey(ctx, &kms.GetPublicKeyInput{KeyId: &keyID})
	if err != nil {
		return nil, &KMSAccountError{Stage: "fetching KMS public key", Err: err}
	}
	if output == nil {
		return nil, &KMSAccountError{
			Stage: "fetching KMS public key",
			Err:   errors.New("KMS returned an empty public key response"),
		}
	}
	algorithm, hash, err := signingAlgorithm(output.KeySpec)
	if err != nil {
		return nil, err
	}
	if len(output.PublicKey) == 0 {
		return nil, errors.New("KMS returned an empty public key")
	}
	publicKey, err := x509.ParsePKIXPublicKey(output.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("parsing KMS public key: %w", err)
	}
	switch key := publicKey.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey:
	default:
		return nil, fmt.Errorf("KMS public key has unsupported type %T", key)
	}
	return &KMSSigner{ctx: ctx, client: client, keyID: keyID, algorithm: algorithm, hash: hash, publicKey: publicKey}, nil
}

func (s *KMSSigner) Public() crypto.PublicKey { return s.publicKey }

func (s *KMSSigner) Sign(_ io.Reader, digest []byte, _ crypto.SignerOpts) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	output, err := s.client.Sign(s.ctx, &kms.SignInput{
		KeyId:            &s.keyID,
		Message:          digest,
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: s.algorithm,
	})
	if err != nil {
		return nil, &KMSAccountError{Stage: "signing CSR with KMS", Err: err}
	}
	if output == nil {
		return nil, &KMSAccountError{
			Stage: "signing CSR with KMS",
			Err:   errors.New("KMS returned an empty signature response"),
		}
	}
	if len(output.Signature) == 0 {
		return nil, &KMSAccountError{
			Stage: "signing CSR with KMS",
			Err:   errors.New("KMS returned an empty signature"),
		}
	}
	return output.Signature, nil
}

func (s *KMSSigner) Hash() crypto.Hash { return s.hash }

func signingAlgorithm(spec types.KeySpec) (types.SigningAlgorithmSpec, crypto.Hash, error) {
	switch spec {
	case types.KeySpecRsa2048, types.KeySpecRsa3072, types.KeySpecRsa4096:
		return types.SigningAlgorithmSpecRsassaPkcs1V15Sha256, crypto.SHA256, nil
	case types.KeySpecEccNistP256:
		return types.SigningAlgorithmSpecEcdsaSha256, crypto.SHA256, nil
	case types.KeySpecEccNistP384:
		return types.SigningAlgorithmSpecEcdsaSha384, crypto.SHA384, nil
	case types.KeySpecEccNistP521:
		return types.SigningAlgorithmSpecEcdsaSha512, crypto.SHA512, nil
	default:
		return "", 0, fmt.Errorf("unsupported KMS Key Spec %q: only RSA and EC Signing Keys are supported", spec)
	}
}
