package main

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func TestResolveCredentialStatusPromptsForSessionTokenWhenDefaultCredentialsAreUnavailable(t *testing.T) {
	cfg := aws.Config{
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("", "", "")),
	}

	missing, source := resolveCredentialStatus(context.Background(), cfg)

	if source != "" {
		t.Fatalf("credential source = %q, want empty", source)
	}
	want := []string{"Region", "Access Key ID", "Secret Access Key", "Session Token"}
	if len(missing) != len(want) {
		t.Fatalf("missing = %#v, want %#v", missing, want)
	}
	for i := range want {
		if missing[i] != want[i] {
			t.Fatalf("missing = %#v, want %#v", missing, want)
		}
	}
}
