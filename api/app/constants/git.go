package constants

const (
	GitProviderGitHub    = "github"
	GitProviderGitLab    = "gitlab"
	GitProviderBitbucket = "bitbucket"

	GitConnectionStatusActive = "active"
	GitConnectionStatusError  = "error"
)

func IsValidGitProvider(provider string) bool {
	switch provider {
	case GitProviderGitHub, GitProviderGitLab, GitProviderBitbucket:
		return true
	default:
		return false
	}
}

func IsValidGitConnectionStatus(status string) bool {
	switch status {
	case GitConnectionStatusActive, GitConnectionStatusError:
		return true
	default:
		return false
	}
}
