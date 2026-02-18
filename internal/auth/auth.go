package auth

import (
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// GetGitAuth returns authentication credentials for Git operations
// based on environment variables (ARCA_GIT_TOKEN, GITHUB_TOKEN, AZURE_DEVOPS_EXTTOKEN).
func GetGitAuth() *http.BasicAuth {
	token := os.Getenv("ARCA_GIT_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		token = os.Getenv("AZURE_DEVOPS_EXTTOKEN")
	}

	if token != "" {
		// For most providers (GitHub, GitLab, etc.), username can be "token"
		// or any non-empty string when using a personal access token as the password.
		return &http.BasicAuth{
			Username: "token",
			Password: token,
		}
	}
	return nil
}
