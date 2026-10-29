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
	missing := missingConfigFields(ctx, cfg)
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
	}, missing)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func missingConfigFields(ctx context.Context, cfg aws.Config) []string {
	missing := make([]string, 0, 4)
	if cfg.Region == "" {
		missing = append(missing, "Region")
	}
	probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if _, err := cfg.Credentials.Retrieve(probeCtx); err != nil {
		missing = append(missing, "Access Key ID", "Secret Access Key")
	}
	return missing
}
