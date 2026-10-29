package tui

import "testing"

func TestFriendlyCredentialSourceMapsEnvironmentVariables(t *testing.T) {
	got := FriendlyCredentialSource("EnvConfigCredentials", "")
	want := "environment variables"
	if got != want {
		t.Fatalf("FriendlyCredentialSource(EnvConfigCredentials) = %q, want %q", got, want)
	}
}

func TestFriendlyCredentialSourceMapsSharedConfigFile(t *testing.T) {
	got := FriendlyCredentialSource("SharedConfigCredentials: /home/user/.aws/credentials", "default")
	want := "shared config profile default"
	if got != want {
		t.Fatalf("FriendlyCredentialSource(SharedConfigCredentials: ..., default) = %q, want %q", got, want)
	}
}

func TestFriendlyCredentialSourceMapsKnownProviders(t *testing.T) {
	cases := map[string]string{
		"SSOProvider":                 "AWS IAM Identity Center (SSO)",
		"EC2RoleProvider":             "EC2 instance role",
		"CredentialsEndpointProvider": "container credentials endpoint",
	}
	for source, want := range cases {
		if got := FriendlyCredentialSource(source, ""); got != want {
			t.Errorf("FriendlyCredentialSource(%q) = %q, want %q", source, got, want)
		}
	}
}

func TestFriendlyCredentialSourceEmptyMeansNoSourceToReport(t *testing.T) {
	if got := FriendlyCredentialSource("", ""); got != "" {
		t.Fatalf("FriendlyCredentialSource(\"\") = %q, want empty", got)
	}
}
