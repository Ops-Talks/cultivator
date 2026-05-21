// Package ci detects the active CI provider and exposes environment variables
// relevant to Cultivator, such as the pull-request base branch for Magic Mode
// (--changed-only).
package ci

import "os"

// Provider identifies a supported CI platform.
type Provider string

const (
	ProviderUnknown   Provider = ""
	ProviderBitbucket Provider = "bitbucket"
	ProviderGitHub    Provider = "github"
	ProviderGitLab    Provider = "gitlab"
)

// Environment holds CI-provider-agnostic values derived from the active CI
// environment. Fields may be empty when the corresponding value is not
// available (e.g. BaseRef is empty when not running in a pull-request
// pipeline).
type Environment struct {
	Provider  Provider
	IsPR      bool   // true when running inside a pull/merge request pipeline
	BaseRef   string // target branch of the PR / merge request
	HeadRef   string // source branch
	CommitSHA string
}

// Detect inspects the current process environment and returns the active CI
// environment. The first matching provider wins. If no supported CI provider
// is detected the returned Environment has Provider == ProviderUnknown.
func Detect() Environment {
	switch {
	case isBitbucket():
		return detectBitbucket()
	case isGitHub():
		return detectGitHub()
	case isGitLab():
		return detectGitLab()
	default:
		return Environment{}
	}
}

func isBitbucket() bool {
	return os.Getenv("BITBUCKET_BUILD_NUMBER") != ""
}

func isGitHub() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func isGitLab() bool {
	return os.Getenv("GITLAB_CI") == "true"
}

func detectBitbucket() Environment {
	return Environment{
		Provider:  ProviderBitbucket,
		IsPR:      os.Getenv("BITBUCKET_PR_ID") != "",
		BaseRef:   os.Getenv("BITBUCKET_PR_DESTINATION_BRANCH"),
		HeadRef:   os.Getenv("BITBUCKET_BRANCH"),
		CommitSHA: os.Getenv("BITBUCKET_COMMIT"),
	}
}

func detectGitHub() Environment {
	baseRef := os.Getenv("GITHUB_BASE_REF")
	return Environment{
		Provider:  ProviderGitHub,
		IsPR:      baseRef != "",
		BaseRef:   baseRef,
		HeadRef:   os.Getenv("GITHUB_HEAD_REF"),
		CommitSHA: os.Getenv("GITHUB_SHA"),
	}
}

func detectGitLab() Environment {
	return Environment{
		Provider:  ProviderGitLab,
		IsPR:      os.Getenv("CI_MERGE_REQUEST_IID") != "",
		BaseRef:   os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"),
		HeadRef:   os.Getenv("CI_COMMIT_REF_NAME"),
		CommitSHA: os.Getenv("CI_COMMIT_SHA"),
	}
}
