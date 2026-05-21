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
	return detectFromEnv(os.Getenv)
}

// detectFromEnv is the testable implementation of Detect. It accepts a getenv
// function so callers can substitute a fake environment without touching global
// os state.
func detectFromEnv(getenv func(string) string) Environment {
	switch {
	case getenv("BITBUCKET_BUILD_NUMBER") != "":
		return Environment{
			Provider:  ProviderBitbucket,
			IsPR:      getenv("BITBUCKET_PR_ID") != "",
			BaseRef:   getenv("BITBUCKET_PR_DESTINATION_BRANCH"),
			HeadRef:   getenv("BITBUCKET_BRANCH"),
			CommitSHA: getenv("BITBUCKET_COMMIT"),
		}
	case getenv("GITHUB_ACTIONS") == "true":
		baseRef := getenv("GITHUB_BASE_REF")
		return Environment{
			Provider:  ProviderGitHub,
			IsPR:      baseRef != "",
			BaseRef:   baseRef,
			HeadRef:   getenv("GITHUB_HEAD_REF"),
			CommitSHA: getenv("GITHUB_SHA"),
		}
	case getenv("GITLAB_CI") == "true":
		return Environment{
			Provider:  ProviderGitLab,
			IsPR:      getenv("CI_MERGE_REQUEST_IID") != "",
			BaseRef:   getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"),
			HeadRef:   getenv("CI_COMMIT_REF_NAME"),
			CommitSHA: getenv("CI_COMMIT_SHA"),
		}
	default:
		return Environment{}
	}
}
