package auth

import (
	"os"
	"testing"
)

func TestGetGitAuth(t *testing.T) {
	// Save existing environment to restore later
	originalArca := os.Getenv("ARCA_GIT_TOKEN")
	originalGitHub := os.Getenv("GITHUB_TOKEN")
	originalAzure := os.Getenv("AZURE_DEVOPS_EXTTOKEN")
	defer func() {
		os.Setenv("ARCA_GIT_TOKEN", originalArca)
		os.Setenv("GITHUB_TOKEN", originalGitHub)
		os.Setenv("AZURE_DEVOPS_EXTTOKEN", originalAzure)
	}()

	tests := []struct {
		name          string
		arcaToken     string
		githubToken   string
		azureToken    string
		expectedPass  string
		expectedFound bool
	}{
		{
			name:          "No tokens set",
			expectedFound: false,
		},
		{
			name:          "ARCA_GIT_TOKEN set",
			arcaToken:     "arca-secret",
			expectedPass:  "arca-secret",
			expectedFound: true,
		},
		{
			name:          "GITHUB_TOKEN set",
			githubToken:   "github-secret",
			expectedPass:  "github-secret",
			expectedFound: true,
		},
		{
			name:          "AZURE_DEVOPS_EXTTOKEN set",
			azureToken:    "azure-secret",
			expectedPass:  "azure-secret",
			expectedFound: true,
		},
		{
			name:          "Priority ARCA over GitHub",
			arcaToken:     "arca-secret",
			githubToken:   "github-secret",
			expectedPass:  "arca-secret",
			expectedFound: true,
		},
		{
			name:          "Priority GitHub over Azure",
			githubToken:   "github-secret",
			azureToken:    "azure-secret",
			expectedPass:  "github-secret",
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ARCA_GIT_TOKEN", tt.arcaToken)
			os.Setenv("GITHUB_TOKEN", tt.githubToken)
			os.Setenv("AZURE_DEVOPS_EXTTOKEN", tt.azureToken)

			auth := GetGitAuth()

			if !tt.expectedFound {
				if auth != nil {
					t.Fatalf("Expected nil auth, got %+v", auth)
				}
				return
			}

			if auth == nil {
				t.Fatalf("Expected non-nil auth")
			}

			if auth.Username != "token" {
				t.Errorf("Expected username 'token', got '%s'", auth.Username)
			}

			if auth.Password != tt.expectedPass {
				t.Errorf("Expected password '%s', got '%s'", tt.expectedPass, auth.Password)
			}
		})
	}
}
