package tui

import (
	"fmt"
	"strings"
)

// FriendlyCredentialSource maps the AWS SDK's raw aws.Credentials.Source
// value to a short, human-readable label for display in the TUI's status
// area, per the Credential source display decision. profile names the
// shared config/credentials file profile in effect (used only when source
// is a SharedConfigCredentials result); pass "" when unknown or
// inapplicable. Returns "" when source is empty (no credential source to
// report, e.g. the user typed everything by hand).
func FriendlyCredentialSource(source, profile string) string {
	switch {
	case source == "EnvConfigCredentials":
		return "environment variables"
	case strings.HasPrefix(source, "SharedConfigCredentials"):
		if profile != "" {
			return fmt.Sprintf("shared config profile %s", profile)
		}
		return "shared config/credentials file"
	case source == "SSOProvider":
		return "AWS IAM Identity Center (SSO)"
	case source == "EC2RoleProvider":
		return "EC2 instance role"
	case source == "CredentialsEndpointProvider":
		return "container credentials endpoint"
	case source == "AssumeRoleProvider":
		return "assumed IAM role"
	case source == "WebIdentityCredentials":
		return "web identity token"
	case source == "ProcessProvider":
		return "credential_process"
	default:
		return source
	}
}
