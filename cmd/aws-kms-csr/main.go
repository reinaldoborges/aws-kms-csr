package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	tea "github.com/charmbracelet/bubbletea"
	csr "github.com/reinaldo/aws-kms-csr"
	"github.com/reinaldo/aws-kms-csr/tui"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "loading AWS configuration:", err)
		os.Exit(1)
	}
	missing, credentialSource := resolveCredentialStatus(ctx, cfg)
	model := tui.NewWithCredentials(func(form tui.Form) ([]byte, error) {
		requestConfig := cfg
		if form.Region != "" {
			requestConfig.Region = form.Region
		}
		if form.AccessKeyID != "" && form.SecretAccessKey != "" {
			requestConfig.Credentials = aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
				form.AccessKeyID, form.SecretAccessKey, form.SessionToken,
			))
		}
		client := kms.NewFromConfig(requestConfig)
		return csr.CreateCSR(ctx, client, form.KeyIdentifier, form.CommonName)
	}, missing, credentialSource)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// resolveCredentialStatus reports which credential/region fields the
// default AWS SDK chain left missing (so the TUI can prompt for them) and,
// when the chain resolved credentials on its own, a human-readable label
// for where they came from (per the Credential source display decision).
func resolveCredentialStatus(ctx context.Context, cfg aws.Config) (missing []string, credentialSource string) {
	missing = make([]string, 0, 4)
	if cfg.Region == "" {
		missing = append(missing, "Region")
	}
	probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	creds, err := cfg.Credentials.Retrieve(probeCtx)
	if err != nil {
		missing = append(missing, "Access Key ID", "Secret Access Key")
		return missing, ""
	}
	return missing, tui.FriendlyCredentialSource(creds.Source, sharedConfigProfile())
}

// sharedConfigProfile returns the shared config/credentials file profile
// the default chain would select, matching the SDK's own AWS_PROFILE /
// "default" resolution. Used only to label a SharedConfigCredentials
// result; ignored for every other credential source.
func sharedConfigProfile() string {
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		return profile
	}
	return config.DefaultSharedConfigProfile
}
